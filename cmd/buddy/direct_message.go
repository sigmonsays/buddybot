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

	err = me.commands.Dispatch(m, ctx, cline)
	if err != nil {
		reply := m.Reply()
		if mention == "all" {
			reply.Op = buddybot.MessageOp
			reply.To = ""
		} else {
			reply.Op = buddybot.DirectMessageOp
			reply.To = mention
		}
		log.Infof("dispatch %s error: %s", line, err)
		ctx.Send(reply.WithMessage("dispatch %s error: %s", line, err))
	}

	return nil
}
