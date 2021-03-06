package lexer

import (
	"github.com/insomnimus/inscript/token"
	"os"
	"testing"
)

func TestReadString(t *testing.T) {
	os.Setenv("test_var", "testvar")
	items := []struct {
		in, out string
	}{
		{`"hello"`, "hello"},
		{`"\x25ab"`, "%ab"},
		{`"x y \"z\""`, `x y "z"`},
		{`"\n\t\r"`, "\n\t\r"},
		{`"\0101BCD"`, "ABCD"},
		{`"\uffffB"`, "\uffffB"},
		{`"$test_var\$test_var"`, "testvar$test_var"},
	}
	for _, s := range items {
		l := New(s.in)
		out, err := l.Next()
		if err != nil {
			t.Errorf("error parsing (%s): %s", s.in, err)
		}
		if out.Type != token.String {
			t.Errorf("expected type (%s) to be string, got %s instead.", s.in, out.Type)
		}
		if out.Literal != s.out {
			t.Errorf("string parsed incorrectly:\ninput: (%s)\nexpected (%s), got (%s)", s.in, s.out, out.Literal)
		}
	}
}

func testNext(t *testing.T) {
	tk := func(tp token.TokenType, lit string, ln int) token.Token {
		return token.Token{
			Type:    tp,
			Literal: lit,
			Line:    ln,
		}
	}
	input := `#!/bin/github.com/insomnimus/inscript
go run main.go
@ echo "haha test string" {
	stdout := outerino
	stderr := errino
	sync := true
}

# haha comment
multi\
line`
	l := New(input)
	tests := []token.Token{
		tk(token.Comment, "!/bin/github.com/insomnimus/inscript", 1),
		tk(token.LF, "\n", 2),
		tk(token.String, "go", 2),
		tk(token.String, "run", 2),
		tk(token.String, "main.go", 2),
		tk(token.LF, "\n", 3),
		tk(token.At, "@", 3),
		tk(token.String, "echo", 3),
		tk(token.String, "haha test string", 3),
		tk(token.LBrace, "{", 3),
		tk(token.LF, "\n", 4),
		tk(token.String, "stdout", 4),
		tk(token.Assign, ":=", 4),
		tk(token.String, "outerino", 4),
		tk(token.LF, "\n", 5),
		tk(token.String, "stderr", 5),
		tk(token.Assign, ":=", 5),
		tk(token.String, "errino", 5),
		tk(token.LF, "\n", 6),
		tk(token.String, "sync", 6),
		tk(token.Assign, ":=", 6),
		tk(token.String, "true", 6),
		tk(token.LF, "\n", 7),
		tk(token.RBrace, "}", 7),
		tk(token.LF, "\n", 8),
		tk(token.Comment, "haha comment", 8),
		tk(token.LF, "\n", 9),
		tk(token.String, "multiline", 9),
		tk(token.EOF, "", 9),
	}
	for _, test := range tests {
		tok, err := l.Next()
		if err != nil {
			t.Errorf("l.Next returned error: %s\n", err)
		}
		if tok.Type != test.Type {
			t.Errorf("type mismatch:\nexpected %#v\ngot %#v\n", test, tok)
		}
		if tok.Literal != test.Literal {
			t.Errorf("literal mismatch:\nexpected %#v\ngot %#v\n", test, tok)
		}
		if tok.Line != test.Line {
			t.Errorf("line mismatch:\nexpected %#v\ngot %#v\n", test, tok)
		}
	}
}
