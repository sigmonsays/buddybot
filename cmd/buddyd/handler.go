package main

import (
	"container/list"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sigmonsays/buddybot"
)

type chatHandler struct {
	mx        sync.Mutex
	staticDir string
	history   *list.List
}

func (h *chatHandler) serveHome(w http.ResponseWriter, r *http.Request) {
	log.Infof("request %s", r.URL)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	home_html := filepath.Join(h.staticDir, "home.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl := template.Must(template.ParseFiles(home_html))
	homeTempl.Execute(w, r.Host)
}

func (h *chatHandler) addHistory(m *buddybot.Message) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.history.PushBack(m)
	if h.history.Len() > 5 {
		if e := h.history.Front(); e != nil {
			h.history.Remove(e)
		}
	}
}

func (h *chatHandler) getHistory() []*buddybot.Message {
	h.mx.Lock()
	defer h.mx.Unlock()
	ret := make([]*buddybot.Message, 0)
	for e := h.history.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(*buddybot.Message))
	}
	return ret
}

// entry point for message handling
// handleMessage will call handle{Operation}Op for the appropriate buddybot.OpCode
// if a message starts with "/" is is dispatched to handleServerCommand

func (h *chatHandler) handleMessage(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	if strings.HasPrefix(m.Message, "/") {
		return h.handleServerCommand(op, hub, c, m)
	}

	// dispatch appropriate operation
	if op == buddybot.MessageOp {
		h.handleMessageOp(op, hub, c, m)

	} else if op == buddybot.RegisterOp {
		h.handleRegisterOp(op, hub, c, m)

	} else if op == buddybot.UnregisterOp {
		h.handleUnregisterOp(op, hub, c, m)

	} else if op == buddybot.JoinOp {
		h.handleJoinOp(op, hub, c, m)

	} else if op == buddybot.NickOp {
		h.handleNickOp(op, hub, c, m)

	} else if op == buddybot.PingOp {
		h.handlePingOp(op, hub, c, m)

	} else if op == buddybot.NoticeOp {
		h.handleNoticeOp(op, hub, c, m)

	} else if op == buddybot.ClientListOp {
		h.handleClientListOp(op, hub, c, m)

	} else {
		log.Infof("Unhandled op %s/%d: %+v", m.Op, m.Op, m)
	}
	return nil
}

func (h *chatHandler) handleMessageOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleMessage %s/%d msg:%s", op, op, m)
	hub.SendBroadcast(m)
	return nil
}

func (h *chatHandler) handleUnregisterOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	return nil
}

func (h *chatHandler) handleRegisterOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has left", c.Identity))
	return nil
}

func (h *chatHandler) setConnectionIdentity(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	if m.From == "" {
		return fmt.Errorf("Join without a name (From not set)")
	}

	c.Identity = m.From
	id, err := buddybot.ParseIdentity(m.Message)
	if err == nil {
		log.Infof("connection %d is now known as %v", c.GetId(), id)
	} else {
		m := &buddybot.Message{
			Op:      buddybot.NoticeOp,
			Message: "You provided a bad identity: " + err.Error(),
		}
		hub.SendTo(c, m)
	}
	return nil
}

func (h *chatHandler) handleJoinOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {

	h.setConnectionIdentity(op, hub, c, m)

	// send out a broadcast so each client knows
	log.Infof("connection %d is now known as %q", c.GetId(), m.From)
	hub.SendBroadcast(m)
	return nil
}

func (h *chatHandler) handleNickOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	if c.Identity == "" {
		hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has joined", m.From))
	} else {
		hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has changed their name to %s", c.Identity, m.From))
	}
	c.Identity = m.From
	h.setConnectionIdentity(op, hub, c, m)
	return nil
}

func (h *chatHandler) handlePingOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	return nil
}

func (h *chatHandler) handleNoticeOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	hub.SendBroadcast(m)
	return nil
}

func (h *chatHandler) handleClientListOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	err := hub.SendClientList(m.Id)
	if err != nil {
		log.Infof("ClientList: %s", err)
	}
	return nil
}

func (h *chatHandler) handleServerCommand(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleServerCommand %s/%d msg:%s", op, op, m)

	return nil
}
