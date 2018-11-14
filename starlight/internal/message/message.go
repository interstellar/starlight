package message

import (
	"encoding/json"

	"github.com/interstellar/starlight/starlight/fsm"
)

// Message stores the messages sent by an agent
// per channel.
type Message struct {
	Messages   []*fsm.Message
	LastSeqNum uint64
}

// Add appends the latest sent message to the Message object,
// updating the latest sequence number
func (m *Message) Add(msg *fsm.Message, num *uint64) {
	m.LastSeqNum++
	*num = m.LastSeqNum
	m.Messages = append(m.Messages, msg)
}

// From returns all message sent from sequence number a, inclusive
// up until sequence number b, exclusive.
func (m *Message) From(a, b uint64) []*fsm.Message {
	msgs := make([]*fsm.Message, 0)
	for _, msg := range m.Messages {
		if a <= msg.MsgNum && msg.MsgNum < b {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

// MarshalJSON implements json.Marshaler. Required for genbolt.
func (m *Message) MarshalJSON() ([]byte, error) {
	type t Message
	return json.Marshal((*t)(m))
}

// UnmarshalJSON implements json.Unmarshaler. Required for genbolt.
func (m *Message) UnmarshalJSON(b []byte) error {
	type t Message
	return json.Unmarshal(b, (*t)(m))
}
