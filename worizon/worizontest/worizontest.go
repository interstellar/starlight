package worizontest

import (
	"context"
	"sync"
	"time"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/xdr"
)

type FakeHorizonClient struct {
	mu                   sync.Mutex
	transactionEnvelopes []string
}

func (c *FakeHorizonClient) Root() (horizon.Root, error) {
	return horizon.Root{}, nil
}

func (c *FakeHorizonClient) SubmitTransaction(txeBase64 string) (horizon.TransactionSuccess, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transactionEnvelopes = append(c.transactionEnvelopes, txeBase64)
	return horizon.TransactionSuccess{}, nil
}

func (c *FakeHorizonClient) StreamLedgers(ctx context.Context, cursor *horizon.Cursor, handler horizon.LedgerHandler) error {
	ledger := horizon.Ledger{
		ClosedAt: time.Now(),
	}
	handler(ledger)
	return nil
}

// Not Implemented

func (c *FakeHorizonClient) LoadAccount(accountID string) (horizon.Account, error) {
	return horizon.Account{}, nil
}

func (c *FakeHorizonClient) SequenceForAccount(accountID string) (xdr.SequenceNumber, error) {
	return xdr.SequenceNumber(0), nil
}

func (c *FakeHorizonClient) StreamTransactions(ctx context.Context, accountID string, cursor *horizon.Cursor, handler horizon.TransactionHandler) error {
	return nil
}
