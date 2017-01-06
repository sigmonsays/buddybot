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
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	log.Infof("request %s %s", r.Method, r.URL)
	home_html := filepath.Join(h.staticDir, "home.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl := template.Must(template.ParseFiles(home_html))
	homeTempl.Execute(w, r.Host)
}

func (h *chatHandler) findNick(name string) (*buddybot.Connection, error) {

	cid, ok := h.nicknames[name]
	if ok == false {
		return nil, fmt.Errorf("nick not found: %s", name)
	}

	c, err := h.hub.FindConnection(cid)

	return c, err
}

// entry point for message handling (received messages)
//
// 	- handleMessage will call handle{Operation}Op for the appropriate buddybot.OpCode
// 	- if a message starts with "/" is is dispatched to handleServerCommand
// 	- if a message starts with "#" is is dispatched to handleDirectMessage
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
	if strings.HasPrefix(m.Message, "#") {
		return h.handleDirectMessage(op, hub, c, m)
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

func (h *chatHandler) setConnectionIdentity(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) (*buddybot.Identity, error) {
	if m.From == "" {
		return nil, fmt.Errorf("Join without a name (From not set)")
	}

	c.Identity = m.From
	id, err := buddybot.ParseIdentity(m.From)
	if err == nil {
		log.Infof("connection %d is now known as %v (nick %s)", c.GetId(), id, id.Nick)

		// store the nickname
		existing_id, ok := h.nicknames[id.Nick]
		if ok {

			existing_conn, err := h.hub.FindConnection(existing_id)
			if err == nil {
				if existing_conn.GetId() != existing_id {
					hub.Send(buddybot.NoticeOp, fmt.Sprintf("Nick name is already taken: %s", id.Nick))
				}
			}

		} else {
			cid := c.GetId()
			h.nicknames[id.Nick] = cid
		}

	} else {
		m := &buddybot.Message{
			Op:      buddybot.NoticeOp,
			Message: "You provided a bad identity: " + err.Error(),
		}
		hub.SendTo(c, m)
	}

	return id, nil
}

func (h *chatHandler) handleMessageOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleMessage %s/%d msg:%s", op, op, m)

	var (
		sconn *buddybot.Connection
		dconn *buddybot.Connection
		err   error
	)

	if m.From != "" {
		sconn, err = h.findNick(m.From)
		if err == nil {
			m.IdFrom = sconn.GetId()
			log.Tracef("set message From=%s: connection id=%d", m.From, m.IdFrom)
		} else {
			log.Warnf("findNick from=%s: %s", m.From, err)
		}
	}

	if m.To != "" {
		dconn, err = h.findNick(m.To)
		if err == nil {
			m.IdTo = dconn.GetId()
			log.Tracef("set message To=%s: connection id=%d", m.To, m.IdTo)
		} else {
			log.Warnf("findNick to=%s: %s", m.To, err)
		}
	}

	if m.IdTo == 0 {
		hub.SendBroadcast(m)
	} else {
		hub.SendTo(dconn, m)
	}
	return nil
}

func (h *chatHandler) handleUnregisterOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	id, err := buddybot.ParseIdentity(m.From)
	if err == nil {
		log.Infof("connection %d is finished (nick %s)", c.GetId(), id, id.Nick)
		delete(h.nicknames, id.Nick)
	}

	return nil
}

func (h *chatHandler) handleRegisterOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has left", c.Identity))
	return nil
}

func (h *chatHandler) handleJoinOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {

	_, err := h.setConnectionIdentity(op, hub, c, m)
	if err != nil {
		log.Warnf("JoinOp: setConnectionIdentity cid:%d: %s", c.GetId(), err)
	}

	// send out a broadcast so each client knows
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

	_, err := h.setConnectionIdentity(op, hub, c, m)
	if err != nil {
		log.Warnf("NickOp: setConnectionIdentity cid:%d: %s", c.GetId(), err)
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
