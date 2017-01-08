package main

import (
	"github.com/sigmonsays/buddybot"
	"github.com/sigmonsays/buddybot/clipboard"
)

func (me *CommandSet) Clipboard(m *buddybot.Message, ctx *Context, cline *CommandLine) error {

	cline.Args = cline.SliceArgs(1)
	cmd := cline.Arg(0)

	clip := clipboard.NewClipboard()
	if cmd == "get" {

		value, err := clip.GetString()
		if err != nil {
			return ctx.Reply(m, "error: GetString %s", err)
		}

		return ctx.Reply(m, "%q", value)

	} else if cmd == "set" {
		value := cline.Arg(1)
		err := clip.SetString(value)
		if err != nil {
			return ctx.Reply(m, "error: %s", err)
		}

		return ctx.Reply(m, "-- clipboard set -- ")

	}

	return nil
}
