package main

import (
	"fmt"

	"github.com/sigmonsays/buddybot"
)

func NewCommandSet() *CommandSet {
	s := &CommandSet{
		cmds: make(map[string]CommandFunc, 0),
	}
	s.init()
	return s
}

type CommandFunc func(m *buddybot.Message, ctx *Context, cline *CommandLine) error
type CommandSet struct {
	cmds map[string]CommandFunc
}

func (me *CommandSet) init() {
	c := me.cmds
	c["ping"] = me.Ping
	c["exec"] = me.Exec
	c["echo"] = me.Echo
	c["clipboard"] = me.Clipboard
}

func (me *CommandSet) Dispatch(m *buddybot.Message, ctx *Context, cline *CommandLine) error {
	log.Tracef("dispatch %s", cline)
	cmd, ok := me.cmds[cline.Arg0]
	if ok == false {
		return fmt.Errorf("command not found: %s", cline.Arg0)
	}
	return cmd(m, ctx, cline)
}

func (me *CommandSet) Ping(m *buddybot.Message, ctx *Context, cline *CommandLine) error {
	log.Debugf("sending pong...")
	ctx.BroadcastMessage("pong")
	return nil
}

func (me *CommandSet) Echo(m *buddybot.Message, ctx *Context, cline *CommandLine) error {
	fmt.Printf("%s\n", cline)
	return nil
}

func (me *CommandSet) Exec(m *buddybot.Message, ctx *Context, cline *CommandLine) error {
	cline.Args = cline.SliceArgs(1)
	return me.execMessage(m.FromIdentity().Nick, m, ctx, cline.Args)
}
