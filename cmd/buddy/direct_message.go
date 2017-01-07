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

	m.From = mention

	if mention != "all" && mention != me.identity.Nick {
		log.Debugf("not for me (@ mention %s)", mention)
		return nil
	}

	if mention == "all" {
		m.From = ""
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
