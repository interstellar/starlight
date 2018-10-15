// Package worizon is a wrapper for horizon.
// It exposes very little of the functionality
// of the underlying library — just enough for Chain.
package worizon

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/stellar/go/network"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net"

	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/support/render/hal"
	"github.com/stellar/go/xdr"
)

var (
	errUninitialized          = errors.New("uninitialized")
	errTxSignerNotInitialized = errors.New("Required field txSigner is not set")
	errMainnet                = errors.New("using mainnet instead of testnet")
)

// Alias some types that don't need to be wrapped.
type (
	Ledger             = horizon.Ledger
	LedgerHandler      = horizon.LedgerHandler
	Tx                 = horizon.Transaction
	Cursor             = horizon.Cursor
	TxSuccess          = horizon.TransactionSuccess
	Account            = horizon.Account
	TransactionHandler = horizon.TransactionHandler
	Error              = horizon.Error
	PaymentHandler     = horizon.PaymentHandler
)

// horizonClient is a minimal subset of horizon.ClientInterface that allows this for simpler test doubles
// It is expected that this interface will expand as horizon implements more of the horizon Client spec
type horizonClient interface {
	Root() (horizon.Root, error)

	LoadAccount(accountID string) (Account, error)
	SequenceForAccount(accountID string) (xdr.SequenceNumber, error)
	StreamLedgers(ctx context.Context, cursor *Cursor, handler LedgerHandler) error
	StreamTransactions(ctx context.Context, accountID string, cursor *Cursor, handler TransactionHandler) error
	StreamPayments(ctx context.Context, accountID string, cursor *Cursor, handler PaymentHandler) error
	LoadMemo(p *horizon.Payment) error
	SubmitTransaction(txeBase64 string) (TxSuccess, error)
}

type PaymentsPage struct {
	Links    hal.Links `json:"_links"`
	Embedded struct {
		Records []horizon.Payment `json:"records"`
	} `json:"_embedded"`
}

type Payment struct {
	Credit
	DestAddr string
}

type Credit struct {
	Asset
	Amount string
}

type Asset struct {
	Issuer string
	Code   string
}

const defaultLimit = 10

type TXSigner interface {
	// SignTransaction will return a build.TransactionEnvelopeBuilder that
	// is signed by all private keys that correspond to the public keys
	// presented as arguments or it will return an error
	SignTransaction(ctx context.Context, tx *build.TransactionBuilder, pubKeys ...string) (build.TransactionEnvelopeBuilder, error)

	// PersistSeed will encrypt and call persistenceProvider with the result of the encryption
	PersistSeed(ctx context.Context, seed string) (string, error)
}

// Client is a wrapper for some of a horizon client's functionality.
// To initialize a Client, call SetURL on the zero value.
// It is okay to call methods on Client concurrently.
// A Client must not be copied after first use.
type Client struct {
	mu       sync.Mutex
	changed  chan struct{}
	hclient  horizonClient
	now      time.Time // updated at each ledger close
	timers   []*timer
	http     horizon.HTTP
	txSigner TXSigner

	// url of the horizon server in case we want to send requests to directly
	// to horizon api endpoints
	url string

	// initHorizon indicates whether the Client was initialized with a
	// horizon client, in which case SetURL has no effect.
	initHorizon bool

	startClockOnce sync.Once
}

type timer struct {
	t time.Time
	f func()
}

func NewClient(rt http.RoundTripper, horizon horizonClient, txSigner TXSigner) *Client {
	return &Client{
		http: &http.Client{
			Transport: rt,
		},
		hclient:     horizon,
		initHorizon: horizon != nil,
		txSigner:    txSigner,
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
	c.url = strings.TrimRight(url, "/")
	c.hclient = &horizon.Client{
		URL:  c.url,
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
func (c *Client) StreamTxs(ctx context.Context, accountID string, cur Cursor, h func(Tx) error) error {
	c.mu.Lock()
	hclient := c.hclient
	c.mu.Unlock()

	if hclient == nil {
		return errUninitialized
	}

	return c.streamHorizon(ctx, &cur, func(ctx context.Context, cur *Cursor, backoff *net.Backoff) error {
		ctx, cancel := context.WithCancel(ctx)
		return hclient.StreamTransactions(ctx, accountID, cur, func(tx Tx) {
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

func (c *Client) buildCreateAccount(ctx context.Context, baseAccountPubKey, channelAccountPubKey, startingBalance string, mainnet bool, newAccountPubKeys ...string) (string, error) {
	c.mu.Lock()
	hclient := c.hclient
	txSigner := c.txSigner
	c.mu.Unlock()

	if hclient == nil {
		return "", errUninitialized
	}
	if txSigner == nil {
		return "", errTxSignerNotInitialized
	}

	amount, err := strconv.ParseFloat(startingBalance, 64)
	if err != nil {
		return "", errors.New("invalid balance string")
	}
	if amount < 1 {
		return "", errors.New("starting balance too low")
	}

	if len(newAccountPubKeys) > 100 {
		return "", errors.New("too many accounts")
	}

	txMutators := make([]build.TransactionMutator, 0, 3+len(newAccountPubKeys))
	if channelAccountPubKey != "" {
		txMutators = append(txMutators, build.SourceAccount{channelAccountPubKey})
	} else {
		txMutators = append(txMutators, build.SourceAccount{baseAccountPubKey})
	}
	txMutators = append(txMutators, build.AutoSequence{hclient})
	if mainnet {
		txMutators = append(txMutators, build.PublicNetwork)
	} else {
		txMutators = append(txMutators, build.TestNetwork)
	}

	var muts []interface{}
	for _, newPubKey := range newAccountPubKeys {
		if channelAccountPubKey != "" {
			muts = append(muts, build.SourceAccount{baseAccountPubKey})
		}

		muts = append(muts, build.Destination{newPubKey}, build.NativeAmount{startingBalance})
		txMutators = append(txMutators, build.CreateAccount(muts...))
		muts = nil
	}

	tx, err := build.Transaction(txMutators...)
	if err != nil {
		return "", errors.Wrap(err, "cannot buildCreateAccount")
	}
	pubKeys := []string{baseAccountPubKey}
	if channelAccountPubKey != "" {
		pubKeys = append(pubKeys, channelAccountPubKey)
	}
	txe, err := txSigner.SignTransaction(ctx, tx, pubKeys...)
	if err != nil {
		return "", errors.Wrap(err, "cannot sign createAccount TX")
	}
	return txe.Base64()
}

// CreateAccounts will create N new accounts specified by newNewAccounts and persist their seeds with txSigner.
// If we encounter any errors while persisting seeds, the function will return an error. However, there might be
// a number of seeds that have been persisted.
// Note that we can create a maximum of 100 accounts at a time.
func (c *Client) CreateAccounts(ctx context.Context, baseAccountPubKey, channelAccountPubKey, startingBalance string, mainnet bool, numNewAccounts int) (TxSuccess, []string, error) {
	c.mu.Lock()
	txSigner := c.txSigner
	c.mu.Unlock()

	if txSigner == nil {
		return TxSuccess{}, nil, errTxSignerNotInitialized
	}

	newAccountPubKeys := make([]string, 0, numNewAccounts)
	seeds := make([]string, 0, numNewAccounts)
	for i := 0; i < numNewAccounts; i++ {
		kp, err := keypair.Random()
		if err != nil {
			return TxSuccess{}, nil, errors.Wrap(err, "generating a new keypair")
		}
		newAccountPubKeys = append(newAccountPubKeys, kp.Address())
		seeds = append(seeds, kp.Seed())
	}

	e, err := c.buildCreateAccount(ctx, baseAccountPubKey, channelAccountPubKey, startingBalance, mainnet, newAccountPubKeys...)
	if err != nil {
		return TxSuccess{}, nil, errors.Wrap(err, "building create account transaction")
	}

	resp, err := c.SubmitTx(e)
	if err != nil {
		return TxSuccess{}, nil, errors.Wrap(err, "submitting create account transaction")
	}

	for _, seed := range seeds {
		_, err = txSigner.PersistSeed(ctx, seed)
		if err != nil {
			return TxSuccess{}, nil, errors.Wrap(err, "Cannot persist encrypted seed to datastore")
		}
	}

	return resp, newAccountPubKeys, nil
}

func (c *Client) BuildCreditAccount(ctx context.Context, baseAccountPubKey, channelAccountPubKey string, mainnet bool, memo string, payments ...Payment) (string, string, error) {
	c.mu.Lock()
	hclient := c.hclient
	txSigner := c.txSigner
	c.mu.Unlock()

	if hclient == nil {
		return "", "", errUninitialized
	}
	if txSigner == nil {
		return "", "", errTxSignerNotInitialized
	}

	if len(payments) == 0 {
		return "", "", errors.New("Nothing to send")
	}
	txMutators := make([]build.TransactionMutator, 0, len(payments)+4)
	if channelAccountPubKey != "" {
		txMutators = append(txMutators, build.SourceAccount{channelAccountPubKey})
	} else {
		txMutators = append(txMutators, build.SourceAccount{baseAccountPubKey})
	}
	txMutators = append(txMutators, build.AutoSequence{hclient})
	if mainnet {
		txMutators = append(txMutators, build.PublicNetwork)
	} else {
		txMutators = append(txMutators, build.TestNetwork)
	}
	txMutators = append(txMutators, build.MemoText{memo})

	var muts []interface{}
	for _, p := range payments {
		if channelAccountPubKey != "" {
			muts = append(muts, build.SourceAccount{baseAccountPubKey})
		}

		if p.Code == "" {
			muts = append(muts, build.Destination{p.DestAddr}, build.NativeAmount{p.Amount})
		} else {
			muts = append(muts, build.Destination{p.DestAddr}, build.CreditAmount{
				Code:   p.Code,
				Issuer: p.Issuer,
				Amount: p.Amount,
			})
		}
		txMutators = append(txMutators, build.Payment(muts...))
		muts = nil
	}

	tx, err := build.Transaction(txMutators...)
	if err != nil {
		return "", "", err
	}
	hash, err := tx.HashHex()
	if err != nil {
		return "", "", err
	}

	pubKeys := []string{baseAccountPubKey}
	if channelAccountPubKey != "" {
		pubKeys = append(pubKeys, channelAccountPubKey)
	}
	txe, err := txSigner.SignTransaction(ctx, tx, pubKeys...)
	if err != nil {
		return "", "", err
	}

	txeBase64, err := txe.Base64()
	return hash, txeBase64, err
}

func (c *Client) CreditAccount(ctx context.Context, baseAccountPubKey, channelAccountPubKey string, mainnet bool, payments ...Payment) (TxSuccess, error) {
	_, txeBase64, err := c.BuildCreditAccount(ctx, baseAccountPubKey, channelAccountPubKey, mainnet, "", payments...)
	if err != nil {
		return TxSuccess{}, err
	}

	return c.SubmitTx(txeBase64)
}

func (c *Client) CreateTrustline(ctx context.Context, srcAddr string, mainnet bool, assets ...Asset) (TxSuccess, error) {
	c.mu.Lock()
	hclient := c.hclient
	txSigner := c.txSigner
	c.mu.Unlock()

	if hclient == nil {
		return TxSuccess{}, errUninitialized
	}
	if txSigner == nil {
		return TxSuccess{}, errTxSignerNotInitialized
	}

	txMutators := make([]build.TransactionMutator, 0, len(assets)+3)
	txMutators = append(txMutators, build.SourceAccount{srcAddr})
	txMutators = append(txMutators, build.AutoSequence{hclient})
	if mainnet {
		txMutators = append(txMutators, build.PublicNetwork)
	} else {
		txMutators = append(txMutators, build.TestNetwork)
	}
	for _, a := range assets {
		txMutators = append(txMutators, build.Trust(
			a.Code,
			a.Issuer,
		))
	}

	tx, err := build.Transaction(txMutators...)
	if err != nil {
		return TxSuccess{}, errors.Wrap(err, "cannot build transaction for trustline creation")
	}
	txe, err := txSigner.SignTransaction(ctx, tx, srcAddr)
	if err != nil {
		return TxSuccess{}, errors.Wrap(err, "cannot sign transaction for trustline creation")
	}
	txeBase64, err := txe.Base64()
	if err != nil {
		return TxSuccess{}, errors.Wrap(err, "cannot base64 encode transaction for trustline creation")
	}

	return c.SubmitTx(txeBase64)
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

// StreamPayments will stream payments of the account starting from cur until it sees a payment whose memo value
// equals to stopAtMemoText. It will stream indefinitely if stopAtMemoText is nil.
func (c *Client) StreamPayments(ctx context.Context, accountID string, cur Cursor, stopAtMemoText *string, h PaymentHandler) error {
	c.mu.Lock()
	hclient := c.hclient
	c.mu.Unlock()

	if hclient == nil {
		return errUninitialized
	}

	origCtx := ctx

	// The base amount of time to wait between retries of StreamPayments.
	// Wait time grows exponentially with each failure.
	// Each time the StreamPayments callback gets called,
	// the wait time resets to this value.
	const baseBackoff = 100 * time.Millisecond
	backoff := &net.Backoff{Base: baseBackoff}

	for {
		ctx, cancel := context.WithCancel(ctx)
		var memoErr error
		streamErr := hclient.StreamPayments(ctx, accountID, &cur, func(payment horizon.Payment) {
			// Reset retry wait time.
			backoff = &net.Backoff{Base: baseBackoff}
			if h != nil {
				h(payment)
			}
			if stopAtMemoText != nil {
				memoErr = hclient.LoadMemo(&payment)
				if memoErr != nil {
					cancel()
				}
				if payment.Memo.Type == "text" && payment.Memo.Value == *stopAtMemoText {
					cancel()
				}
			}
			cur = Cursor(payment.PagingToken)
		})
		if origCtx.Err() == nil {
			if ctx.Err() == context.Canceled {
				return memoErr
			}
			if streamErr != nil {
				time.Sleep(backoff.Next())
				continue
			}
		}

		return streamErr
	}
}

// PaymentsForAccount returns a page of payments associated with the provided accountID.
// asc: asc = true means older payments first; asc = false means newer payments first.
// cursor: a payment paging token specifying from where to begin results.
// limit: the count of records at most to return.
// TODO: replace this function with one from horizon.Client when horizon Client has it.
func (c *Client) PaymentsForAccount(accountID string, asc bool, cursor Cursor, limit int) (PaymentsPage, error) {
	c.mu.Lock()
	url := c.url
	if c.http == nil {
		c.http = new(http.Client)
	}
	http := c.http
	c.mu.Unlock()

	order := "desc"
	if asc {
		order = "asc"
	}
	if limit < 1 {
		limit = defaultLimit
	}
	url += "/accounts/" + accountID + "/payments?order=" + order + "&limit=" + strconv.Itoa(limit)
	if cursor != "" {
		url = url + "&cursor=" + string(cursor)
	}
	resp, err := http.Get(url)
	if err != nil {
		return PaymentsPage{}, err
	}

	var page PaymentsPage
	err = decodeResponse(resp, &page)
	return page, err
}

// https://github.com/stellar/go/blob/27a025ec0a44e547ab859aea1636dd3a2f8e59b4/clients/horizon/internal.go#L10-L30
func decodeResponse(resp *http.Response, object interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		horizonError := &Error{
			Response: resp,
		}
		err := decoder.Decode(&horizonError.Problem)
		if err != nil {
			return errors.Wrap(err, "error decoding horizon.Problem")
		}
		return horizonError
	}

	return decoder.Decode(&object)
}
