package buddybot

import (
	"fmt"

	gologging "github.com/sigmonsays/go-logging"
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
	verbose     bool
	connections map[*Connection]bool
	broadcast   chan *message
	register    chan *Connection
	unregister  chan *Connection

	callbacks map[OpCode][]CallbackFn
	log       gologging.Logger
}

func NewHub() *Hub {
	callbacks := make(map[OpCode][]CallbackFn, 0)
	l := gologging.NewStd2Logger3("warn", "buddybot.hub")
	h := &Hub{
		broadcast:   make(chan *message, 50),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]bool),
		callbacks:   callbacks,
		log:         l,
	}

	return h
}

func (h *Hub) SetLogger(l gologging.Logger) {
	h.log = l
}

func (h *Hub) SetVerbose(verbose bool) {
	h.verbose = verbose
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
	if m.Id == 0 {
		m.Id = c.id
	}

	// fill in the destination address w/ the connection we're sending to
	if m.To == "" {
		m.To = c.Identity
	}

	// they are allowed to specify the identity with JoinOp
	/*
		if m.Op != JoinOp {
			m.From = h.getIdentity(m.Id)
		}
	*/

}

func (h *Hub) getIdentity(id int64) string {
	c, err := h.findConnection(id)
	if err != nil {
		return fmt.Sprintf("no-id=%d", id)
	}
	return c.Identity
}

func (h *Hub) SendTo(c *Connection, m *Message) error {
	h.setMessageIdentity(c, m)
	h.log.Tracef("SendTo to=(%s) msg=%s", c, m)
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
		h.log.Tracef("send client(%s) msg(%+v)", c, m)
		select {
		case c.send <- m:
		default:
			err++
			close(c.send)
			delete(h.connections, c)
		}
	}
	h.log.Tracef("SendBroadcast num-clients=%d (failures %d)", cnt, err)
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
		Op:      ClientListOp,
		Message: string(message),
	}

	h.log.Debugf("sent client list (%d clients) to connection id %d", len(ls.List), id)

	return h.SendTo(destination, msg)
}

func (h *Hub) FindConnection(id int64) (*Connection, error) {
	return h.findConnection(id)
}

func (h *Hub) findConnection(id int64) (*Connection, error) {
	for c := range h.connections {
		if c.id == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found cid:%d", id)
}

func (h *Hub) dispatch(op OpCode, c *Connection, m *Message) error {
	callbacks, ok := h.callbacks[op]
	if ok == false {
		return nil
	}
	h.log.Tracef("dispatch %s/%d callbacks:%d", op, op, len(callbacks))

	var err error
	for _, callback := range callbacks {
		err = callback(op, h, c, m)
		if err != nil {
			h.log.Warnf("callback op=%s/%d: %s", op, op, err)
		}
	}

	return nil
}

func (h *Hub) Register(c *Connection) {
	h.register <- c
}

func (h *Hub) Start() {

	if h.verbose {
		h.log.SetLevel("TRACE")
	}

	h.log.Debugf("start")
	for {
		select {
		case c := <-h.register:
			h.log.Infof("register connection cid:%d remote:%s", c.id, c.ws.RemoteAddr())
			h.connections[c] = true
			h.dispatch(RegisterOp, c, nil)

		case c := <-h.unregister:
			h.log.Infof("unregister connection cid:%d remote:%s", c.id, c.ws.RemoteAddr())
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
				h.log.Warnf(" FromJson [ %s ]: %s", data, err)
				continue
			}
			h.setMessageIdentity(data.connection, m)

			if h.verbose {
				h.log.Tracef("dispatch cid=%d op=%s/%d ip=%s data=%s",
					data.connection.id, m.Op, m.Op, data.connection.ws.RemoteAddr(), string(data.data))
			}

			err = h.dispatch(m.Op, data.connection, m)
			if err != nil {
				h.log.Warnf("dispatch op=%s/%d: %s", m.Op, m.Op, err)
			}
		}
	}
}
