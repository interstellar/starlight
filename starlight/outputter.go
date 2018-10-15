package starlight

import (
	"time"

	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/starlight/fsm"
)

// Satisfies fsm.Outputter.
type outputter struct {
	timer time.Time // zero if no timer
	msgs  []*fsm.Message
	txs   []xdr.TransactionEnvelope
}

// OutputMsg queues m to to be sent to the peer.
func (o *outputter) OutputMsg(m *fsm.Message) {
	o.msgs = append(o.msgs, m)
}

// OutputTx queues xdrEnv to to be submitted to the ledger.
func (o *outputter) OutputTx(xdrEnv xdr.TransactionEnvelope) {
	o.txs = append(o.txs, xdrEnv)
}

// SetTimer sets a timer for the channel to fire
// as soon as possible after ledger time t.
// If t is the zero time, it doesn't set a timer.
// Successive calls to SetTimer clear the previous timer in o, if any.
func (o *outputter) SetTimer(t time.Time) {
	o.timer = t
}
