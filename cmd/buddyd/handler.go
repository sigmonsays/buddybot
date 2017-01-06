package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sigmonsays/buddybot"
)

func NewChatHandler(hub *buddybot.Hub, opts *ChatHandlerOptions) *chatHandler {
	handler := &chatHandler{
		hub:       hub,
		staticDir: opts.StaticDir,
		nicknames: make(map[string]int64, 0),
	}
	return handler
}

func DefaultChatHandlerOptions() *ChatHandlerOptions {
	o := &ChatHandlerOptions{
		StaticDir: "/tmp/chat",
	}
	return o
}

type ChatHandlerOptions struct {
	StaticDir string
}

type chatHandler struct {
	hub       *buddybot.Hub
	staticDir string

	// nickname associates to connection id
	nicknames map[string]int64
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

func (h *chatHandler) findNick(name string) (*buddybot.Connection, error) {

	cid, ok := h.nicknames[name]
	if ok == false {
		return nil, fmt.Errorf("not found: %s", name)
	}

	c, err := h.hub.FindConnection(cid)

	return c, err
}

// entry point for message handling
//
// 	- handleMessage will call handle{Operation}Op for the appropriate buddybot.OpCode
// 	- if a message starts with "/" is is dispatched to handleServerCommand
//
func (h *chatHandler) handleMessage(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	if m == nil {
		// just make an empty message
		m = &buddybot.Message{
			Op: buddybot.InvalidOp,
		}
	}

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

func (h *chatHandler) setConnectionIdentity(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) (*buddybot.Identity, error) {
	if m.From == "" {
		return nil, fmt.Errorf("Join without a name (From not set)")
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

	return id, nil
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

	id, err := h.setConnectionIdentity(op, hub, c, m)
	if err != nil {
		log.Warnf("setConnectionIdentity cid:%d: %s", c.GetId(), err)
	}

	// store the nickname
	if id != nil {
		if _, ok := h.nicknames[id.Nick]; ok {
			hub.Send(buddybot.NoticeOp, fmt.Sprintf("Nick name is already taken: %s", id.Nick))
			return nil
		}

		cid := c.GetId()
		h.nicknames[id.Nick] = cid
	}

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
