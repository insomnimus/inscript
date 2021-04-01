package main

import (
	"fmt"
	"inscript/lexer"
	"inscript/parser"
)

func main() {
	input := `#!/bin/inscript
#this is a test program
:go run main.go

@ cat go.mod {
	sync:= true
	stdout:= cat.out
	every:= 1h
}

bash -c 'echo hello'`
	l := lexer.New(input)
	p, err := parser.New(l)
	if err != nil {
		fmt.Println(err)
		return
	}
	for cmd, err := p.Next(); err != &parser.ErrEOF; cmd, err = p.Next() {
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%#v", cmd)
	}
}
