package buddybot

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Message struct {
	connection *Connection
	Id         int64  `json:"id"`
	Op         OpCode `json:"op"`
	From       string `json:"from,omitempty"`
	Message    string `json:"message,omitempty"`

	// uniquely identify the message
	Tag string `json:"tag,omitempty"`
}

// set a new tag on the message
func (m *Message) GenerateTag() {
	uu := uuid.New()
	hx := hex.EncodeToString(uu[:])
	m.Tag = hx
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
