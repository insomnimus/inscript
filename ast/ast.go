package ast

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	Command               string
	Args                  []string
	Dir                   string
	Name                  string
	Stdin, Stdout, Stderr string
	Sync                  bool
	Every                 time.Duration
	Times                 int
}

func (a Command) Equal(b Command) bool {
	if a.Name != b.Name ||
		a.Command != b.Command ||
		a.Stdin != b.Stdin ||
		a.Stdout != b.Stdout ||
		a.Dir != b.Dir ||
		a.Sync != b.Sync ||
		a.Every != b.Every ||
		a.Times != b.Times ||
		len(a.Args) != len(b.Args) {
		return false
	}
	for i, arg := range a.Args {
		if arg != b.Args[i] {
			return false
		}
	}
	return true
}

func (c Command) GoString() string {
	var buff strings.Builder
	fmt.Fprintf(&buff, "Command{\n\tCommand: %q,\n", c.Command)
	if len(c.Args) > 0 {
		fmt.Fprintf(&buff, "\tArgs: %v,\n", c.Args)
	}
	if c.Dir != "" {
		fmt.Fprintf(&buff, "\tDir: %q,\n", c.Dir)
	}
	if c.Name != "" {
		fmt.Fprintf(&buff, "\tName: %q,\n", c.Name)
	}
	fmt.Fprintf(&buff, "\tSync: %t,\n", c.Sync)
	if c.Stdin != "" {
		fmt.Fprintf(&buff, "\tStdin: %q,\n", c.Stdin)
	}
	if c.Stdout != "" {
		fmt.Fprintf(&buff, "\tStdout: %q,\n", c.Stdout)
	}
	if c.Stderr != "" {
		fmt.Fprintf(&buff, "\tStderr: %q,\n", c.Stderr)
	}
	if c.Every > 0 {
		fmt.Fprintf(&buff, "\tEvery: %d,\n", c.Every)
	}
	if c.Times > 0 {
		fmt.Fprintf(&buff, "\tTimes: %d,\n", c.Times)
	}
	buff.WriteRune('}')
	return buff.String()
}
