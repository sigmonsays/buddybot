package main

import (
	"github.com/gorilla/websocket"
)

type Context struct {
	Conn *websocket.Conn
}

func (me *Context) SendMessage(msg []byte) error {
	err := me.Conn.WriteMessage(websocket.TextMessage, msg)
	return err
}
