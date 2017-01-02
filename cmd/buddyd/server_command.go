package main

import (
	"fmt"
	"strings"

	"github.com/sigmonsays/buddybot"

	"github.com/mattn/go-shellwords"
)

func parseCommand(line string) ([]string, error) {
	var cmdline string
	var args []string
	if strings.HasPrefix(line, "/") == false {
		return args, fmt.Errorf("not a server command")
	}

	if len(line) > 0 {
		cmdline = line[1:]
	}

	if len(cmdline) > 0 {
		a, err := shellwords.Parse(cmdline)
		if err != nil {
			return args, err
		}
		args = a
	}

	log.Tracef("cmdline %q: arguments parsed: %#v", cmdline, args)

	return args, nil
}

func (h *chatHandler) handleServerCommand(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleServerCommand %s/%d msg:%s", op, op, m)

	args, err := parseCommand(m.Message)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := args[0]
	log.Debugf("dispatch cmd=%s args=%s", cmd, args)

	if cmd == "who" {
		hub.SendClientList(c.GetId())

	} else {
		log.Warnf("unknown command: cmd=%s", args)

		m := &buddybot.Message{
			Op:      buddybot.NoticeOp,
			Message: fmt.Sprintf("No such command: %s", cmd),
		}
		hub.SendTo(c, m)
	}

	return nil
}
