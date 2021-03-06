package parser

import (
	"github.com/insomnimus/inscript/ast"
	"github.com/insomnimus/inscript/lexer"
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	input := `#!/bin/github.com/insomnimus/inscript
#this is a test program
x:=42
:go run main.go

@ cat go.mod {
	sync:= true
	stdout:= cat.out
	every:= 1h
}

bash -c 'echo hello'

!:+echo $x

# test directive
#<sync=true>

ls
sleep
mkdir
#<sync=>
sed
`
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
		}, {
			Command: "echo",
			Args:    []string{"42"},
			Stdout:  "!stdout",
			Stdin:   "!stdin",
			Stderr:  "!stderr",
			Sync:    true,
		}, {
			Command: "ls",
			Sync:    true,
		}, {
			Command: "sleep",
			Sync:    true,
		}, {
			Command: "mkdir",
			Sync:    true,
		}, {
			Command: "sed",
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
