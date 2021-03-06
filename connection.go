package buddybot

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewConnection(hub *Hub, id int64, ws *websocket.Conn) *Connection {
	c := &Connection{
		id:   id,
		hub:  hub,
		send: make(chan *Message, 256),
		ws:   ws,
	}
	return c
}

// connection is an middleman between the websocket connection and the hub.
type Connection struct {
	hub *Hub
	id  int64

	// the identity of connection
	Identity string

	// The websocket connection.
	ws *websocket.Conn

	mx sync.Mutex

	// Buffered channel of outbound messages.
	send chan *Message
}

func (c *Connection) GetId() int64 {
	return c.id
}

func (c *Connection) String() string {
	return fmt.Sprintf("cid:%d identity:%s ip:%s", c.id, c.Identity, c.ws.RemoteAddr())
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Connection) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		c.hub.broadcast <- &message{
			connection: c,
			data:       msg,
		}
	}
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Connection) Start() {
	go c.readPump()
	go c.writePump()
}

func (c *Connection) WriteMessage(mt int, payload []byte) error {
	return c.write(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message.Json()); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
