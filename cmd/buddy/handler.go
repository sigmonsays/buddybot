package main

import (
	"fmt"
	"strings"

	"github.com/sigmonsays/buddybot"
)

func NewHandler(identity *buddybot.Identity) *handler {
	h := &handler{
		identity: identity,
		commands: NewCommandSet(),
	}
	return h
}

type handler struct {
	identity *buddybot.Identity
	commands *CommandSet
}

func (me *handler) OnNotice(m *buddybot.Message, ctx *Context) error {
	log.Debugf("got notice %s", m)
	fmt.Printf("NOTICE <%s> %s\n", m.From, m.Message)
	return nil
}

// handles both direct messages and broadcast messages
func (me *handler) OnMessage(m *buddybot.Message, ctx *Context) error {
	log.Debugf("got message %s", m)
	fmt.Printf("MESSAGE <%s> %s\n", m.From, m.Message)

	line := m.Message

	if strings.HasPrefix(line, "@") {
		return me.DirectMessage(m, ctx)
	}

	cline, err := ParseCommandLine(line)
	if err != nil {
		log.Warnf("parse command %q failed. %s", line, err)
		return nil
	}

	me.commands.Dispatch(m, ctx, cline)
	return nil
}
