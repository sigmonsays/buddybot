package buddybot

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func NewMessage() *Message {
	return &Message{
		Op: MessageOp,
	}
}

type Message struct {
	connection *Connection

	// connection id
	Id int64 `json:"id"`

	// type of message
	Op OpCode `json:"op"`

	// integer identifiers
	IdTo   int64
	IdFrom int64

	// to and from addresses
	To   string `json:"to"`
	From string `json:"from"`

	Message string `json:"message"`

	// uniquely identify the message
	Tag string `json:"tag,omitempty"`
}

func (m *Message) Copy() *Message {
	var m2 Message
	m2 = *m
	return &m2
}

// copies the message and flips the to/from
func (m *Message) Reply() *Message {
	m2 := m.Copy()
	to := m2.To
	m2.To = m.From
	m2.From = to
	return m2
}

func (m *Message) FromIdentity() *Identity {
	ident, _ := ParseIdentity(m.From)
	return ident
}

func (m *Message) ToIdentity() *Identity {
	ident, _ := ParseIdentity(m.To)
	return ident
}

func (m *Message) WithMessage(s string, args ...interface{}) *Message {
	m.Message = fmt.Sprintf(s, args...)
	return m
}

func (m *Message) WithTo(s string) *Message {
	m.To = s
	return m
}

func (m *Message) WithFrom(s string) *Message {
	m.From = s
	return m
}

// set a new tag on the message
func (m *Message) GenerateTag() {
	uu := uuid.New()
	hx := hex.EncodeToString(uu[:])
	m.Tag = hx
}

func (m *Message) String() string {
	return fmt.Sprintf("cid:%d op:%s/%d to:%q from:%q message:%q",
		m.Id, m.Op, m.Op, m.To, m.From, m.Message)
}
func (m *Message) Json() []byte {
	data, _ := json.Marshal(m)
	return data
}
func (m *Message) FromJson(data []byte) error {
	err := json.Unmarshal(data, m)
	return err
}

// convenience
func NewNotice(s string, args ...interface{}) *Message {
	m := NewMessage()
	m.Op = NoticeOp
	m.Message = fmt.Sprintf(s, args...)
	return m
}
func NewDirectMessage(s string, args ...interface{}) *Message {
	m := NewMessage()
	m.Op = DirectMessageOp
	m.Message = fmt.Sprintf(s, args...)
	return m
}
