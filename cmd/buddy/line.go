package main

import (
	"fmt"

	"github.com/mattn/go-shellwords"
)

type CommandLine struct {
	line string
	Arg0 string
	Args []string
}

func (me *CommandLine) String() string {
	return fmt.Sprintf("%s", me.Args)
}

func ParseCommandLine(line string) (*CommandLine, error) {
	cl := &CommandLine{
		line: line,
	}

	args, err := shellwords.Parse(line)
	if err != nil {
		return cl, err
	}

	cl.Args = args
	if len(args) > 0 {
		cl.Arg0 = args[0]
	}

	return cl, nil
}
func (me *CommandLine) Arg(i int) string {
	if i < len(me.Args) {
		return me.Args[i]
	}
	return ""
}

func (me *CommandLine) SliceArgs(offset int) []string {
	log.Tracef("slice args len=%d offset=%d", len(me.Args), offset)
	if len(me.Args) >= offset {
		me.Args = me.Args[offset:]
		log.Debugf("new args (len %d) %q", len(me.Args), me.Args)
	}
	return me.Args
}
