package buddybot

import (
	"encoding/json"
	"fmt"
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
