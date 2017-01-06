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

	if strings.HasPrefix(m.Message, "@") {
		return me.DirectMessage(m, ctx)
	}

	if m.Message == "ping" {
		log.Debugf("sending pong...")
		ctx.SendMessage("pong")
	}
	return nil
}

// @nick <blah> is a direct message
//
// more precisely only the node <nick> should response
//
func (me *handler) DirectMessage(m *buddybot.Message, ctx *Context) error {
	log.Debugf("direct message %s", m)

	return nil
}
