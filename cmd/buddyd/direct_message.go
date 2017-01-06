package main

import (
	"fmt"
	"strings"

	"github.com/sigmonsays/buddybot"
)

func (h *chatHandler) handleDirectMessage(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleDirectMessage %s/%d msg:%s", op, op, m)

	var who string

	tmp := strings.SplitN(m.Message, " ", 2)
	if len(tmp) > 0 {
		if len(tmp[0]) > 1 {
			who = tmp[0][1:]
		}
	}

	if who == "" {
		return fmt.Errorf("bad format for #username direct message")
	}

	id, err := h.findNick(who)
	if err != nil {
		return err
	}

	dest, err := h.hub.FindConnection(id.GetId())
	if err != nil {
		return err
	}

	err = hub.SendTo(dest, m)
	if err != nil {
		return err
	}

	return nil
}
