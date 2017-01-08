package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/sigmonsays/buddybot"
)

type Context struct {
	Identity *buddybot.Identity
	Conn     *websocket.Conn
	mx       sync.Mutex
}

func (me *Context) NewMessage() *buddybot.Message {
	m := &buddybot.Message{}
	m.Id = 0
	m.Op = buddybot.MessageOp
	m.From = me.Identity.String()
	return m
}

func (me *Context) BroadcastMessage(msg string) error {
	me.mx.Lock()
	defer me.mx.Unlock()

	m := me.NewMessage()
	m.Message = msg

	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = me.Conn.WriteMessage(websocket.TextMessage, buf)
	return err
}

func (me *Context) Send(m *buddybot.Message) error {
	me.mx.Lock()
	defer me.mx.Unlock()
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}

	log.Tracef("Send cid=%d: %s", m.Id, buf)
	err = me.Conn.WriteMessage(websocket.TextMessage, buf)
	return err
}

// convenience function to reply to a given message
func (me *Context) Reply(m *buddybot.Message, s string, args ...interface{}) error {
	reply := m.Copy()
	reply.Op = buddybot.RawMessageOp
	reply.Message = fmt.Sprintf(s, args...)
	return me.Send(reply)
}

func (me *Context) SendTo(cid int64, m *buddybot.Message) error {
	me.mx.Lock()
	defer me.mx.Unlock()
	m.Id = cid

	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}

	log.Tracef("SendTo cid=%d: %s", cid, buf)

	err = me.Conn.WriteMessage(websocket.TextMessage, buf)
	return err
}
