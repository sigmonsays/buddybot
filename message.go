//go:generate stringer -type=OpCode

package buddybot

import (
	"encoding/json"
	"fmt"
)

type OpCode int

const (
	InvalidOp OpCode = iota

	// client has connected
	RegisterOp

	// client has disconnected
	UnregisterOp

	// a message has been sent
	MessageOp

	// a notice is a informational message. likely from the system
	NoticeOp
	// a user has joined
	JoinOp

	// a user has changed their nick name
	NickOp

	// a ping to keep the websocket connection alive
	PingOp

	// Client List op sends a list of connected clients to the connection
	ClientListOp
)

type Message struct {
	connection *Connection
	Id         int64  `json:"id"`
	Op         OpCode `json:"op"`
	From       string `json:"from,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (m *Message) String() string {
	return fmt.Sprintf("cid:%d op:%s/%d from:%q message:%q",
		m.Id, m.Op, m.Op, m.From, m.Message)
}
func (m *Message) Json() []byte {
	data, _ := json.Marshal(m)
	return data
}
func (m *Message) FromJson(data []byte) error {
	err := json.Unmarshal(data, m)
	return err
}
