package token

import "fmt"

type TokenType uint8

//go:generate stringer -type=TokenType

const (
	_ TokenType = iota
	İllegal
	EOF
	LF
	At
	LBrace
	RBrace
	Comment
	String
	Assign
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

func (t Token) GoString() string {
	return fmt.Sprintf(`Token{
		Type: %s,
		Line: %d,
		Literal: %q,
	}`, t.Type, t.Line, t.Literal)
}
