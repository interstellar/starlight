package db

import (
	"encoding"
	"encoding/json"

	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/internal/update"
)

var (
	_ json.Marshaler = (*fsm.Channel)(nil)
	_ json.Marshaler = (*fsm.WalletAcct)(nil)
	_ json.Marshaler = (*fsm.Message)(nil)
	_ json.Marshaler = (*update.Update)(nil)

	_ encoding.BinaryMarshaler = (*fsm.AccountID)(nil)
)

// Root is the type of the root bucket, as required by genbolt.
type Root struct {
	Agent *Agent
}

// Agent is the db layout for a Starlight agent.
type Agent struct {
	Config  *Config
	Updates []*update.Update

	// Ready indicates whether or not the Agent is ready to accept
	// and process new commands. The Agent is only in a not-ready
	// state when it is closing.
	Ready bool

	// Channels holds the state of all open channels. Closed channels
	// are deleted. (Their history is still available in Updates.)
	Channels map[string]*fsm.Channel

	// Messages persists all of the channel messages.
	Messages map[string]*Message

	EncryptedSeed    []byte
	NextKeypathIndex uint32
	PrimaryAcct      *fsm.AccountID
	Wallet           *fsm.WalletAcct
}

// Message represents all of the messages that an agent has sent
// for a given channel.
type Message struct {
	Messages []*fsm.Message
}

// Config is the db layout for Starlight agent-level configuration.
type Config struct {
	HorizonURL string
	Username   string

	// PwType records which hashing function was used for PwHash.
	// Currently, it's always "bcrypt".
	PwType string
	PwHash []byte

	MaxRoundDurMins   int64
	FinalityDelayMins int64
	ChannelFeerate    int64
	HostFeerate       int64

	KeepAlive bool
}
