package fsm

import (
	"log"
	"time"

	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/errors"
)

type Updater struct {
	C          *Channel
	O          Outputter
	H          *WalletAcct
	Seed       []byte
	LedgerTime time.Time
	Passphrase string
}

func (u *Updater) Tx(tx *Tx) error {
	log.Printf("received tx: %+v", *tx)
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

func (u *Updater) Msg(m *Message) error {
	log.Printf("received message: %+v", *m)
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

func (u *Updater) Cmd(c *Command) error {
	log.Printf("received command: %+v", *c)
	c.Time = u.LedgerTime
	f := commandFuncs[c.UserCommand]
	return f(c, u)
}

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
		log.Printf("PreFundTimeout...")
		if u.C.Role == Guest {
			return u.transitionTo(Closed)
		}

		// Unreserve wallet balance
		// TODO(bobg): double-check the logic here and test for expected balances.
		// In particular,
		// we should not expect to recover the fees of any txs that have been published.
		u.H.Balance += u.C.SetupAndFundingReserveAmount()

		u.C.FundingTimedOut = true
		return u.transitionTo(AwaitingCleanup)

	case ChannelProposed:
		// ChannelProposedTimeout
		log.Printf("ChannelProposedTimeout...")
		if u.C.Role == Host {
			u.H.Balance += u.C.SetupAndFundingReserveAmount()
			u.H.Seqnum++
			return u.transitionTo(AwaitingCleanup)
		}
		return nil

	case Open, PaymentProposed, PaymentAccepted, AwaitingClose:
		// RoundTimeout
		log.Printf("RoundTimeout...")
		return u.setForceCloseState()

	case AwaitingSettlementMintime:
		// SettlementMintimeTimeout
		log.Printf("SettlementMintimeTimeout...")
		u.transitionTo(AwaitingSettlement)
	}

	return nil
}

// Close transitions the channel in the given Updater to Closed.
func Close(u *Updater) error {
	return u.transitionTo(Closed)
}
