// Package worizon is a wrapper for horizon.
// It exposes very little of the functionality
// of the underlying library — just enough for Interstellar.
package worizon

import (
	"context"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/network"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net"
)

var (
	errUninitialized = errors.New("uninitialized")
	errMainnet       = errors.New("using mainnet instead of testnet")
)

// Alias some types that don't need to be wrapped.
type (
	Ledger             = horizon.Ledger
	LedgerHandler      = horizon.LedgerHandler
	Transaction        = horizon.Transaction
	Cursor             = horizon.Cursor
	TxSuccess          = horizon.TransactionSuccess
	Account            = horizon.Account
	TransactionHandler = horizon.TransactionHandler
)

// horizonClient is a minimal subset of horizon.ClientInterface that allows this for simpler test doubles
// It is expected that this interface will expand as horizon implements more of the horizon Client spec
type horizonClient interface {
	Root() (horizon.Root, error)

	LoadAccount(accountID string) (Account, error)
	SequenceForAccount(accountID string) (xdr.SequenceNumber, error)
	StreamLedgers(ctx context.Context, cursor *Cursor, handler LedgerHandler) error
	StreamTransactions(ctx context.Context, accountID string, cursor *Cursor, handler TransactionHandler) error
	SubmitTransaction(txeBase64 string) (TxSuccess, error)
}

// Client is a wrapper for some of a horizon client's functionality.
// To initialize a Client, call SetURL on the zero value.
// It is okay to call methods on Client concurrently.
// A Client must not be copied after first use.
type Client struct {
	mu      sync.Mutex
	changed chan struct{}
	hclient horizonClient
	now     time.Time // updated at each ledger close
	timers  []*timer
	http    horizon.HTTP

	// initHorizon indicates whether the Client was initialized with a
	// horizon client, in which case SetURL has no effect.
	initHorizon bool

	startClockOnce sync.Once
}

type timer struct {
	t time.Time
	f func()
}

func NewClient(rt http.RoundTripper, horizon horizonClient) *Client {
	return &Client{
		http: &http.Client{
			Transport: rt,
		},
		hclient:     horizon,
		initHorizon: horizon != nil,
	}
}

// SetURL sets the URL for c to url.
//
// If a non-nil horizon client is provided in NewClient,
// SetURL has no effect.
// Otherwise, SetURL must be called before any other method on c.
// After that, it is safe to call SetURL concurrently
// with other methods on Client.
func (c *Client) SetURL(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initHorizon {
		return
	}
	if c.http == nil {
		c.http = new(http.Client)
	}
	c.hclient = &horizon.Client{
		URL:  strings.TrimRight(url, "/"),
		HTTP: c.http,
	}
	changed := c.changed
	c.changed = make(chan struct{})

	if changed != nil {
		close(changed)
	}
}

func (c *Client) getHorizonClient(url string) *horizon.Client {
	c.mu.Lock()
	if c.http == nil {
		c.http = new(http.Client)
	}
	http := c.http
	c.mu.Unlock()
	horizonClient := &horizon.Client{
		URL:  url,
		HTTP: http,
	}
	return horizonClient
}

// ValidateURL tests that the URL is a valid horizon URL
// by opening a test connection.
func (c *Client) ValidateURL(url string) error {
	horizonClient := c.getHorizonClient(url)
	_, err := horizonClient.Root()
	return err
}

// ValidateTestnetURL tests that the URL is a valid horizon URL
// on the Stellar testnet by opening a test connection.
func (c *Client) ValidateTestnetURL(url string) error {
	horizonClient := c.getHorizonClient(url)
	root, err := horizonClient.Root()
	if err != nil {
		return err
	}
	if root.NetworkPassphrase != network.TestNetworkPassphrase {
		return errMainnet
	}
	return nil
}

// Now returns the time of the last seen ledger.
func (c *Client) Now() time.Time {
	c.startClockOnce.Do(c.startClock)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// AfterFunc waits for a Stellar ledger at or after time t
// to commit, and then calls f in its own goroutine.
func (c *Client) AfterFunc(t time.Time, f func()) {
	c.startClockOnce.Do(c.startClock)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timers = append(c.timers, &timer{t, f})
	sort.Slice(c.timers, func(i, j int) bool {
		return c.timers[i].t.Before(c.timers[j].t)
	})
}

func (c *Client) startClock() {
	ready := make(chan struct{})
	now := Cursor("now")
	go func() {
		ready := ready
		ctx := context.Background()
		err := c.streamLedgers(ctx, &now, func(l Ledger) {
			c.mu.Lock()
			defer c.mu.Unlock()
			if l.ClosedAt.Before(c.now) {
				return // don't let the timestamp go backward
			}

			c.now = l.ClosedAt

			for len(c.timers) > 0 && c.now.After(c.timers[0].t) {
				f := c.timers[0].f
				c.timers = c.timers[1:]
				go f()
			}

			if ready != nil {
				close(ready)
				ready = nil
			}
			return
		})
		if err != nil {
			panic(err)
		}
	}()
	<-ready
}

// StreamTxs reads from the ledger
// all transactions that affect account accountID,
// beginning at cur,
// and calls h for each one.
// If h returns a non-nil error,
// StreamTxs returns it.
// If the underlying call to StreamTransactions
// returns an error, StreamTxs will retry.
func (c *Client) StreamTxs(ctx context.Context, accountID string, cur Cursor, h func(Transaction) error) error {
	c.mu.Lock()
	hclient := c.hclient
	c.mu.Unlock()

	if hclient == nil {
		return errUninitialized
	}

	return c.streamHorizon(ctx, &cur, func(ctx context.Context, cur *Cursor, backoff *net.Backoff) error {
		ctx, cancel := context.WithCancel(ctx)
		return hclient.StreamTransactions(ctx, accountID, cur, func(tx Transaction) {
			backoff = &net.Backoff{Base: backoff.Base}
			handlerErr := h(tx)
			if handlerErr != nil {
				cancel()
			}
			*cur = Cursor(tx.PT)
		})
	})
}

func (c *Client) streamLedgers(ctx context.Context, cur *Cursor, h func(l Ledger)) error {
	c.mu.Lock()
	hclient := c.hclient
	c.mu.Unlock()

	if hclient == nil {
		return errUninitialized
	}

	return c.streamHorizon(ctx, cur, func(ctx context.Context, cur *Cursor, backoff *net.Backoff) error {
		return hclient.StreamLedgers(ctx, cur, func(l Ledger) {
			backoff = &net.Backoff{Base: backoff.Base}
			h(l)
			*cur = Cursor(l.PT)
		})
	})
}

func (c *Client) streamHorizon(ctx context.Context, cur *Cursor, s func(context.Context, *Cursor, *net.Backoff) error) error {
	origCtx := ctx

	// The base amount of time to wait between retries of the streaming callback s.
	// Wait time grows exponentially with each failure.
	// The streaming callback should reset the wait time
	// to this value each time it is called.
	const baseBackoff = 100 * time.Millisecond
	backoff := &net.Backoff{Base: baseBackoff}

	for {
		c.mu.Lock()
		hclient := c.hclient
		changed := c.changed
		c.mu.Unlock()
		if hclient == nil {
			return errUninitialized
		}
		ctx, cancel := context.WithCancel(ctx)
		go func() {
			select {
			case <-changed:
				cancel()
			case <-ctx.Done():
			}
		}()

		streamErr := s(ctx, cur, backoff)

		if origCtx.Err() == nil {
			if ctx.Err() != nil || streamErr != nil {
				dur := backoff.Next()
				log.Printf("received error %s streaming from horizon, retrying in %s", streamErr, dur)
				time.Sleep(dur)
				continue
			}
		}

		// TODO(kr): consider guaranteeing that we return
		// the error from origCtx. Streams currently
		// does that but doesn't guarantee it.
		return streamErr
	}
}

// SequenceForAccount implements SequenceProvider
// from package github.com/stellar/go/build.
func (c *Client) SequenceForAccount(accountID string) (xdr.SequenceNumber, error) {
	c.mu.Lock()
	hclient := c.hclient
	c.mu.Unlock()
	if hclient == nil {
		return 0, errUninitialized
	}
	return hclient.SequenceForAccount(accountID)
}

// SubmitTx submits a transaction to the network.
// The returned error can be (but is not necessarily)
// an instance of horizon.Error.
func (c *Client) SubmitTx(envXdr string) (response TxSuccess, err error) {
	c.mu.Lock()
	hclient := c.hclient
	c.mu.Unlock()
	if hclient == nil {
		return TxSuccess{}, errUninitialized
	}
	return hclient.SubmitTransaction(envXdr)
}

func (c *Client) LoadAccount(id string) (Account, error) {
	c.mu.Lock()
	hclient := c.hclient
	c.mu.Unlock()
	if hclient == nil {
		return Account{}, errUninitialized
	}
	return hclient.LoadAccount(id)
}
