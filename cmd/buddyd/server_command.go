package main

import (
	"fmt"
	"strconv"
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

func (h *chatHandler) sendNotice(hub *buddybot.Hub, c *buddybot.Connection, s string, args ...interface{}) error {
	m := &buddybot.Message{
		Op:      buddybot.NoticeOp,
		Message: fmt.Sprintf(s, args...),
	}
	hub.SendTo(c, m)
	return nil
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

	} else if cmd == "list-opcodes" {
		for _, opcode := range buddybot.OpCodes() {
			h.sendNotice(hub, c, "opcode %d %s", opcode, opcode)
		}

	} else if cmd == "send-message" {

		m := buddybot.NewMessage()

		opcode, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		m.Op = buddybot.OpCode(opcode)
		m.From = args[2]
		m.To = args[3]
		m.Message = args[4]
		log.Infof("Sending server generated message %s", m)
		hub.SendBroadcast(m)

	} else {
		log.Warnf("unknown command: cmd=%s", args)
		h.sendNotice(hub, c, "No such command: %s", cmd)
	}

	return nil
}
