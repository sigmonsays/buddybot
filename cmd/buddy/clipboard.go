package main

import (
	"github.com/sigmonsays/buddybot"
	"github.com/sigmonsays/buddybot/clipboard"
)

func (me *CommandSet) Clipboard(m *buddybot.Message, ctx *Context, cline *CommandLine) error {

	cline.Args = cline.SliceArgs(1)
	cmd := cline.Arg(0)

	xclip := clipboard.NewXClip()
	if cmd == "get" {

		value, err := xclip.GetString()
		if err != nil {
			return ctx.Reply(m, "error: GetString %s", err)
		}

		return ctx.Reply(m, "%q", value)

	} else if cmd == "set" {
		value := cline.Arg(1)
		err := xclip.SetString(value)
		if err != nil {
			return ctx.Reply(m, "error: %s", err)
		}

		return ctx.Reply(m, "-- clipboard set -- ")

	}

	return nil
}
