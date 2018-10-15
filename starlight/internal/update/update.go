package update

import (
	"encoding/json"
	"time"

	"github.com/interstellar/starlight/starlight/fsm"
)

type Type string

const (
	InitType      Type = "init"
	ConfigType    Type = "config"
	AccountType   Type = "account"
	ChannelType   Type = "channel"
	WarningType   Type = "warning"
	TxSuccessType Type = "tx_success"
	TxFailureType Type = "tx_failed"
)

type Update struct {
	// Type denotes what kind of state change this value represents.
	// If Type is Init or Config, field Config will be set.
	// along with one of the InputX fields.
	// If Type is Channel, field Channel will be set,
	// along with one of the InputX fields.
	// If Type is Warning, field Warning will be set.
	Type Type

	// UpdateNum is the number of this update.
	// Each update is assigned a number in a contiguous sequence:
	// 1, 2, 3, etc.
	UpdateNum uint64

	// Account describes the account ID this update affects and
	// the balance of the account. This field is set for every
	// update type.
	Account *Account

	// The following fields are all the type-specific payloads.
	// For example, a "config" update sets Config to describe
	// the configuration change it made.

	Config *Config `json:",omitempty"`

	// Channel describes the result of a channel state transition.
	// It is set when Type is channel.
	Channel *fsm.Channel

	// The InputX fields describe input events handled by the channel.
	// When Type is channel or account, one of these fields will be set,
	// to show how the channel arrived at its new state.
	InputCommand    *fsm.Command
	InputMessage    *fsm.Message
	InputTx         *fsm.Tx
	InputLedgerTime time.Time

	// UpdateLedgerTime is the ledger time of this update.
	// Not to be confused with InputLedgerTime.
	UpdateLedgerTime time.Time

	// In some cases when InputTx is set, this field might be set to show
	// which particular operation was responsible for the state change.
	OpIndex int

	Warning string

	// if this update included an outgoing transaction from the wallet account,
	// this is its sequence number (as a string, so JS can read it)
	PendingSequence string
}

type Account struct {
	Balance uint64
	ID      string
}

// Config has user-facing, primary options for the Starlight agent
// WARNING: this software is not compatible with Stellar mainnet.
type Config struct {
	Username   string `json:",omitempty"`
	Password   string `json:",omitempty"` // always "[redacted]" if set
	HorizonURL string `json:",omitempty"`
}

// Satisfy json.Marshaler. Required for genbolt.
func (u *Update) MarshalJSON() ([]byte, error) {
	type t Update
	return json.Marshal((*t)(u))
}

// Satisfy json.Unmarshaler. Required for genbolt.
func (u *Update) UnmarshalJSON(b []byte) error {
	type t Update
	return json.Unmarshal(b, (*t)(u))
}
