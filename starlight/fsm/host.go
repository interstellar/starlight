package fsm

import (
	"encoding/json"

	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/worizon/xlm"
)

// WalletAcct represents the point-in-time state of the
// channel's wallet account, passed to the FSM for state
// transitions that access or update host-level data.
type WalletAcct struct {
	Balance xlm.Amount
	Seqnum  xdr.SequenceNumber
	Address string // Stellar federation address
	Cursor  string
}

// Satisfy json.Marshaler. Required for genbolt.
func (w *WalletAcct) MarshalJSON() ([]byte, error) {
	type t WalletAcct
	return json.Marshal((*t)(w))
}

// Satisfy json.Unmarshaler. Required for genbolt.
func (w *WalletAcct) UnmarshalJSON(b []byte) error {
	type t WalletAcct
	return json.Unmarshal(b, (*t)(w))
}
