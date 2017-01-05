package main

import (
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

func (me *handler) OnMessage(m *buddybot.Message, ctx *Context) error {
	log.Debugf("got message %s", m)

	if strings.HasPrefix(m.Message, "@") {
	}

	return nil
}
