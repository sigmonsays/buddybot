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

	var mention string
	line := m.Message
	if strings.HasPrefix(m.Message, "@") {
		offset := strings.Index(m.Message, " ")
		if offset > 0 {
			line = m.Message[offset:]
		}
		mention = m.Message[1:offset]
	}

	if mention != "all" && mention != me.identity.Nick {
		log.Debugf("not for me (@ mention %s)", mention)
		return nil
	}

	cline, err := ParseCommandLine(line)
	if err != nil {
		log.Warnf("ParseCommandLine: %s", err)
		return nil
	}

	cmd := cline.Arg(0)

	if cmd == "" {
		log.Warnf("no command given: %s", line)
		return nil
	}

	if cmd == "exec" {
		cline.Args = cline.SliceArgs(1)
		me.execMessage(m.FromIdentity().Nick, m, ctx, cline.Args)

	} else {
		reply := m.Reply()
		log.Infof("No such command: %s", line)
		ctx.Send(reply.WithMessage("no such command: %s", line))

	}

	return nil
}
