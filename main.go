package main

import (
	"github.com/insomnimus/inscript/ast"
	"github.com/insomnimus/inscript/lexer"
	"github.com/insomnimus/inscript/parser"
	"github.com/insomnimus/inscript/runtime"
	"io"
	"log"
	"os"
	"os/signal"
	"time"
)

func showAbout() {
	log.Println("github.com/insomnimus/github.com/insomnimus/inscript interpreter\nuse inscript --help for the usage")
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

	//shouldWait := false

	if fileName != "" {
		log.Printf("starting %s\n", fileName)
	}
	/*
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
	*/
	done := make(chan struct{}, len(commands))
	for _, cmd := range commands {
		pr, err := runtime.CreateProcess(cmd)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
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
		time.Sleep(50 * time.Millisecond)
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
	if fileName != "" {
		log.Printf("done %s\n", fileName)
	}
}
