package main

import (
	"strings"

	"github.com/sigmonsays/buddybot"
)

// execute a shell command and send response back
func (me *handler) execMessage(from string, m *buddybot.Message, ctx *Context, line []string) error {
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
	msg.To = from

Dance:
	for {
		select {
		case buf, ok = <-res.Stdout:
			if ok == false {
				res.Stdout = nil
				continue
			}
			msg.Message = "<stdout> " + strings.TrimRight(buf, "\n")
			ctx.SendTo(m.Id, msg)

		case buf, ok = <-res.Stderr:
			if ok == false {
				res.Stderr = nil
				continue
			}
			msg.Message = "<stderr> " + strings.TrimRight(buf, "\n")
			ctx.SendTo(m.Id, msg)

		case <-res.Done:
			break Dance
		}
	}

	log.Debugf("finished exec - %q", line)

	return nil
}
