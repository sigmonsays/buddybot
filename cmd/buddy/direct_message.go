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

	if mention != me.identity.Nick {
		log.Debugf("not for me (@ mention %s)", mention)
		return nil
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

// execute a shell command and send response back
func (me *handler) execMessage(m *buddybot.Message, ctx *Context, line []string) error {
	log.Debugf("exec message - exec %q", line)

	se := NewShellExec()
	res, err := se.ExecMessage(line)
	if err != nil {
		ctx.SendTo(m.Id, buddybot.NewNotice("%s", err))
		return nil
	}

	log.Debugf("execMessage pid=%d", res.Pid)

	var buf string
	var ok bool
	msg := ctx.NewMessage()
Dance:
	for {
		select {
		case buf, ok = <-res.Stdout:
			if ok == false {
				break Dance
			}
			msg.Message = "<stdout> " + strings.TrimRight(buf, "\n")
			ctx.SendTo(m.Id, msg)
		case buf, ok = <-res.Stderr:
			if ok == false {
				break Dance
			}
			msg.Message = "<stderr> " + strings.TrimRight(buf, "\n")
			ctx.SendTo(m.Id, msg)
		}
	}

	err = res.Wait()
	if err != nil {
		log.Warnf("Wait returned %s", err)
	}

	log.Debugf("finished exec - %q", line)

	return nil
}
