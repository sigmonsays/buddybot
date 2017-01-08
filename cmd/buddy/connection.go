package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

func NewConnection(ws *websocket.Conn) *Connection {
	return &Connection{ws: ws}
}

type Connection struct {
	ws *websocket.Conn
	mx sync.Mutex
}

func (c *Connection) WriteMessage(messageType int, data []byte) error {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.ws.WriteMessage(messageType, data)
}

func (c *Connection) ReadMessage() (int, []byte, error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.ws.ReadMessage()
}
