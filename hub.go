package buddybot

import (
	"fmt"
	"sync"
)

// a low level message
type message struct {
	data       []byte
	connection *Connection
}

func (m *message) String() string {
	return fmt.Sprintf("<conn:%d data:%d>", m.connection.id, len(m.data))
}

type CallbackFn func(op OpCode, hub *Hub, c *Connection, m *Message) error

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	connections map[*Connection]bool
	broadcast   chan *message
	register    chan *Connection
	unregister  chan *Connection
	mx          sync.Mutex

	callbacks map[OpCode][]CallbackFn
}

func NewHub() *Hub {
	callbacks := make(map[OpCode][]CallbackFn, 0)
	h := &Hub{
		broadcast:   make(chan *message, 50),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]bool),
		callbacks:   callbacks,
	}
	return h
}

func (h *Hub) OnCallback(callback OpCode, fn CallbackFn) {
	ls, ok := h.callbacks[callback]
	if ok == false {
		ls = make([]CallbackFn, 0)
	}
	ls = append(ls, fn)
	h.callbacks[callback] = ls
}

func (h *Hub) setMessageIdentity(c *Connection, m *Message) {
	m.Id = c.id
	m.From = h.getIdentity(c)
}

func (h *Hub) SendTo(c *Connection, m *Message) error {
	log.Tracef("SendTo conn %s", c)
	h.setMessageIdentity(c, m)
	select {
	case c.send <- m:
	default:
		close(c.send)
		delete(h.connections, c)
	}
	return nil
}

func (h *Hub) Send(op OpCode, msg string) {
	m := &Message{Op: op, Message: msg}
	h.SendBroadcast(m)
}

func (h *Hub) SendBroadcast(m *Message) {
	cnt := 0
	err := 0
	for c := range h.connections {
		cnt++
		h.setMessageIdentity(c, m)
		select {
		case c.send <- m:
		default:
			err++
			close(c.send)
			delete(h.connections, c)
		}
	}
	log.Tracef("SendBroadcast num-clients=%d (failures %d)", cnt, err)
}

func (h *Hub) SendClientList(id int64) error {
	destination, err := h.findConnection(id)
	if err != nil {
		return err
	}

	ls := NewClientList()
	for c := range h.connections {
		ls.AddClient(c)
	}

	message := ls.ToJson()

	msg := &Message{
		Op:      MessageOp,
		Message: string(message),
	}

	log.Debugf("sent client list (%d clients) to connection id %d", len(ls.List), id)

	return h.SendTo(destination, msg)
}

func (h *Hub) findConnection(id int64) (*Connection, error) {
	for c := range h.connections {
		if c.id == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found cid:%d", id)
}

func (h *Hub) getIdentity(c *Connection) string {
	return c.Identity
}

func (h *Hub) dispatch(op OpCode, c *Connection, m *Message) error {
	callbacks, ok := h.callbacks[op]
	if ok == false {
		return nil
	}
	log.Tracef("dispatch %s: callbacks:%d", op, len(callbacks))

	var err error
	for _, callback := range callbacks {
		err = callback(op, h, c, m)
		if err != nil {
			log.Warnf("callback op=%s/%d: %s", op, op, err)
		}
	}

	return nil
}

func (h *Hub) Start() {
	log.Debugf("start")
	for {
		select {
		case c := <-h.register:
			log.Infof("register connection cid:%d remote:%s", c.id, c.ws.RemoteAddr())
			h.connections[c] = true

			h.dispatch(RegisterOp, c, nil)

		case c := <-h.unregister:
			log.Infof("unregister connection cid:%d remote:%s", c.id, c.ws.RemoteAddr())
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
				h.dispatch(UnregisterOp, c, nil)
			}

		case data := <-h.broadcast:
			m := &Message{
				connection: data.connection,
				Id:         data.connection.id,
			}

			err := m.FromJson(data.data)
			if err != nil {
				log.Infof("ERROR: FromJson [ %s ]: %s", data, err)
				continue
			}
			h.setMessageIdentity(data.connection, m)

			log.Tracef("dispatch cid=%d op=%s/%d ip=%s data=%s",
				data.connection.id, m.Op, m.Op, data.connection.ws.RemoteAddr(), string(data.data))

			err = h.dispatch(m.Op, data.connection, m)
			if err != nil {
				log.Warnf("dispatch op=%s/%d: %s", m.Op, m.Op, err)
			}
		}
	}
}
