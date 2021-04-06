package parser

import (
	"fmt"
	"github.com/insomnimus/inscript/ast"
	"github.com/insomnimus/inscript/lexer"
	"github.com/insomnimus/inscript/token"
	"os"
	"strings"
)

type Parser struct {
	l                 *lexer.Lexer
	prev, token, peek token.Token

	// parsed directives
	stdin, stdout, stderr, dir, sync string
}

func New(l *lexer.Lexer) (*Parser, error) {
	p := &Parser{
		l: l,
	}
	err := p.read()
	if err != nil {
		return nil, err
	}
	err = p.read()
	return p, err
}

func (p *Parser) Next() (*ast.Command, error) {
	err := p.skipLF()
	if err != nil {
		return nil, err
	}
	var cmd *ast.Command
	switch p.token.Type {
	case token.EOF:
		return nil, &ErrEOF
	case token.String:
		if p.peek.Type == token.Assign {
			err = p.parseVariable()
			if err != nil {
				return nil, err
			}
			return p.Next()
		}
		cmd, err = p.parseInlineCommand()
	case token.At:
		cmd, err = p.parseCommand()
	case token.Comment:
		err = p.skipComment()
		if err != nil {
			return nil, err
		}
		return p.Next()
	default:
		return nil, fmt.Errorf("line %d: unexpected token of type %s", p.token.Line, p.token.Type)
	}
	if err != nil {
		return nil, err
	}
	if cmd == nil {
		fmt.Printf("cmd is nil, current token: %#v\n", p.token)
	}
	err = p.read()
	return cmd, err
}

// reads only related tokens
func (p *Parser) parseInlineCommand() (*ast.Command, error) {
	// sanity check
	if p.token.Type != token.String {
		pnc("internal error: p.parseInlineCommand called with token type %s instead of %s.", p.token.Type, token.String)
	}

	cmd := &ast.Command{
		Command: p.token.Literal,
	}
	setFields := make(map[string]struct{})
FOR:
	for i, c := range cmd.Command {
		switch c {
		case ':':
			setFields["sync"] = struct{}{}
			cmd.Sync = true
		case '!':
			setFields["stderr"] = struct{}{}
			setFields["stdout"] = struct{}{}
			cmd.Stderr = "!stderr"
			cmd.Stdout = "!stdout"
		case '+':
			setFields["stdin"] = struct{}{}
			cmd.Stdin = "!stdin"
		default:
			cmd.Command = cmd.Command[i:]
			break FOR
		}
	}
	err := p.read()
	if err != nil {
		return nil, err
	}
	for p.token.Type == token.String {
		cmd.Args = append(cmd.Args, p.token.Literal)
		err = p.read()
		if err != nil {
			return nil, err
		}
	}
	// apply directives, if any
	p.applyDirectives(cmd, setFields)
	return cmd, err
}

// reads only related tokens
func (p *Parser) parseCommand() (*ast.Command, error) {
	// sanity check
	if p.token.Type != token.At {
		pnc("internal error: p.parseCommand called on token %s, expected %s instead.", p.token.Type, token.At)
	}
	err := p.expect(token.String)
	if err != nil {
		return nil, err
	}
	cmd := &ast.Command{
		Command: p.token.Literal,
	}
	err = p.read()
	if err != nil {
		return nil, err
	}
	for p.token.Type == token.String {
		cmd.Args = append(cmd.Args, p.token.Literal)
		err = p.read()
		if err != nil {
			return nil, err
		}
	}
	if p.token.Type != token.LBrace {
		return nil, fmt.Errorf("line %d: expected left brace, got %s instead", p.token.Line, p.token.Type)
	}
	err = p.read()
	if err != nil {
		return nil, err
	}
	// skip over line feeds
	for p.token.Type == token.LF {
		err = p.read()
		if err != nil {
			return nil, err
		}
	}
	var fields []field
	var f field
	// read fields if any
LOOP:
	for {
		switch p.token.Type {
		case token.Comment:
			err = p.skipComment()
			if err != nil {
				return nil, err
			}
			continue LOOP
		case token.RBrace:
			break LOOP
		case token.String:
			f, err = p.parseField()
			if err != nil {
				return nil, err
			}
			fields = append(fields, f)
			continue LOOP
		case token.LF:
		case token.EOF:
			return nil, fmt.Errorf("unexpected end of file in command block")
		default:
			return nil, fmt.Errorf("line %d: unexpected token of type %s in command block", p.token.Line, p.token.Type)
		}
		err = p.read()
		if err != nil {
			return nil, err
		}
	}
	setFields := make(map[string]struct{})

	for _, f := range fields {
		switch strings.ToLower(f.key) {
		case "name":
			setFields["name"] = struct{}{}
			cmd.Name = f.val
		case "stdin":
			setFields["stdin"] = struct{}{}
			cmd.Stdin = f.val
		case "stdout":
			setFields["stdout"] = struct{}{}
			cmd.Stdout = f.val
		case "stderr":
			setFields["stderr"] = struct{}{}
			cmd.Stderr = f.val
		case "times":
			setFields["times"] = struct{}{}
			cmd.Times, err = parseTimes(f.val)
			if err != nil {
				return nil, err
			}
		case "sync":
			setFields["sync"] = struct{}{}
			switch strings.ToLower(f.val) {
			case "yes", "true":
				cmd.Sync = true
			case "false", "no", "":
			default:
				return nil, fmt.Errorf("invalid boolean value for sync field %q", f.val)
			}
		case "every":
			setFields["every"] = struct{}{}
			cmd.Every, err = parseInterval(f.val)
			if err != nil {
				return nil, err
			}
		case "dir", "workingdirectory":
			setFields["dir"] = struct{}{}
			cmd.Dir = f.val
		default:
			return nil, fmt.Errorf("unknown field %q in command block", f.key)
		}
	}
FOR:
	for i, c := range cmd.Command {
		switch c {
		case ':':
			if _, ok := setFields["sync"]; !ok {
				setFields["sync"] = struct{}{}
				cmd.Sync = true
			}
		case '!':
			if _, ok := setFields["stderr"]; !ok {
				setFields["stderr"] = struct{}{}
				cmd.Stderr = "!stderr"
			}
			if _, ok := setFields["stdout"]; !ok {
				setFields["stdout"] = struct{}{}
				cmd.Stdout = "!stdout"
			}
		case '+':
			if _, ok := setFields["stdin"]; !ok {
				setFields["stdin"] = struct{}{}
				cmd.Stdin = "!stdin"
			}
		default:
			cmd.Command = cmd.Command[i:]
			break FOR
		}
	}
	p.applyDirectives(cmd, setFields)

	return cmd, err
}

// reads only related tokens
func (p *Parser) parseField() (f field, err error) {
	// sanity check
	if p.token.Type != token.String {
		pnc("internal error: p.parseField called with token type %s, expected %s instead.", p.token.Type, token.String)
	}
	f.key = p.token.Literal
	err = p.expect(token.Assign)
	if err != nil {
		return
	}
	err = p.read()
	if err != nil {
		return
	}
	// read and concat the rest
	var fields []string
	for p.token.Type == token.String {
		fields = append(fields, p.token.Literal)
		err = p.read()
		if err != nil {
			return
		}
	}
	f.val = strings.Join(fields, " ")
	return
}

func (p *Parser) checkForDirective(t token.Token) (err error) {
	if !strings.HasPrefix(t.Literal, "<") ||
		!strings.HasSuffix(t.Literal, ">") ||
		!strings.Contains(t.Literal, "=") {
		return
	}
	comment := strings.TrimPrefix(t.Literal, "<")
	comment = strings.TrimSuffix(comment, ">")
	split := strings.SplitN(comment, "=", 2)
	key := split[0]
	val := split[1]
	switch strings.ToLower(key) {
	case "dir", "workingdirectory":
		p.dir = val
	case "sync":
		switch strings.ToLower(val) {
		case "true", "yes":
			p.sync = "true"
		case "":
			p.sync = ""
		case "false", "no":
			p.sync = "false"
		default:
			return fmt.Errorf("line %d: invalid value %q for 'sync' directive, values must be true or false", t.Line, val)
		}
	case "stdin":
		p.stdin = val
	case "stdout":
		p.stdout = val
	case "stderr":
		p.stderr = val
	default:
		return fmt.Errorf("line %d: unrecognized directive: %s", t.Line, t.Literal)
	}
	return
}

// only reads related tokens
func (p *Parser) parseVariable() error {
	// sanity check
	if p.token.Type != token.String {
		pnc("internal error: line %d: p.parseVariable called on a %s token, expected %s instead", p.token.Line, p.token.Type, token.String)
	}
	if p.peek.Type != token.Assign {
		pnc("internal error: line %d: p.parseVariable called on peek %s token, expected %s instead", p.peek.Line, p.peek.Type, token.Assign)
	}
	key := p.token.Literal
	err := p.read()
	if err != nil {
		return err
	}
	var val string
	switch p.peek.Type {
	case token.EOF, token.LF:
	case token.String:
		err = p.read()
		if err != nil {
			return err
		}
		val = p.token.Literal
		err = p.read()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("line %d: can't assign %s to a variable, the value has to be a string", p.token.Line, p.token.Literal)
	}
	return os.Setenv(key, val)
}
