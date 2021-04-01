package parser

import (
	"fmt"
	"inscript/token"
	"strconv"
	"time"
)

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
	if d < time.Minute {
		return 0, fmt.Errorf("every:= %s: time interval can't be shorter than a minute.", s)
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
		return fmt.Errorf("line %d: unexpected token %s, expected %s instead.", p.token.Line, p.token.Type, t)
	}
	return nil
}

func (p *Parser) read() error {
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
	err := p.read()
	if err != nil {
		return err
	}
	for p.token.Type != token.LF && p.token.Type != token.EOF {
		err = p.read()
		if err != nil {
			return err
		}
	}
	return nil
}