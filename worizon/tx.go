package worizon

import (
	"time"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/xdr"
)

// Tx represents a Stellar transaction.
type Tx struct {
	Env    *xdr.TransactionEnvelope
	Result *xdr.TransactionResult

	// The following fields may not be available for failed transactions.

	PT         string // paging token a.k.a. cursor
	LedgerNum  int32
	LedgerTime time.Time
	SeqNum     string
}

// NewTx produces a Tx from a Horizon Transaction object.
func NewTx(htx *horizon.Transaction) (*Tx, error) {
	var env xdr.TransactionEnvelope
	err := xdr.SafeUnmarshalBase64(htx.EnvelopeXdr, &env)
	if err != nil {
		return nil, err
	}

	var result xdr.TransactionResult
	err = xdr.SafeUnmarshalBase64(htx.ResultXdr, &result)
	if err != nil {
		return nil, err
	}

	tx := &Tx{
		Env:        &env,
		Result:     &result,
		PT:         htx.PT,
		LedgerNum:  htx.Ledger,
		LedgerTime: htx.LedgerCloseTime,
		SeqNum:     htx.AccountSequence,
	}
	return tx, nil
}
