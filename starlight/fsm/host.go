package fsm

import (
	"encoding/json"

	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/worizon/xlm"
)

// Balance represents the point-in-time state of a
// non-XLM Asset in the WalletAcct.
type Balance struct {
	Asset   xdr.Asset
	Amount  int64
	Pending bool
}

// WalletAcct represents the point-in-time state of the
// channel's wallet account, passed to the FSM for state
// transitions that access or update host-level data.
type WalletAcct struct {
	NativeBalance xlm.Amount
	Seqnum        xdr.SequenceNumber
	Address       string // Stellar federation address
	Cursor        string
	Balances      []Balance
}

// MarshalJSON implements json.Marshaler. Required for genbolt.
func (w *WalletAcct) MarshalJSON() ([]byte, error) {
	type t WalletAcct
	return json.Marshal((*t)(w))
}

// UnmarshalJSON implements json.Unmarshaler. Required for genbolt.
func (w *WalletAcct) UnmarshalJSON(b []byte) error {
	type t WalletAcct
	return json.Unmarshal(b, (*t)(w))
}
