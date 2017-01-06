package main

import (
	"fmt"
	"strings"

	"github.com/sigmonsays/buddybot"
)

func NewHandler(identity *buddybot.Identity) *handler {
	h := &handler{
		identity: identity,
	}
	return h
}

type handler struct {
	identity *buddybot.Identity
}

func (me *handler) OnNotice(m *buddybot.Message, ctx *Context) error {
	log.Debugf("got notice %s", m)
	fmt.Printf("NOTICE <%s> %s\n", m.From, m.Message)
	return nil
}

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

	if cline.Arg0 == "ping" {

		log.Debugf("sending pong...")
		ctx.BroadcastMessage("pong")

	} else if cline.Arg0 == "exec" {

		cline.Args = cline.SliceArgs(1)

		return me.execMessage(m, ctx, cline.Args)

	} else if cline.Arg0 == "echo" {
		fmt.Printf("%s\n", line)
	}

	return nil
}
