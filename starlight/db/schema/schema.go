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
	_ json.Marshaler = (*update.Update)(nil)

	_ encoding.BinaryMarshaler = (*fsm.AccountId)(nil)
)

type Root struct {
	Agent *Agent
}

type Agent struct {
	Config  *Config
	Updates []*update.Update

	// Channels holds the state of all open channels. Closed channels
	// are deleted. (Their history is still available in Updates.)
	Channels map[string]*fsm.Channel

	EncryptedSeed    []byte
	NextKeypathIndex uint32
	PrimaryAcct      *fsm.AccountId
	Wallet           *fsm.WalletAcct
}

type Config struct {
	HorizonURL string
	Username   string

	// PwType records which hashing function was used for PwHash.
	// Currently, it's always "bcrypt".
	PwType string
	PwHash []byte

	MaxRoundDurMin   int64
	FinalityDelayMin int64
	ChannelFeerate   int64
	HostFeerate      int64

	KeepAlive bool
}
