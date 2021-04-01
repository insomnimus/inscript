package parser

import (
	"inscript/ast"
	"inscript/lexer"
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	input := `#!/bin/inscript
#this is a test program
:go run main.go

@ cat go.mod {
	sync:= true
	stdout:= cat.out
	every:= 1h
}

bash -c 'echo hello'`
	tests := []*ast.Command{
		{
			Command: "go",
			Args:    []string{"run", "main.go"},
			Sync:    true,
		}, {
			Command: "cat",
			Args:    []string{"go.mod"},
			Sync:    true,
			Stdout:  "cat.out",
			Every:   time.Hour,
		}, {
			Command: "bash",
			Args:    []string{"-c", "echo hello"},
		},
	}
	l := lexer.New(input)
	p, err := New(l)
	if err != nil {
		t.Errorf("parser.New: returned error: %s\n", err)
	}
	for _, test := range tests {
		cmd, err := p.Next()
		if err != nil {
			t.Errorf("p.Next returned error: %s\n\nwas expecting: %#v\n", err, test)
		}
		if !test.Equal(*cmd) {
			t.Errorf("command mismatch:\nexpected %#v\ngot %#v\n", test, cmd)
		}
	}
}
