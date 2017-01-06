package main

import (
	"strings"

	"github.com/sigmonsays/buddybot"

	"github.com/mattn/go-shellwords"
)

func (h *chatHandler) handleServerCommand(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleServerCommand %s/%d msg:%s", op, op, m)

	var cmdline string
	var args []string

	tmp := strings.Split(m.Message, "/")
	if len(tmp) < 2 {
		cmdline = tmp[1]
	}
	if len(cmdline) > 0 {
		a, err := shellwords.Parse(cmdline)
		if err != nil {
			return err
		}
		args = a
	}

	log.Debugf("arguments parsed: %#v", args)

	return nil
}
