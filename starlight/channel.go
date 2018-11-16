package starlight

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net"
	"github.com/interstellar/starlight/starlight/db"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/internal/update"
	"github.com/interstellar/starlight/starlight/log"
	"github.com/interstellar/starlight/worizon"
	"github.com/interstellar/starlight/worizon/xlm"
)

// watchEscrowAcct watches chanID's escrow account for transactions,
// and updates the channel state in response to them.
//
// It runs until its context is canceled.
func (g *Agent) watchEscrowAcct(ctx context.Context, chanID string) {
	var c fsm.Channel
	var acctReady, canceled <-chan struct{}

	db.View(g.db, func(root *db.Root) error {
		c = *g.getChannel(root, chanID)
		acctReady = g.acctsReady[chanID]
		canceled = g.rootCtx.Done()
		return nil
	})

	if acctReady != nil {
		select {
		case <-acctReady:
			break
		case <-canceled:
			return
		}
	}

	err := g.wclient.StreamTxs(ctx, chanID, (horizon.Cursor)(c.Cursor), func(htx worizon.Transaction) error {
		ftx, err := worizon.NewTx(&htx)
		if err != nil {
			return err
		}
		err = g.preupdateLookups(chanID, ftx)
		if err != nil {
			return err
		}
		return g.updateChannel(chanID, updateFromTxCaller(ftx))
	})
	if err != nil {
		log.Debugf("updating channel %s from tx: %s", string(chanID), err)
		g.mustDeauthenticate()
	}
}

// pollGuestMessages is a goroutine run by the Host for each channel, sending
// requests to the Guest's public URL to get the messages it wants to send to
// the Host.
func (g *Agent) pollGuestMessages(ctx context.Context, chanID string) error {
	var acctReady <-chan struct{}

	var remoteURL string
	var from uint64
	db.View(g.db, func(root *db.Root) error {
		c := g.getChannel(root, chanID)
		remoteURL = c.RemoteURL
		from = c.CounterpartyMsgIndex + 1
		acctReady = g.acctsReady[chanID]
		return nil
	})

	if acctReady != nil {
		select {
		case <-acctReady:
			break
		case <-ctx.Done():
			return nil
		}
	}

	for {
		if ctx.Err() != nil {
			log.Debugf("context canceled, keepAlive(%s) exiting", chanID)
			return nil
		}

		body := fmt.Sprintf(`
		{
			"channel_id":"%s",
			"From":%d
		}`, chanID, from)
		url := strings.TrimRight(remoteURL, "/") + "/api/messages"
		req, err := http.NewRequest("POST", url, strings.NewReader(body))
		if err != nil {
			log.Debug("unexpected error building request", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		resp, err := g.httpclient.Do(req)
		if err != nil {
			log.Debug("unexpected error requesting messages", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			log.Debugf("unexpected bad status %s", resp.Status)
			continue
		}

		var messages []*fsm.Message
		err = json.NewDecoder(resp.Body).Decode(&messages)
		if err != nil {
			log.Debug("unexpected error decoding messages", err)
			continue
		}

		for _, msg := range messages {
			g.updateChannel(msg.ChannelID, func(root *db.Root, updater *fsm.Updater, update *Update) error {
				from = msg.MsgNum + 1
				update.InputMessage = msg
				return updater.Msg(msg)
			})
		}
	}
}

func (g *Agent) preupdateLookups(chanID string, tx *worizon.Tx) error {
	var c *fsm.Channel
	err := db.View(g.db, func(root *db.Root) error {
		c = g.getChannel(root, chanID)
		return nil
	})
	if err != nil {
		return err
	}
	if c.State == fsm.SettingUp {
		baseSeqNum, guestSeqNum, hostSeqNum, err := g.getSequenceNumbers(chanID, c.GuestRatchetAcct, c.HostRatchetAcct)
		if err != nil {
			return err
		}
		return db.Update(g.db, func(root *db.Root) error {
			c = g.getChannel(root, chanID)
			c.BaseSequenceNumber = baseSeqNum
			c.GuestRatchetAcctSeqNum = guestSeqNum
			c.HostRatchetAcctSeqNum = hostSeqNum
			g.putChannel(root, chanID, c)
			return nil
		})
	}
	if fsm.MatchesFundingTx(c, tx) && c.Role == fsm.Guest {
		var escrowAcctID xdr.AccountId
		err := escrowAcctID.SetAddress(string(chanID))
		if err != nil {
			return err
		}
		escrowAcct, err := g.wclient.LoadAccount(escrowAcctID.Address())
		if err != nil {
			return err
		}
		guestRatchetAcct, err := g.wclient.LoadAccount(c.GuestRatchetAcct.Address())
		if err != nil {
			return err
		}
		// don't need to check HostRatchetAccount
		// TODO(dan): reflect that in spec
		escrowSeq, err := strconv.ParseUint(escrowAcct.Sequence, 10, 64)
		if err != nil {
			return err
		}
		guestSeq, err := strconv.ParseUint(guestRatchetAcct.Sequence, 10, 64)
		if err != nil {
			return err
		}
		return db.Update(g.db, func(root *db.Root) error {
			c = g.getChannel(root, chanID)
			if c.BaseSequenceNumber != xdr.SequenceNumber(escrowSeq) {
				return g.closeAfterFunding(root, chanID)
			}
			if c.GuestRatchetAcctSeqNum != xdr.SequenceNumber(guestSeq) {
				return g.closeAfterFunding(root, chanID)
			}
			if len(escrowAcct.Signers) != 2 {
				return g.closeAfterFunding(root, chanID)
			}
			if len(guestRatchetAcct.Signers) != 3 {
				return g.closeAfterFunding(root, chanID)
			}
			nativeBalanceString, err := escrowAcct.GetNativeBalance()
			if err != nil {
				return err
			}
			nativeBalance, err := xlm.Parse(nativeBalanceString)
			if err != nil {
				return err
			}
			if nativeBalance < c.HostAmount+3*xlm.Lumen/2+8*c.ChannelFeerate {
				return g.closeAfterFunding(root, chanID)
			}
			g.putChannel(root, chanID, c)
			return nil
		})
	}
	return nil
}

// closeAfterFunding is called if a Guest detects an inconsistency in the channel
// immediately after it is funded.
// It immediately transitions the channel to Closed.
func (g *Agent) closeAfterFunding(root *db.Root, chanID string) error {
	return g.doUpdateChannel(root, chanID, func(_ *db.Root, updater *fsm.Updater, _ *Update) error {
		return fsm.Close(updater)
	})
}

func updateFromTxCaller(tx *worizon.Tx) func(*db.Root, *fsm.Updater, *Update) error {
	return func(_ *db.Root, updater *fsm.Updater, update *Update) error {
		update.InputTx = tx
		return updater.Tx(tx)
	}
}

// keepAlive runs in its own goroutine, sending a 0-value payment
// once in a while to keep the channel open.
// Both host and guest do this, in case the peer is running
// a different implementation that doesn't.
// It stops after the channel is closed or the done channel closes.
func (g *Agent) keepAlive(ctx context.Context, channelID string) {
	for {
		var ch fsm.Channel
		db.View(g.db, func(root *db.Root) error {
			ch = *g.getChannel(root, channelID)
			return nil
		})

		if ch.State == fsm.Start {
			break // channel has been closed
		}

		timer := time.NewTimer(net.Jitter(ch.MaxRoundDuration / 2))
		select {
		case <-ctx.Done():
			log.Debugf("context canceled, keepAlive(%s) exiting", channelID)
			return

		case <-timer.C:
			// ok
		}

		err := g.DoCommand(channelID, &fsm.Command{
			Name:   fsm.ChannelPay,
			Amount: 0,
		})
		if err != nil {
			log.Debugf("keep-alive payment on channel %s: %s", channelID, err)
		}
	}
}

func (g *Agent) getChannel(root *db.Root, chanID string) *fsm.Channel {
	return root.Agent().Channels().Get([]byte(chanID))
}

func (g *Agent) putChannel(root *db.Root, chanID string, channel *fsm.Channel) {
	root.Agent().Channels().Put([]byte(chanID), channel)
}

// Function startChannel schedules any timer,
// and sets watchers for the channel.
// Must be called from within an update transaction.
func (g *Agent) startChannel(root *db.Root, chanID string) error {
	c := g.getChannel(root, chanID)
	t, err := c.TimerTime()
	if err != nil {
		return err
	}
	if t != nil {
		g.scheduleTimer(root.Tx(), *t, chanID)
	}
	g.watchChannel(root, chanID)
	return nil
}

// updateChannel uses f to update the state of channel chanID.
// It calls f inside a database transaction.
// It is f's responsibility to update the channel struct
// appropriately, and to set one of the InputX fields on the
// Update struct to record the input event.
// When f returns, updateChannel stores the new channel state,
// adds the Update to the updates list,
// and produces any FSM side effects (by calling OutputTo)
// for this transition.
func (g *Agent) updateChannel(chanID string, f func(*db.Root, *fsm.Updater, *Update) error) error {
	return db.Update(g.db, func(root *db.Root) error {
		return g.doUpdateChannel(root, chanID, f)
	})
}

// Must be called from within an update transaction.
func (g *Agent) doUpdateChannel(root *db.Root, chanID string, f func(*db.Root, *fsm.Updater, *Update) error) error {
	if !g.isReadyFunded(root) {
		return errNotFunded
	}

	chans := root.Agent().Channels()
	c := g.getChannel(root, chanID)
	h := root.Agent().Wallet()
	u := &Update{Type: update.ChannelType}
	if c.TopUpAmount != 0 {
		c.TopUpAmount = 0
	}
	o := new(outputter)
	updater := &fsm.Updater{
		C:          c,
		O:          o,
		H:          h,
		Seed:       g.seed,
		LedgerTime: g.wclient.Now(),
		Passphrase: g.passphrase(root),
	}
	err := f(root, updater, u)
	if err != nil {
		return err
	}
	if c.State == fsm.Start {
		return nil // channel (still) does not exist; do not store it
	}

	g.putChannel(root, chanID, c)

	root.Agent().PutWallet(h)
	u.Channel = c
	g.putUpdate(root, u)

	if c.State == fsm.Closed {
		// other states
		switch c.PrevState {
		case fsm.AwaitingCleanup, fsm.AwaitingClose, fsm.AwaitingSettlement:
		default:
			g.putUpdate(root, &Update{
				Type:    update.WarningType,
				Warning: "channel close from unrecoverable error",
				Channel: c,
			})
		}
		// tear down channel
		err = chans.Bucket().Delete([]byte(chanID))
		if err != nil {
			return err
		}
		if canceler := g.cancelers[string(chanID)]; canceler != nil {
			canceler()
			delete(g.cancelers, string(chanID))
		}
		return nil
	}

	// After saving the current state, start the channel, creating cancelers and starting the
	// watch escrow account goroutine. We only want to start new channel watch routines when
	// the update came from a message or a command, otherwise this indicates an update on an
	// already-started channel.
	if c.PrevState == fsm.Start && (u.InputCommand != nil || u.InputMessage != nil) {
		_, err := g.wclient.LoadAccount(c.HostAcct.Address())
		if err != nil {
			return errors.Wrapf(err, "error looking up host account %s", c.HostAcct.Address())
		}
		g.watchChannel(root, chanID)
	}

	// Process any state-transition actions accumulated in o.

	tx := root.Tx()
	for _, stx := range o.txs {
		err := g.addTxTask(tx, chanID, stx)
		if err != nil {
			return err
		}
	}
	for _, m := range o.msgs {
		err := g.addMsgTask(root, c, m)
		if err != nil {
			return err
		}
	}
	t, err := c.TimerTime()
	if err != nil {
		return err
	}
	if t != nil {
		g.scheduleTimer(tx, *t, c.ID)
	}
	return nil
}

// watchChannel sets a watcher for the escrow account,
// and starts a goroutine to make 0-value payments as necessary
// to keep the channel alive.
// Must be called from within an update transaction.
func (g *Agent) watchChannel(root *db.Root, chanID string) {
	ctx, cancel := context.WithCancel(g.rootCtx)
	g.cancelers[string(chanID)] = cancel
	keepAlive := root.Agent().Config().KeepAlive()

	c := g.getChannel(root, chanID)
	switch c.State {
	case fsm.Start, fsm.SettingUp:
		acctReady := make(chan struct{})
		g.acctsReady[chanID] = acctReady
	}

	root.Tx().OnCommit(func() {
		g.allez(func() { g.watchEscrowAcct(ctx, chanID) }, fmt.Sprintf("watchEscrowAcct(%s)", chanID))
		if c.Role == fsm.Host {
			g.allez(func() { g.pollGuestMessages(ctx, chanID) }, fmt.Sprintf("pollGuestMessages(%s)", chanID))
		}
		if keepAlive {
			g.allez(func() { g.keepAlive(ctx, chanID) }, fmt.Sprintf("keepAlive(%s)", chanID))
		}
	})
}
