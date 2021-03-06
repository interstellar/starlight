package fsm

import (
	"encoding/json"
	"time"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/worizon"
)

// Updater contains the state necessary to effect a state transition in a channel.
type Updater struct {
	C          *Channel
	O          Outputter
	H          *WalletAcct
	Seed       []byte
	LedgerTime time.Time
	Passphrase string

	debug bool
}

// Tx causes the updater to update its channel in response to a transaction appearing in a Stellar ledger.
func (u *Updater) Tx(tx *worizon.Tx) error {
	txstr, err := xdr.MarshalBase64(*tx.Env)
	if err != nil {
		return err
	}
	u.debugf("received tx: %s", txstr)
	success := tx.Result.Result.Code == xdr.TransactionResultCodeTxSuccess

	if tx.PT != "" {
		u.C.Cursor = tx.PT
	}
	for _, f := range txHandlerFuncs {
		ok, err := f(u, tx, success)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return errors.WithData(errNoMatch, "tx", tx)

}

// Msg causes the updater to update its channel in response to a protocol message received from a peer Agent.
func (u *Updater) Msg(m *Message) error {
	bytes, err := json.Marshal(*m)
	if err != nil {
		return err
	}
	u.debugf("received message: %s", string(bytes))
	if err := u.verifyMsg(m); err != nil {
		return err
	}
	u.C.CounterpartyMsgIndex = m.MsgNum
	switch {
	case m.ChannelProposeMsg != nil:
		return u.handleChannelProposeMsg(m)

	case m.ChannelAcceptMsg != nil:
		return u.handleChannelAcceptMsg(m)

	case m.PaymentProposeMsg != nil:
		return u.handlePaymentProposeMsg(m)

	case m.PaymentAcceptMsg != nil:
		return u.handlePaymentAcceptMsg(m)

	case m.PaymentCompleteMsg != nil:
		return u.handlePaymentCompleteMsg(m)

	case m.CloseMsg != nil:
		return u.handleCloseMsg(m)
	}
	return errors.New("no message specified")
}

// Cmd causes the updater to update its channel in response to a user command.
func (u *Updater) Cmd(c *Command) error {
	u.debugf("received command: %+v", *c)
	c.Time = u.LedgerTime
	f := commandFuncs[c.Name]
	return f(c, u)
}

// Time causes the updater to update its channel in response to a deadline arriving.
func (u *Updater) Time() error {
	t, err := u.C.TimerTime()
	if err != nil {
		return err
	}
	if t == nil || u.LedgerTime.Before(*t) {
		return nil // nothing to do
	}

	switch u.C.State {
	case AwaitingFunding:
		// PreFundTimeout
		u.debugf("PreFundTimeout...")
		if u.C.Role == Guest {
			return u.transitionTo(Closed)
		}

		// Unreserve wallet balance
		// We should only recover the balance of the funding tx,
		// since both the setup and funding txes have been published.
		// TODO(debnil): test for expected balances.
		u.H.NativeBalance += u.C.fundingBalanceAmount()

		u.C.FundingTimedOut = true
		return u.transitionTo(AwaitingCleanup)

	case ChannelProposed:
		// ChannelProposedTimeout
		u.debugf("ChannelProposedTimeout...")
		if u.C.Role == Host {
			u.H.NativeBalance += u.C.fundingBalanceAmount() + u.C.fundingFeeAmount() + u.C.fundedAcctsTxFeeAmount()
			u.H.Seqnum++
			return u.transitionTo(AwaitingCleanup)
		}
		return nil

	case Open, PaymentProposed, PaymentAccepted, AwaitingClose:
		// RoundTimeout
		u.debugf("RoundTimeout...")
		return u.setForceCloseState()

	case AwaitingSettlementMintime:
		// SettlementMintimeTimeout
		u.debugf("SettlementMintimeTimeout...")
		u.transitionTo(AwaitingSettlement)
	}

	return nil
}

// Close transitions the channel in the given Updater to Closed.
func Close(u *Updater) error {
	return u.transitionTo(Closed)
}

func (u *Updater) verifyMsg(m *Message) error {
	var (
		err error
		kp  keypair.KP
	)
	switch u.C.Role {
	case Guest:
		kp, err = keypair.Parse(u.C.HostAcct.Address())
		if err != nil {
			return err
		}
	case Host:
		kp, err = keypair.Parse(u.C.GuestAcct.Address())
		if err != nil {
			return err
		}
	}
	// Ensure that m has exactly one non-nil field.
	counter := 0
	if m.ChannelProposeMsg != nil {
		counter++
	}
	if m.ChannelAcceptMsg != nil {
		counter++
	}
	if m.PaymentProposeMsg != nil {
		counter++
	}
	if m.PaymentAcceptMsg != nil {
		counter++
	}
	if m.PaymentCompleteMsg != nil {
		counter++
	}
	if m.CloseMsg != nil {
		counter++
	}

	if counter == 0 {
		return errors.New("no message field set")
	}

	if counter != 1 {
		return errors.New("multiple message fields set")
	}

	// Ensure m Version matches software version.
	if m.Version != version {
		return ErrInvalidVersion
	}
	bytes, err := m.bytesToSign()
	if err != nil {
		return err
	}
	return kp.Verify(bytes, m.Signature)
}
