package main

import (
	"fmt"
	"github.com/insomnimus/inscript/ast"
	"github.com/insomnimus/inscript/lexer"
	"github.com/insomnimus/inscript/parser"
	"github.com/insomnimus/inscript/runtime"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
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
	log.SetFlags(0)
	log.SetPrefix("")

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	for i, a := range os.Args[1:] {
		os.Setenv(fmt.Sprint(i), a)
	}
	os.Setenv("#", fmt.Sprint(len(os.Args)-2))
	os.Setenv("@", strings.Join(os.Args[2:], " "))
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
	done := make(chan struct{}, len(commands))

	for _, cmd := range commands {
		pr, err := runtime.CreateProcess(cmd)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
		if pr.Async {
			//shouldWait = true
			go func() {
				err := pr.Run()
				if err != nil {
					log.Fatal(err)
				}
				done <- struct{}{}
			}()
			continue
		}
		err = pr.Run()
		if err != nil {
			log.Fatal(err)
		}
		done <- struct{}{}

	}

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt)
	for i := 0; i < len(commands); i++ {
		select {
		case <-sig:
			return
		case <-done:
		}
	}
}
