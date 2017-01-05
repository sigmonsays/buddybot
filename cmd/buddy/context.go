package main

import (
	"encoding/json"

	"github.com/gorilla/websocket"

	"github.com/sigmonsays/buddybot"
)

type Context struct {
	Identity *buddybot.Identity
	Conn     *websocket.Conn
}

func (me *Context) NewMessage() *buddybot.Message {
	m := &buddybot.Message{}
	m.Id = 0
	m.Op = buddybot.MessageOp
	m.From = me.Identity.String()
	return m
}

func (me *Context) SendMessage(msg string) error {
	m := me.NewMessage()
	m.Message = msg

	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = me.Conn.WriteMessage(websocket.TextMessage, buf)
	return err
}
