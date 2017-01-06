package main

import (
	"strings"

	"github.com/sigmonsays/buddybot"
)

// @nick <blah> is a direct message
//
// more precisely only the node <nick> should response
//
func (me *handler) DirectMessage(m *buddybot.Message, ctx *Context) error {
	log.Debugf("direct message %s", m)

	line := m.Message
	if strings.HasPrefix(m.Message, "@") {
		offset := strings.Index(m.Message, " ")
		if offset > 0 {
			line = m.Message[offset:]
		}
	}

	cline, err := ParseCommandLine(line)
	if err != nil {
		log.Warnf("ParseCommandLine: %s", err)
		return nil
	}

	cline.Args = cline.SliceArgs(1)

	me.execMessage(m, ctx, cline.Args)

	return nil
}

func (me *handler) execMessage(m *buddybot.Message, ctx *Context, line []string) error {
	log.Debugf("exec message - exec %q", line)

	se := NewShellExec()
	res, err := se.ExecMessage(line)
	if err != nil {
		ctx.SendTo(m.Id, buddybot.NewNotice("%s", err))
		return nil
	}

	log.Debugf("execMessage %#v", res)

	return nil
}
