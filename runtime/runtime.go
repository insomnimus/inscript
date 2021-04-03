package runtime

import (
	"fmt"
	"inscript/ast"
	"log"
	"os"
	"os/exec"
	"time"
)

type Process struct {
	Cmd        *exec.Cmd
	Command    *ast.Command
	ShouldWait bool
	Kill       func()
	Run        func() error
	killed     bool
}

func CreateProcess(cmd *ast.Command) (*Process, error) {
	command := exec.Command(cmd.Command, cmd.Args...)
	if cmd.Dir != "" {
		command.Dir = cmd.Dir
	}
	var stdin, stdout, stderr *os.File
	var err error
	if cmd.Stderr != "" {
		switch {
		case cmd.Stderr == "!stderr":
			command.Stderr = os.Stderr
		case cmd.Stderr == "!stdout":
			command.Stderr = os.Stdout
		default:
			if _, e := os.Stat(cmd.Stderr); os.IsNotExist(e) {
				stderr, err = os.Create(cmd.Stderr)
			} else {
				stderr, err = os.OpenFile(cmd.Stderr, os.O_WRONLY, 0764)
			}
			if err != nil {
				return nil, err
			}
			command.Stderr = stderr
		}
	}
	if cmd.Stdout != "" {
		switch {
		case cmd.Stdout == "!stdout":
			command.Stdout = os.Stdout
		case cmd.Stdout == "!stderr":
			command.Stdout = os.Stderr
		case cmd.Stdout == cmd.Stderr && stderr != nil:
			command.Stdout = stderr
		default:
			if _, e := os.Stat(cmd.Stdout); os.IsNotExist(e) {
				stdout, err = os.Create(cmd.Stdout)
			} else {
				stdout, err = os.OpenFile(cmd.Stdout, os.O_WRONLY, 0764)
			}
			if err != nil {
				return nil, err
			}
			command.Stdout = stdout
		}
	}
	if cmd.Stdin != "" {
		switch {
		case cmd.Stdin == "!stdin":
			command.Stdin = os.Stdin
		case cmd.Stdin != cmd.Stdout && cmd.Stdin != cmd.Stderr:
			if _, e := os.Stat(cmd.Stdin); !os.IsNotExist(e) {
				stdin, err = os.Open(cmd.Stdin)
				if err != nil {
					return nil, err
				}
				command.Stdin = stdin
			}
		}
	}
	shouldWait := !cmd.Sync
	if cmd.Times > 0 {
		shouldWait = false
	}

	p := &Process{
		Cmd:        command,
		Command:    cmd,
		ShouldWait: shouldWait,
	}

	p.Kill = func() {
		if p.killed {
			return
		}
		p.killed = true
		if !shouldWait {
			p.Cmd.Process.Signal(os.Interrupt)
		}
		if stdin != nil {
			stdin.Close()
		}
		if stdout != nil {
			stdout.Close()
		}
		if stderr != nil {
			stderr.Close()
		}
	}

	// certain amount of times
	if cmd.Times > 0 {
		p.timesRunFunc()
		return p, nil
	}

	// monotonic
	if cmd.Every > 0 {
		p.monotonicRunFunc()
		return p, nil
	}

	// sync or async, doesn't matter here
	p.Run = func() error {
		defer p.Kill()
		return p.Cmd.Run()
	}
	return p, nil
}

func (p *Process) LogError(err error) {
	if err == nil {
		return
	}
	var s string
	if p.Command.Name != "" {
		s = fmt.Sprintf("command %s: %s", p.Command.Name, err)
	} else {
		s = err.Error()
	}
	log.Println(s)
}

func (p *Process) Refresh() {
	cmd := exec.Command(p.Command.Command, p.Command.Args...)
	cmd.Dir = p.Cmd.Dir
	if p.Cmd.Stdin != nil {
		cmd.Stdin = p.Cmd.Stdin
	}
	if p.Cmd.Stderr != nil {
		cmd.Stderr = p.Cmd.Stderr
	}
	if p.Cmd.Stdout != nil {
		cmd.Stdout = p.Cmd.Stdout
	}
	p.Cmd.Process.Signal(os.Interrupt)
	p.Cmd = cmd
}

func (p *Process) timesRunFunc() {
	p.Run = func() error {
		defer p.Kill()
		for i := 0; i < p.Command.Times; i++ {
			err := p.Cmd.Run()
			if err != nil {
				return err
			}
			if i+1 < p.Command.Times {
				p.Refresh()
			}
		}
		return nil
	}
}

func (p *Process) monotonicRunFunc() {
	p.Run = func() (err error) {
		defer p.Kill()
		ticker := time.NewTicker(p.Command.Every)
		defer ticker.Stop()
		done := make(chan error, 5)
		run := func() {
			done <- p.Cmd.Run()
		}
		go func() {
			for {
				go run()
				select {
				case <-ticker.C:
					p.Refresh()
				case err = <-done:
					if err != nil {
						p.LogError(err)
						return
					}
					p.Refresh()
				}
			}
		}()
		return
	}
}
