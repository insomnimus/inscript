package lexer

import (
	"encoding/binary"
	"fmt"
	"inscript/token"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	text         []rune
	ch           rune
	pos, readpos int
	line         int
}

func New(s string) *Lexer {
	s = strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(s)
	l := &Lexer{
		text: []rune(s),
	}
	l.read()
	return l
}

func (l *Lexer) Next() (token.Token, error) {
	ln := l.line
	var t token.Token
	l.skipSpace()
	switch l.ch {
	case '\n':
		t = l.newToken(token.LF, "\n")
	case 0:
		t = l.newToken(token.EOF, "")
	case ':':
		if l.peek() == '=' {
			l.read()
			t = l.newToken(token.Assign, ":=")
		} else {
			return l.newToken(
				token.String,
				l.readStringBare(),
				ln), nil
		}
	case '"':
		s, err := l.readString()
		if err != nil {
			return t, err
		}
		t = l.newToken(token.String, s, ln)
	case '`', '\'':
		s, err := l.readStringLiteral(l.ch)
		if err != nil {
			return t, err
		}
		t = l.newToken(token.String, s, ln)
	case '#':
		return l.readComment(), nil
	case '@':
		t = l.newToken(token.At, "@")
	case '{':
		t = l.newToken(token.LBrace, "{")
	case '}':
		t = l.newToken(token.RBrace, "}")
	default:
		s := l.readStringBare()
		return l.newToken(token.String, s, ln), nil
	}
	l.read()
	return t, nil
}

func (l *Lexer) readString() (string, error) {
	// sanity check
	if l.ch != '"' {
		panic(fmt.Sprintf("line %d: l.readString called on character %q", l.line, l.ch))
	}
	startLn := l.line
	l.read()
	var buff strings.Builder
LOOP:
	for {
		switch l.ch {
		case 0:
			return "", l.err(startLn, "quoted string not terminated with '\"'")
		case '"':
			break LOOP
		case '\\':
			l.read()
			switch l.ch {
			case 'f':
				buff.WriteRune('\f')
			case 't':
				buff.WriteRune('\t')
			case 'v':
				buff.WriteRune('\v')
			case 'a':
				buff.WriteRune('\a')
			case 'b':
				buff.WriteRune('\b')
			case 'n':
				buff.WriteRune('\n')
			case 'r':
				buff.WriteRune('\r')
			case '0':
				err := l.readOctal(&buff)
				if err != nil {
					return "", err
				}
				continue LOOP
			case 'x':
				err := l.readHex(&buff)
				if err != nil {
					return "", err
				}
				continue LOOP
			case 'u':
				err := l.readUnicode(&buff)
				if err != nil {
					return "", err
				}
				continue LOOP
			case '$':
				buff.WriteRune(unicode.ReplacementChar)
			case '"', '\\':
				buff.WriteRune(l.ch)
			case 0:
				return "", l.err(l.line, "quoted string not terminated with '\"'")
			default:
				return "", l.err(l.line, "invalid escape sequence '\\%c'", l.ch)
			}
		default:
			buff.WriteRune(l.ch)
		}
		l.read()
	}
	return strings.ReplaceAll(
		os.ExpandEnv(buff.String()), string(unicode.ReplacementChar), "$"), nil
}

func (l *Lexer) readStringLiteral(ch rune) (string, error) {
	// sanity check
	if l.ch != ch {
		panic(fmt.Sprintf("line %d: l.readStringLiteral called on %q, expecting %q", l.line, l.ch, ch))
	}
	startLn := l.line
	l.read()
	var buff strings.Builder

LOOP:
	for {
		switch l.ch {
		case 0:
			return "", l.err(startLn, "quoted string literal not terminated with \"%c\"", ch)
		case '\n':
			if ch == '\'' {
				return "", l.err(startLn, "new line now allowed in single quoted string literals")
			}
			buff.WriteRune(l.ch)
		case '\\':
			if l.peek() == ch {
				l.read()
				buff.WriteRune(l.ch)
			} else {
				buff.WriteRune(l.ch)
			}
		case ch:
			break LOOP
		default:
			buff.WriteRune(l.ch)
		}
		l.read()
	}
	return buff.String(), nil
}

func (l *Lexer) readHex(buff *strings.Builder) error {
	// sanity check
	if l.ch != 'x' {
		panic(fmt.Sprintf("line %d: l.readHex called on %q, expected 'x' instead", l.line, l.ch))
	}
	startLn := l.line
	l.read()
	if l.ch == 0 {
		return l.err(l.line, "unexpected EoF in hex escape")
	}
	if !isHex(l.ch) {
		return l.err(startLn, "invalid hex escape sequence '\\x%c'", l.ch)
	}
	rs := make([]rune, 2)
	rs[0] = l.ch
	l.read()
	if !isHex(l.ch) {
		return l.err(startLn, "invalid hex escape sequence '\\x%c%c'", rs[0], l.ch)
	}
	rs[1] = l.ch
	l.read()
	n, err := strconv.ParseUint(string(rs), 16, 8)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buff, binary.LittleEndian, uint8(n))
	if err != nil {
		panic(err)
	}
	return nil
}

func (l *Lexer) readOctal(buff *strings.Builder) error {
	// sanity check
	if l.ch != '0' {
		panic(fmt.Sprintf("l.readOctal called with char %q; expected char '0'", l.ch))
	}
	startLn := l.line
	l.read()
	if !isOctal(l.ch) {
		return l.err(startLn, "invalid octal escape sequence: '\\%c'", l.ch)
	}
	rs := make([]rune, 3)
	rs[0] = l.ch
	l.read()
	if !isOctal(l.ch) {
		return l.err(startLn, "invalid octal escape sequence '\\0%c%c'", rs[0], l.ch)
	}
	rs[1] = l.ch
	l.read()
	if !isOctal(l.ch) {
		return l.err(l.line, "invalid octal escape sequence '\\0%c%c%c'", rs[0], rs[1], l.ch)
	}
	rs[2] = l.ch
	l.read()

	n, err := strconv.ParseUint(string(rs), 8, 16)
	if err != nil {
		panic(err)
	}
	//bs:= make([]byte, 16)
	//binary.LittleEndian.PutUint16(bs, uint16(n))
	//buff.Write(bs[:9])

	err = binary.Write(buff, binary.LittleEndian, uint8(n))
	if err != nil {
		panic(err)
	}
	return nil
}

func (l *Lexer) readUnicode(buff *strings.Builder) error {
	// sanity check
	if l.ch != 'u' {
		panic(fmt.Sprintf("l.readUnicode called with char %q; expected 'u' instead", l.ch))
	}
	startLn := l.line
	l.read()
	rs := make([]rune, 4)
	for i := 0; i < 4; i++ {
		if !isHex(l.ch) {
			return l.err(startLn, "invalid unicode short escape sequence '\\u%s%c'", string(rs[:i]), l.ch)
		}
		rs[i] = l.ch
		l.read()
	}
	s, err := strconv.Unquote(fmt.Sprintf(`"\u%s"`, string(rs)))
	if err != nil {
		panic(err)
	}
	buff.WriteString(s)
	return nil
	/*
		n, err := strconv.ParseUint(string(rs), 16, 16)
		if err != nil {
			panic(err)
		}
		err = binary.Write(buff, binary.LittleEndian, uint8(n))
		if err != nil {
			panic(err)
		}
	*/
}

func (l *Lexer) readComment() token.Token {
	// sanity check
	if l.ch != '#' {
		panic(fmt.Sprintf("line %d: l.readComment called with char %q; expected '#'", l.line, l.ch))
	}
	ln := l.line
	var buff strings.Builder
	l.read()
	for l.ch != '\n' && l.ch != 0 {
		buff.WriteRune(l.ch)
		l.read()
	}
	return l.newToken(token.Comment,
		strings.TrimSpace(buff.String()),
		ln)
}

func (l *Lexer) readStringBare() string {
	// sanity check
	if unicode.IsSpace(l.ch) {
		panic(fmt.Sprintf("line %d: l.readStringBare called on a space char (%d)", l.line, l.ch))
	}
	var buff strings.Builder
LOOP:
	for {
		switch l.ch {
		case '\\':
			switch l.peek() {
			case 0:
				break LOOP
			case '{', '}':
				l.read()
				buff.WriteRune(l.ch)
			default:
				if unicode.IsSpace(l.peek()) {
					l.read()
				} else {
					buff.WriteRune(l.ch)
				}
			}
		case 0, '\n', '{', '}':
			break LOOP
		case ':':
			if l.peek() == '=' {
				break LOOP
			}
			buff.WriteRune(l.ch)
		default:
			if unicode.IsSpace(l.ch) {
				break LOOP
			}
			buff.WriteRune(l.ch)
		}
		l.read()
	}
	return os.ExpandEnv(buff.String())
}
