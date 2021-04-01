package lexer

import (
	"fmt"
	"inscript/token"
	"unicode"
)

const (
	octalChars = "01234567"
	hexChars   = "0123456789abcdefABCDEF"
)

func (l *Lexer) read() {
	if l.readpos >= len(l.text) {
		l.ch = 0
	} else {
		l.ch = l.text[l.readpos]
	}
	if l.ch == '\n' {
		l.line++
	}
	l.pos = l.readpos
	l.readpos++
}

func (l *Lexer) peek() rune {
	if l.readpos >= len(l.text) {
		return 0
	}
	return l.text[l.readpos]
}

func (l *Lexer) newToken(t token.TokenType, s string, ln ...int) token.Token {
	line := l.line
	if len(ln) > 0 {
		line = ln[0]
	}
	return token.Token{
		Type:    t,
		Literal: s,
		Line:    line,
	}
}

func isHex(c rune) bool {
	for _, ch := range hexChars {
		if c == ch {
			return true
		}
	}
	return false
}

func isOctal(c rune) bool {
	for _, ch := range octalChars {
		if c == ch {
			return true
		}
	}
	return false
}

func (l *Lexer) err(ln int, format string, args ...interface{}) error {
	if len(args) == 0 {
		return fmt.Errorf("line %d: %s", ln, format)
	}
	args = append([]interface{}{ln}, args...)
	return fmt.Errorf("line %d: "+format, args...)
}

func (l *Lexer) skipSpace() {
	for unicode.IsSpace(l.ch) && l.ch != '\n' && l.ch != 0 {
		l.read()
	}
}
