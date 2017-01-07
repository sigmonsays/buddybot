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

	} else if op == buddybot.DirectMessageOp {
		h.handleDirectMessageOp(op, hub, c, m)

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

// store the nickname
func (h *chatHandler) storeNick(hub *buddybot.Hub, c *buddybot.Connection, id *buddybot.Identity) error {
	existing_id, ok := h.nicknames[id.Nick]
	if ok {

		existing_conn, err := h.hub.FindConnection(existing_id)
		if err == nil {
			if existing_conn.GetId() != existing_id {
				log.Warnf("nick name is already taken: %s by connection %s", id.Nick, existing_conn)
				hub.Send(buddybot.NoticeOp, fmt.Sprintf("Nick name is already taken: %s", id.Nick))
			}
		}

	} else {
		cid := c.GetId()
		h.nicknames[id.Nick] = cid
		log.Infof("connection %s is now known as %q", c, id.Nick)
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
		h.storeNick(hub, c, id)

	} else {
		m := &buddybot.Message{
			Op:      buddybot.NoticeOp,
			Message: "You provided a bad identity: " + err.Error(),
		}
		hub.SendTo(c, m)
	}

	return id, nil
}

func (h *chatHandler) handleDirectMessageOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleMessage %s/%d msg:%s", op, op, m)
	if m.From == "" {
		log.Warnf("dropping direct message without from address: %s", m)
		return nil
	}
	if m.To == "" {
		log.Warnf("dropping direct message without to address: %s", m)
		return nil
	}

	var (
		sconn *buddybot.Connection
		dconn *buddybot.Connection
		err   error
	)

	from_nick := m.FromIdentity().Nick
	sconn, err = h.findNick(from_nick)
	if err == nil {
		m.IdFrom = sconn.GetId()
		log.Tracef("set message From=%s: connection id=%d", from_nick, m.IdFrom)
	} else {
		log.Warnf("findNick from=%s: %s", from_nick, err)
	}

	to_nick := m.ToIdentity().Nick
	dconn, err = h.findNick(to_nick)
	if err != nil {
		log.Warnf("no destination connection for nick %s: %s", to_nick, err)
		return nil
	}

	m.IdTo = dconn.GetId()
	log.Tracef("set message To=%s: connection id=%d", to_nick, m.IdTo)
	hub.SendTo(dconn, m)
	return nil
}

// this message operation is broadcasted to everyone
func (h *chatHandler) handleMessageOp(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Debugf("handleMessage %s/%d msg:%s", op, op, m)

	if m.From == "" {
		log.Warnf("dropping message without from address: %s", m)
		return nil
	}

	hub.SendBroadcast(m)

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
	hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has disconnected", c.Identity))
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
