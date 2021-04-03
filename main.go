package main

import (
	"inscript/ast"
	"inscript/lexer"
	"inscript/parser"
	"inscript/runtime"
	"io"
	"log"
	"os"
	"os/signal"
)

func showAbout() {
	log.Println("inscript interpreter\nuse inscript --help for the usage")
	os.Exit(0)
}

func showHelp() {
	log.Println("help coming soon")
	os.Exit(0)
}

func main() {
	var fileName string
	log.SetFlags(0)
	log.SetPrefix("")
	var reader io.Reader
	if fi, e := os.Stdin.Stat(); e == nil {
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			reader = os.Stdin
		}
	}
	if len(os.Args) == 1 && reader == nil {
		showAbout()
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
	}
	if reader == nil {
		fileName = os.Args[1]
		f, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		reader = f
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	l := lexer.New(string(data))
	p, err := parser.New(l)
	if err != nil {
		log.Fatal(err)
	}
	var commands []*ast.Command
	for cmd, err := p.Next(); err != &parser.ErrEOF; cmd, err = p.Next() {
		if err != nil {
			log.Fatal(err)
		}
		commands = append(commands, cmd)
	}

	shouldWait := false
	var processes []*runtime.Process
	for _, cmd := range commands {
		procs, err := runtime.CreateProcess(cmd)
		if err != nil {
			log.Fatal(err)
		}
		if procs.ShouldWait {
			shouldWait = true
		}
		processes = append(processes, procs)
	}
	if fileName != "" {
		log.Printf("starting %s\n", fileName)
	}
	if !shouldWait {
		for _, pr := range processes {
			err = pr.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
		if fileName != "" {
			log.Printf("done %s\n", fileName)
		}
		return
	}

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt)
	for _, pr := range processes {
		go func() {
			err := pr.Run()
			if err != nil {
				log.Println(fileName, err)
			}
		}()
		defer pr.Kill()
	}
	<-sig
	if fileName != "" {
		log.Printf("done %s\n", fileName)
	}
}
