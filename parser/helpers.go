package parser

import (
	"fmt"
	"github.com/insomnimus/inscript/ast"
	"github.com/insomnimus/inscript/token"
	"strconv"
	"time"
)

var zeroToken = token.Token{}

type EOFError struct{}

func (*EOFError) Error() string { return "end of file" }

var ErrEOF = EOFError{}

func pnc(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func parseInterval(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 30*time.Second {
		return 0, fmt.Errorf("every:= %s: time interval can't be shorter than 30 seconds", s)
	}
	return d, nil
}

func parseTimes(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, nil
	}
	return n, nil
}

type field struct {
	key string
	val string
}

func (p *Parser) expect(t token.TokenType) error {
	err := p.read()
	if err != nil {
		return err
	}
	if p.token.Type != t {
		return fmt.Errorf("line %d: unexpected token %s, expected %s instead", p.token.Line, p.token.Type, t)
	}
	return nil
}

func (p *Parser) read() error {
	p.prev = p.token
	p.token = p.peek
	var err error
	p.peek, err = p.l.Next()
	return err
}

func (p *Parser) skipLF() error {
	var err error
	for p.token.Type == token.LF {
		err = p.read()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) skipComment() error {
	// sanity check
	if p.token.Type != token.Comment {
		pnc("internal error: line %d: p.skipComment called on token of type %s, expected %s instead.", p.token.Line, p.token.Type, token.Comment)
	}
	startOfLine := false
	if p.prev.Type == token.LF || p.prev == zeroToken {
		startOfLine = true
	}
	err := p.read()
	if err != nil {
		return err
	}
	// check for directives
	var tokens []token.Token
	for p.token.Type != token.LF && p.token.Type != token.EOF {
		tokens = append(tokens, p.token)
		err = p.read()
		if err != nil {
			return err
		}
	}
	if startOfLine {
		return p.checkForDirective(tokens...)
	}
	return nil
}

func (p *Parser) applyDirectives(cmd *ast.Command, set map[string]struct{}) {
	if cmd == nil {
		return
	}
	if set == nil {
		set = make(map[string]struct{})
	}
	if _, ok := set["dir"]; !ok && p.dir != "" {
		cmd.Dir = p.dir
	}
	if _, ok := set["sync"]; !ok && p.sync != "" {
		switch p.sync {
		case "true", "yes":
			cmd.Sync = true
		default:
			cmd.Sync = false
		}
	}
	if _, ok := set["stdin"]; !ok && p.stdin != "" {
		cmd.Stdin = p.stdin
	}
	if _, ok := set["stdout"]; !ok && p.stdout != "" {
		cmd.Stdout = p.stdout
	}
	if _, ok := set["stderr"]; !ok && p.stderr != "" {
		cmd.Stderr = p.stderr
	}
}
