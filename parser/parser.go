package parser

import (
	"fmt"
	"github.com/insomnimus/inscript/ast"
	"github.com/insomnimus/inscript/lexer"
	"github.com/insomnimus/inscript/token"
	"strings"
)

type Parser struct {
	l           *lexer.Lexer
	token, peek token.Token
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
		return nil, fmt.Errorf("line %d: unexpected token of type %s.", p.token.Line, p.token.Type)
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
	if strings.HasPrefix(cmd.Command, ":") {
		cmd.Sync = true
		cmd.Command = strings.TrimPrefix(cmd.Command, ":")
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
		return nil, fmt.Errorf("line %d: expected left brace, got %s instead.", p.token.Line, p.token.Type)
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
			return nil, fmt.Errorf("line %d: unexpected token of type %s in command block.", p.token.Line, p.token.Type)
		}
		err = p.read()
		if err != nil {
			return nil, err
		}
	}
	syncSet := false

	for _, f := range fields {
		switch strings.ToLower(f.key) {
		case "name":
			cmd.Name = f.val
		case "stdin":
			cmd.Stdin = f.val
		case "stdout":
			cmd.Stdout = f.val
		case "stderr":
			cmd.Stderr = f.val
		case "times":
			cmd.Times, err = parseTimes(f.val)
			if err != nil {
				return nil, err
			}
		case "sync":
			syncSet = true
			switch strings.ToLower(f.val) {
			case "yes", "true":
				cmd.Sync = true
			case "false", "no", "":
			default:
				return nil, fmt.Errorf("invalid boolean value for sync field %q.", f.val)
			}
		case "every":
			cmd.Every, err = parseInterval(f.val)
			if err != nil {
				return nil, err
			}
		case "dir", "workingdirectory":
			cmd.Dir = f.val
		default:
			return nil, fmt.Errorf("unknown field %q in command block.", f.key)
		}
	}

	if !syncSet && strings.HasPrefix(cmd.Command, ":") {
		cmd.Sync = true
		cmd.Command = strings.TrimPrefix(cmd.Command, ":")
	}
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
