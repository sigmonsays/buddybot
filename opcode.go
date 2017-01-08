package buddybot

//go:generate stringer -type=OpCode

type OpCode int

const (
	InvalidOp OpCode = iota

	// client has connected
	RegisterOp

	// client has disconnected
	UnregisterOp

	// a message has been sent
	// this type of message is broadcasted to everyone
	// messages are dispatched to command handlers in buddy too
	MessageOp

	// a message has been sent
	// this type of message is broadcasted to everyone
	// this message is not disaptched to command handlers in buddy
	RawMessageOp

	// message directly to a single person
	DirectMessageOp

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

	// instruct client to disconnect
	DisconnectOp

	// do nothing but print a log line
	NoOp
)

var MaxOpCode = NoOp

// return a list of all opcodes
func OpCodes() []OpCode {
	opcodes := make([]OpCode, 0)
	for i := 0; i <= int(MaxOpCode); i++ {
		opcodes = append(opcodes, OpCode(i))
	}
	return opcodes
}
