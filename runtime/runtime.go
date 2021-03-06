package runtime

import (
	"fmt"
	"github.com/insomnimus/inscript/ast"
	"log"
	"os"
	"os/exec"
	"time"
)

type Process struct {
	Cmd     *exec.Cmd
	Command *ast.Command
	Async   bool
	Kill    func()
	Run     func() error
	killed  bool
}

func CreateProcess(cmd *ast.Command) (*Process, error) {
	command := exec.Command(cmd.Command, cmd.Args...)
	if cmd.Dir != "" {
		command.Dir = cmd.Dir
	}
	var stdin, stdout, stderr *File
	var err error
	if cmd.Stderr != "" {
		switch {
		case cmd.Stderr == "!stderr":
			command.Stderr = os.Stderr
		case cmd.Stderr == "!stdout":
			command.Stderr = os.Stdout
		default:
			if file, ok := LookupFile(cmd.Dir + cmd.Stderr); ok {
				command.Stderr = file.File
				stderr = file
				file.Add()
				break
			}
			var file *os.File
			if _, e := os.Stat(cmd.Dir + cmd.Stderr); os.IsNotExist(e) {
				file, err = os.Create(cmd.Dir + cmd.Stderr)
			} else {
				file, err = os.OpenFile(cmd.Dir+cmd.Stderr, os.O_WRONLY, 0764)
			}
			if err != nil {
				return nil, err
			}
			command.Stderr = file
			stderr = RegisterFile(cmd.Dir+cmd.Stderr, file)
		}
	}

	if cmd.Stdout != "" {
		switch {
		case cmd.Stdout == "!stdout":
			command.Stdout = os.Stdout
		case cmd.Stdout == "!stderr":
			command.Stdout = os.Stderr
		case cmd.Stdout == cmd.Stderr && stderr != nil:
			command.Stdout = stderr.File
		default:
			if file, ok := LookupFile(cmd.Dir + cmd.Stdout); ok {
				file.Add()
				command.Stdout = file.File
				stdout = file
				break
			}
			var file *os.File
			if _, e := os.Stat(cmd.Dir + cmd.Stdout); os.IsNotExist(e) {
				file, err = os.Create(cmd.Dir + cmd.Stdout)
			} else {
				file, err = os.OpenFile(cmd.Dir+cmd.Stdout, os.O_WRONLY, 0764)
			}
			if err != nil {
				return nil, err
			}
			command.Stdout = file
			stdout = RegisterFile(cmd.Dir+cmd.Stdout, file)
		}
	}

	if cmd.Stdin != "" {
		switch {
		case cmd.Stdin == "!stdin":
			command.Stdin = os.Stdin
		case cmd.Stdin != cmd.Stdout && cmd.Stdin != cmd.Stderr:
			if file, ok := LookupFile(cmd.Dir + cmd.Stdin); ok {
				command.Stdin = file.File
				stdin = file
				file.Add()
				break
			}
			var file *os.File
			if _, e := os.Stat(cmd.Dir + cmd.Stdin); !os.IsNotExist(e) {
				file, err = os.Open(cmd.Dir + cmd.Stdin)
				if err != nil {
					return nil, err
				}
				command.Stdin = file
				stdin = RegisterFile(cmd.Dir+cmd.Stdin, file)
			}
		}
	}

	async := !cmd.Sync
	if cmd.Every > 0 && cmd.Times == 0 {
		async = true
	}

	p := &Process{
		Cmd:     command,
		Command: cmd,
		Async:   async,
	}

	p.Kill = func() {
		if p.Cmd == nil || p.Cmd.Process == nil {
			p.killed = true
		}
		if p.killed {
			return
		}
		p.killed = true
		if p.Async {
			// process.Signal(sigint) does nothing on windows so we kill instead
			if os.PathSeparator == '\\' {
				p.Cmd.Process.Kill()
			} else {
				p.Cmd.Process.Signal(os.Interrupt)
			}
		}
		if stdin != nil {
			stdin.Done()
		}
		if stdout != nil {
			stdout.Done()
		}
		if stderr != nil {
			stderr.Done()
		}
	}

	// monotonic and certain amount of iterations
	if cmd.Every > 0 && cmd.Times > 0 {
		p.monotonicTimesRunFunc()
		return p, nil
	}

	// certain amount of times
	if cmd.Times > 0 {
		p.timesRunFunc()
		return p, nil
	}

	// monotonic
	if cmd.Every > 0 {
		p.Async = true
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

func (p *Process) _monotonicRunFunc() {
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

func (p *Process) _monotonicTimesRunFunc() {
	p.Run = func() (err error) {
		defer p.Kill()
		ticker := time.NewTicker(p.Command.Every)
		defer ticker.Stop()
		done := make(chan error, 5)
		run := func() {
			done <- p.Cmd.Run()
		}
		go func() {
			for i := 0; i < p.Command.Times; i++ {
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

func (p *Process) monotonicTimesRunFunc() {
	p.Run = func() (err error) {
		defer p.Kill()
		for i := 0; i < p.Command.Times; i++ {
			err = p.Cmd.Run()
			if err != nil {
				return
			}
			if i+1 == p.Command.Times {
				break
			}
			time.Sleep(p.Command.Every)
			p.Refresh()
		}
		return
	}
}

func (p *Process) monotonicRunFunc() {
	p.Run = func() error {
		defer p.Kill()
		for {
			err := p.Cmd.Run()
			if err != nil {
				return err
			}
			p.Refresh()
			time.Sleep(p.Command.Every)
		}
	}
}
