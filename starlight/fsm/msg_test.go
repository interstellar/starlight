package fsm

import (
	"log"
	"testing"
	"time"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/worizon/xlm"
)

func TestHandleChannelProposeMsg(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	ch.Role = Guest
	ch.KeyIndex = 0
	h := createTestHost()
	m, err := createChannelProposeMsg([]byte(guestSeed), ch, h)
	if err != nil {
		t.Fatal(err)
	}
	u := &Updater{
		C:          ch,
		O:          ono{},
		Seed:       []byte(seed),
		H:          h,
		Passphrase: ch.Passphrase,
	}
	err = u.transitionTo(Closed)
	if err != nil {
		t.Fatal(err)
	}
	err = u.handleChannelProposeMsg(m)
	if err == nil {
		t.Fatal("returned ok with incorrect State")
	}
	err = u.transitionTo(Start)
	if err != nil {
		t.Fatal(err)
	}
	err = u.handleChannelProposeMsg(m)
	if err != nil {
		t.Fatal(err)
	}
	propose := m.ChannelProposeMsg
	if ch.HostAmount != propose.HostAmount {
		t.Fatalf("got HostAmount %v, want %v", ch.HostAmount, propose.HostAmount)
	}
	if ch.MaxRoundDuration != propose.MaxRoundDuration {
		t.Fatalf("got MaxRoundDuration %v, want %v", ch.MaxRoundDuration, propose.MaxRoundDuration)
	}
	if ch.FinalityDelay != propose.FinalityDelay {
		t.Fatalf("got FinalityDelay %v, want %v", ch.FinalityDelay, propose.FinalityDelay)
	}
	if ch.FundingTime != propose.FundingTime {
		t.Fatalf("got FundingTime %v, want %v", ch.FundingTime, propose.FundingTime)
	}
	if ch.MaxRoundDuration != propose.MaxRoundDuration {
		t.Fatalf("got MaxRoundDuration %v, want %v", ch.MaxRoundDuration, propose.MaxRoundDuration)
	}
	if !ch.HostAcct.Equals(propose.HostAcct) {
		t.Fatalf("got HostAcct %v, want %v", ch.HostAcct, propose.HostAcct)
	}
	if !ch.HostRatchetAcct.Equals(propose.HostRatchetAcct) {
		t.Fatalf("got HostRatchetAcct %v, want %v", ch.HostRatchetAcct, propose.HostRatchetAcct)
	}
	if !ch.GuestRatchetAcct.Equals(propose.GuestRatchetAcct) {
		t.Fatalf("got GuestRatchetAcct %v, want %v", ch.GuestRatchetAcct, propose.GuestRatchetAcct)
	}
	if ch.BaseSequenceNumber != propose.BaseSequenceNumber {
		t.Fatalf("got BaseSequenceNumber %v, want %v", ch.BaseSequenceNumber, propose.BaseSequenceNumber)
	}
}

func TestHandleChannelAcceptMessage(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	ch.Role = Guest
	ch.KeyIndex = 0 // this is the KeyIndex for all the Guest's channels
	m, err := createChannelAcceptMsg([]byte(guestSeed), ch, ch.FundingTime)
	if err != nil {
		t.Fatal(err)
	}
	h := createTestHost()
	u := &Updater{
		C:          ch,
		O:          ono{},
		H:          h,
		LedgerTime: ch.FundingTime,
		Seed:       []byte(guestSeed),
	}
	u.transitionTo(Start)
	if err != nil {
		t.Fatal(err)
	}
	err = u.handleChannelAcceptMsg(m)
	if errors.Root(err) != errUnexpectedState {
		t.Fatalf("got error %s, want %s", err, errUnexpectedState)
	}
	u.transitionTo(ChannelProposed)
	err = u.handleChannelAcceptMsg(m)
	if ch.Role != Guest {
		t.Fatalf("got Role %v, want %v", ch.Role, Guest)
	}
	if ch.State != ChannelProposed {
		t.Fatalf("got State %v, want %v", ch.State, ChannelProposed)
	}
	ch.Role = Host
	err = u.handleChannelAcceptMsg(m)
	if err != nil {
		t.Fatal(err)
	}
	if ch.State != AwaitingFunding {
		t.Fatalf("got State %v, want %v", ch.State, AwaitingFunding)
	}
	u.transitionTo(ChannelProposed)
	if err != nil {
		t.Fatal(err)
	}
	m, err = createChannelAcceptMsg([]byte(""), ch, ch.FundingTime)
	if err != nil {
		t.Fatal(err)
	}
	err = u.handleChannelAcceptMsg(m)
	if err == nil {
		t.Fatal("correctly verified incorrect signature")
	}
}

func TestHandlePaymentProposeMessage(t *testing.T) {
	cases := []struct {
		name         string
		msgFunc      func(m *PaymentProposeMsg)
		senderUFunc  func(u *Updater)
		handlerUFunc func(u *Updater)
		wantErr      error
		wantState    State
		handler      Role
	}{
		{
			name:         "incorrect state",
			handlerUFunc: func(u *Updater) { u.transitionTo(Closed) },
			wantState:    Closed,
			wantErr:      errUnexpectedState,
			handler:      Guest,
		},
		{
			name:      "invalid ledger time",
			msgFunc:   func(m *PaymentProposeMsg) { m.PaymentTime = m.PaymentTime.Add(-1 * time.Hour) },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Guest,
		},
		{
			name:         "incorrect channel round number",
			handlerUFunc: func(u *Updater) { u.C.RoundNumber += 2 },
			wantErr:      keypair.ErrInvalidSignature,
			wantState:    Open,
			handler:      Guest,
		},
		{
			name:      "invalid payment amount",
			msgFunc:   func(m *PaymentProposeMsg) { m.PaymentAmount = 3 * xlm.Lumen },
			wantErr:   nil,
			wantState: Open,
			handler:   Guest,
		},
		{
			name: "guest zero guest balance unused sig error",
			senderUFunc: func(u *Updater) {
				u.C.GuestAmount = -1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			handlerUFunc: func(u *Updater) {
				u.C.GuestAmount = -1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			msgFunc: func(m *PaymentProposeMsg) {
				m.SenderSettleWithGuestSig = m.SenderSettleWithHostSig
			},
			wantErr:   errUnusedSettleWithGuestSig,
			wantState: Open,
			handler:   Guest,
		},
		{
			name: "guest zero guest balance nil sig success",
			senderUFunc: func(u *Updater) {
				u.C.GuestAmount = -1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			handlerUFunc: func(u *Updater) {
				u.C.GuestAmount = -1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			wantErr:   nil,
			wantState: PaymentAccepted,
			handler:   Guest,
		},
		{
			name:      "guest handle insufficient payment amount",
			msgFunc:   func(m *PaymentProposeMsg) { m.PaymentAmount = -1 },
			wantErr:   nil,
			wantState: Open,
			handler:   Guest,
		},
		{
			name:      "guest handle success",
			wantErr:   nil,
			wantState: PaymentAccepted,
			handler:   Guest,
		},
		{
			name:      "guest handle invalid settle with guest signature",
			msgFunc:   func(m *PaymentProposeMsg) { m.SenderSettleWithGuestSig = xdr.DecoratedSignature{} },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Guest,
		},
		{
			name:      "guest handle invalid settle with host signature",
			msgFunc:   func(m *PaymentProposeMsg) { m.SenderSettleWithHostSig = xdr.DecoratedSignature{} },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Guest,
		},
		{
			name:      "host handle success",
			wantErr:   nil,
			wantState: PaymentAccepted,
			handler:   Host,
		},
		{
			name:      "host handle insufficient payment",
			msgFunc:   func(m *PaymentProposeMsg) { m.PaymentAmount = -1 },
			wantErr:   nil,
			wantState: Open,
			handler:   Host,
		},
		{
			name:      "host handle invalid settle with guest signature",
			msgFunc:   func(m *PaymentProposeMsg) { m.SenderSettleWithGuestSig = xdr.DecoratedSignature{} },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Host,
		},
		{
			name:      "host handle invalid settle with host signature",
			msgFunc:   func(m *PaymentProposeMsg) { m.SenderSettleWithHostSig = xdr.DecoratedSignature{} },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Host,
		},
		{
			name: "host zero guest balance unused sig error",
			senderUFunc: func(u *Updater) {
				u.C.GuestAmount = 1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			handlerUFunc: func(u *Updater) {
				u.C.GuestAmount = 1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			msgFunc: func(m *PaymentProposeMsg) {
				m.SenderSettleWithGuestSig = m.SenderSettleWithHostSig
			},
			wantErr:   errUnusedSettleWithGuestSig,
			wantState: Open,
			handler:   Host,
		},
		{
			name: "host zero guest balance nil sig success",
			senderUFunc: func(u *Updater) {
				u.C.GuestAmount = 1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			handlerUFunc: func(u *Updater) {
				u.C.GuestAmount = 1 * xlm.Microlumen
				u.C.PendingAmountSent = 1 * xlm.Microlumen
			},
			wantErr:   nil,
			wantState: PaymentAccepted,
			handler:   Host,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hostChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			hostChannel.Role = Host
			u := &Updater{
				C: hostChannel,
				O: ono{},
			}
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}

			guestChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			guestChannel.Role = Guest
			guestChannel.KeyIndex = 0
			u.C = guestChannel
			u.Seed = []byte(guestSeed)
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}

			var sender, recipient *Channel
			var senderSeed, recipientSeed []byte

			switch c.handler {
			case Host:
				sender, senderSeed = guestChannel, []byte(guestSeed)
				recipient, recipientSeed = hostChannel, []byte(hostSeed)
			case Guest:
				sender, senderSeed = hostChannel, []byte(hostSeed)
				recipient, recipientSeed = guestChannel, []byte(guestSeed)
			}
			command := &Command{
				Name:   ChannelPay,
				Amount: sender.PendingAmountSent,
				Time:   sender.PendingPaymentTime,
			}
			h := createTestHost()

			u.C = sender
			u.Seed = senderSeed
			u.H = h
			err = channelPayFn(command, u)
			if err != nil {
				t.Error(err)
			}

			if c.senderUFunc != nil {
				c.senderUFunc(u)
			}
			m, err := createPaymentProposeMsg(senderSeed, sender)
			if err != nil {
				t.Error(err)
			}

			u = &Updater{
				C:          recipient,
				O:          ono{},
				LedgerTime: sender.PendingPaymentTime.Add(time.Minute),
				Seed:       recipientSeed,
			}

			if c.handlerUFunc != nil {
				c.handlerUFunc(u)
			}
			if c.msgFunc != nil {
				c.msgFunc(m.PaymentProposeMsg)
			}

			err = u.handlePaymentProposeMsg(m)
			gotState := recipient.State

			if errors.Root(err) != c.wantErr {
				t.Errorf("got error %v, want %v", err, c.wantErr)
			}
			if gotState != c.wantState {
				t.Errorf("got state %v, want %v", gotState, c.wantState)
			}
		})
	}
}

func TestHandlePaymentAcceptMessage(t *testing.T) {
	cases := []struct {
		name    string
		chFunc  func(ch *Channel)
		msgFunc func(m *PaymentAcceptMsg)
		wantErr error
		sender  Role
	}{
		{
			name:    "host success",
			wantErr: nil,
			sender:  Host,
		},
		{
			name:    "host wrong guest amount",
			chFunc:  func(ch *Channel) { ch.GuestAmount = 100 * xlm.Lumen },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Host,
		},
		{
			name:    "host wrong host ratchet acct",
			chFunc:  func(ch *Channel) { ch.HostRatchetAcct = ch.EscrowAcct },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Host,
		},
		{
			name:    "host wrong guest ratchet seq num",
			chFunc:  func(ch *Channel) { ch.GuestRatchetAcctSeqNum++ },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Host,
		},
		{
			name:    "host wrong key",
			chFunc:  func(ch *Channel) { ch.GuestAcct = ch.HostRatchetAcct },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Host,
		},
		{
			name:    "host wrong ratchet sig",
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientRatchetSig = m.RecipientSettleWithHostSig },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Host,
		},
		{
			name:    "host wrong settle with guest sig",
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientSettleWithGuestSig = &m.RecipientRatchetSig },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Host,
		},
		{
			name:    "host wrong settle with host sig",
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientSettleWithHostSig = m.RecipientRatchetSig },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Host,
		},
		{
			name: "host zero guest balance not-nil sig",
			chFunc: func(ch *Channel) {
				ch.GuestAmount = -1 * xlm.Microlumen
				ch.PendingAmountSent = 1 * xlm.Microlumen
			},
			wantErr: errUnusedSettleWithGuestSig,
			sender:  Host,
		},
		{
			name: "host zero guest balance nil sig",
			chFunc: func(ch *Channel) {
				ch.GuestAmount = -1 * xlm.Microlumen
				ch.PendingAmountSent = 1 * xlm.Microlumen
			},
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientSettleWithGuestSig = nil },
			wantErr: nil,
			sender:  Host,
		},
		{
			name:    "guest success",
			wantErr: nil,
			sender:  Guest,
		},
		{
			name:    "guest wrong guest amount",
			chFunc:  func(ch *Channel) { ch.GuestAmount = 100 * xlm.Lumen },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name:    "guest wrong guest ratchet acct",
			chFunc:  func(ch *Channel) { ch.GuestRatchetAcct = ch.EscrowAcct },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name:    "guest wrong host ratchet acct",
			chFunc:  func(ch *Channel) { ch.HostRatchetAcct = ch.EscrowAcct },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name:    "guest wrong host ratchet seq num",
			chFunc:  func(ch *Channel) { ch.HostRatchetAcctSeqNum++ },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name:    "guest wrong key",
			chFunc:  func(ch *Channel) { ch.GuestAcct = ch.HostRatchetAcct },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name:    "guest wrong ratchet sig",
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientRatchetSig = m.RecipientSettleWithHostSig },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name:    "guest wrong settle with guest sig",
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientSettleWithGuestSig = &m.RecipientRatchetSig },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name:    "guest wrong settle with host sig",
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientSettleWithHostSig = m.RecipientRatchetSig },
			wantErr: keypair.ErrInvalidSignature,
			sender:  Guest,
		},
		{
			name: "guest zero guest balance not nil sig",
			chFunc: func(ch *Channel) {
				ch.GuestAmount = 1 * xlm.Microlumen
				ch.PendingAmountSent = 1 * xlm.Microlumen
			},
			wantErr: errUnusedSettleWithGuestSig,
			sender:  Guest,
		},
		{
			name: "guest zero guest balance nil sig",
			chFunc: func(ch *Channel) {
				ch.GuestAmount = 1 * xlm.Microlumen
				ch.PendingAmountSent = 1 * xlm.Microlumen
			},
			msgFunc: func(m *PaymentAcceptMsg) { m.RecipientSettleWithGuestSig = nil },
			wantErr: nil,
			sender:  Guest,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hostChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			hostChannel.Role = Host
			u := &Updater{
				C: hostChannel,
				O: ono{},
			}
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}

			guestChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			guestChannel.Role = Guest
			u.C = guestChannel
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}
			guestChannel.KeyIndex = 0

			var sender, recipient *Channel
			var senderSeed, recipientSeed []byte

			switch c.sender {
			case Guest:
				sender, senderSeed = guestChannel, []byte(guestSeed)
				recipient, recipientSeed = hostChannel, []byte(hostSeed)
			case Host:
				sender, senderSeed = hostChannel, []byte(hostSeed)
				recipient, recipientSeed = guestChannel, []byte(guestSeed)
			}
			payCmd := &Command{
				Name:   ChannelPay,
				Amount: sender.PendingAmountSent,
				Time:   sender.PendingPaymentTime,
			}
			h := createTestHost()

			u = &Updater{
				C:    sender,
				O:    ono{},
				H:    h,
				Seed: senderSeed,
			}

			err = channelPayFn(payCmd, u)
			if err != nil {
				t.Error(err)
			}
			proposalMsg, err := createPaymentProposeMsg(senderSeed, sender)
			if err != nil {
				t.Error(err)
			}
			u = &Updater{
				C:          recipient,
				O:          ono{},
				LedgerTime: sender.PendingPaymentTime.Add(time.Minute),
				Seed:       recipientSeed,
			}
			err = u.handlePaymentProposeMsg(proposalMsg)
			if err != nil {
				t.Error(err)
			}

			acceptMsg, err := createPaymentAcceptMsg(recipientSeed, recipient)
			if err != nil {
				t.Error(err)
			}

			if c.chFunc != nil {
				c.chFunc(sender)
			}
			if c.msgFunc != nil {
				c.msgFunc(acceptMsg.PaymentAcceptMsg)
			}
			u.C = sender
			u.Seed = senderSeed
			err = u.handlePaymentAcceptMsg(acceptMsg)

			if errors.Root(err) != c.wantErr {
				t.Errorf("got error %v, want %v", err, c.wantErr)
			}
		})
	}
}

func TestHandlePaymentCompleteMessage(t *testing.T) {
	successfulTransition := Open
	failedTransition := PaymentAccepted
	cases := []struct {
		name      string
		msgFunc   func(m *PaymentCompleteMsg)
		chFunc    func(ch *Channel)
		wantErr   error
		wantState State
		recipient Role
	}{
		{
			name:      "incorrect starting state",
			chFunc:    func(ch *Channel) { ch.State = Closed },
			wantErr:   errUnexpectedState,
			wantState: Closed,
			recipient: Guest,
		},
		{
			name:      "guest success",
			wantErr:   nil,
			wantState: successfulTransition,
			recipient: Guest,
		},
		// Incorrect Amounts do not affect correctness in handleCompleteMsg.
		{
			name:      "guest wrong guest amount",
			chFunc:    func(ch *Channel) { ch.GuestAmount = 0 },
			wantErr:   nil,
			wantState: successfulTransition,
			recipient: Guest,
		},
		{
			name:      "guest wrong host amount",
			chFunc:    func(ch *Channel) { ch.HostAmount = 0 },
			wantErr:   nil,
			wantState: successfulTransition,
			recipient: Guest,
		},
		{
			name:      "guest wrong ratchet acct",
			chFunc:    func(ch *Channel) { ch.HostRatchetAcct = ch.GuestRatchetAcct },
			wantState: failedTransition,
			wantErr:   keypair.ErrInvalidSignature,
			recipient: Guest,
		},
		{
			name:      "guest wrong ratchet acct seq num",
			chFunc:    func(ch *Channel) { ch.HostRatchetAcctSeqNum++ },
			wantState: failedTransition,
			wantErr:   keypair.ErrInvalidSignature,
			recipient: Guest,
		},
		{
			name:      "guest wrong sender key address",
			chFunc:    func(ch *Channel) { ch.EscrowAcct = ch.HostRatchetAcct },
			wantState: failedTransition,
			wantErr:   keypair.ErrInvalidSignature,
			recipient: Guest,
		},
		{
			name:      "host success",
			wantErr:   nil,
			wantState: successfulTransition,
			recipient: Host,
		},
		// Incorrect Amounts do not affect correctness in handleCompleteMsg.
		{
			name:      "host wrong guest amount",
			chFunc:    func(ch *Channel) { ch.GuestAmount = 0 },
			wantErr:   nil,
			wantState: successfulTransition,
			recipient: Host,
		},
		{
			name:      "host wrong host amount",
			chFunc:    func(ch *Channel) { ch.HostAmount = 0 },
			wantErr:   nil,
			wantState: successfulTransition,
			recipient: Host,
		},
		{
			name:      "host wrong ratchet acct",
			chFunc:    func(ch *Channel) { ch.GuestRatchetAcct = ch.HostRatchetAcct },
			wantState: failedTransition,
			wantErr:   keypair.ErrInvalidSignature,
			recipient: Host,
		},
		{
			name:      "host wrong ratchet acct seq num",
			chFunc:    func(ch *Channel) { ch.GuestRatchetAcctSeqNum++ },
			wantState: failedTransition,
			wantErr:   keypair.ErrInvalidSignature,
			recipient: Host,
		},
		{
			name:      "guest wrong sender key address",
			chFunc:    func(ch *Channel) { ch.GuestAcct = ch.HostRatchetAcct },
			wantState: failedTransition,
			wantErr:   keypair.ErrInvalidSignature,
			recipient: Host,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hostChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			hostChannel.Role = Host
			u := &Updater{
				C: hostChannel,
				O: ono{},
			}
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}

			guestChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			guestChannel.Role = Guest
			u.C = guestChannel
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}
			guestChannel.KeyIndex = 0

			var sender, recipient *Channel
			var senderSeed, recipientSeed []byte

			switch c.recipient {
			case Host:
				sender, senderSeed = guestChannel, []byte(guestSeed)
				recipient, recipientSeed = hostChannel, []byte(hostSeed)
			case Guest:
				sender, senderSeed = hostChannel, []byte(hostSeed)
				recipient, recipientSeed = guestChannel, []byte(guestSeed)
			}
			payCmd := &Command{
				Name:   ChannelPay,
				Amount: sender.PendingAmountSent,
				Time:   sender.PendingPaymentTime,
			}
			h := createTestHost()
			u = &Updater{
				C:          sender,
				O:          ono{},
				H:          h,
				LedgerTime: sender.PendingPaymentTime.Add(time.Minute),
				Seed:       senderSeed,
			}
			err = channelPayFn(payCmd, u)
			if err != nil {
				t.Error(err)
			}
			proposalMsg, err := createPaymentProposeMsg(senderSeed, sender)
			if err != nil {
				t.Error(err)
			}
			u.C = recipient
			u.Seed = recipientSeed
			err = u.handlePaymentProposeMsg(proposalMsg)
			if err != nil {
				t.Error(err)
			}

			acceptMsg, err := createPaymentAcceptMsg(recipientSeed, recipient)
			if err != nil {
				t.Error(err)
			}

			u.C = sender
			u.Seed = senderSeed
			err = u.handlePaymentAcceptMsg(acceptMsg)
			if err != nil {
				t.Error(err)
			}

			completeMsg, err := createPaymentCompleteMsg(senderSeed, sender)
			if err != nil {
				t.Error(err)
			}
			if c.chFunc != nil {
				c.chFunc(recipient)
			}
			if c.msgFunc != nil {
				c.msgFunc(completeMsg.PaymentCompleteMsg)
			}
			u.C = recipient
			u.Seed = recipientSeed
			err = u.handlePaymentCompleteMsg(completeMsg)
			if errors.Root(err) != c.wantErr {
				t.Errorf("got error %v, want %v", err, c.wantErr)
			}
			if recipient.State != c.wantState {
				t.Errorf("got state %v, want %v", recipient.State, c.wantState)
			}
		})
	}
}

func TestHandleCloseMsg(t *testing.T) {
	cases := []struct {
		name      string
		msgFunc   func(m *CloseMsg)
		chFunc    func(ch *Channel)
		wantErr   error
		wantState State
		handler   Role
	}{
		{
			name:      "incorrect starting state",
			chFunc:    func(ch *Channel) { ch.State = Closed },
			wantErr:   errUnexpectedState,
			wantState: Closed,
			handler:   Host,
		},
		{
			name:      "coop close tx tampering",
			chFunc:    func(ch *Channel) { ch.BaseSequenceNumber++ },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Host,
		},
		{
			name:      "host success",
			wantState: AwaitingClose,
			handler:   Host,
		},
		{
			name:      "host wrong key",
			chFunc:    func(ch *Channel) { ch.EscrowAcct = ch.GuestAcct },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Host,
		},
		{
			name:      "host payment proposed success",
			chFunc:    func(ch *Channel) { ch.State = PaymentProposed },
			wantState: AwaitingClose,
			handler:   Host,
		},
		{
			name:      "guest success",
			wantState: AwaitingClose,
			handler:   Guest,
		},
		{
			name:      "guest wrong key",
			chFunc:    func(ch *Channel) { ch.GuestAcct = ch.EscrowAcct },
			wantErr:   keypair.ErrInvalidSignature,
			wantState: Open,
			handler:   Guest,
		},
		{
			name:      "guest payment proposed success",
			chFunc:    func(ch *Channel) { ch.State = PaymentProposed },
			wantState: AwaitingClose,
			handler:   Guest,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			log.Println(c.name)
			hostChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			hostChannel.Role = Host
			u := &Updater{
				C: hostChannel,
				O: ono{},
			}
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}

			guestChannel, err := createTestChannel()
			if err != nil {
				t.Fatal(err)
			}
			guestChannel.Role = Guest
			u.C = guestChannel
			err = u.transitionTo(Open)
			if err != nil {
				t.Fatal(err)
			}
			guestChannel.KeyIndex = 0

			var sender, recipient *Channel
			var senderSeed, recipientSeed []byte

			switch c.handler {
			case Host:
				sender, senderSeed = guestChannel, []byte(guestSeed)
				recipient, recipientSeed = hostChannel, []byte(hostSeed)
				recipient.PendingAmountSent = 0
			case Guest:
				sender, senderSeed = hostChannel, []byte(hostSeed)
				recipient, recipientSeed = guestChannel, []byte(guestSeed)
				recipient.PendingAmountSent = 0
			}

			h := createTestHost()
			payCmd := &Command{
				Name:   ChannelPay,
				Amount: sender.PendingAmountSent,
				Time:   sender.PendingPaymentTime,
			}
			u = &Updater{
				C:          sender,
				O:          ono{},
				H:          h,
				LedgerTime: sender.PendingPaymentTime,
				Seed:       senderSeed,
			}
			err = channelPayFn(payCmd, u)
			if err != nil {
				t.Error(err)
			}
			proposalMsg, err := createPaymentProposeMsg(senderSeed, sender)
			if err != nil {
				t.Error(err)
			}
			u.C = recipient
			u.LedgerTime = sender.PendingPaymentTime.Add(time.Minute)
			u.Seed = recipientSeed
			err = u.handlePaymentProposeMsg(proposalMsg)
			if err != nil {
				t.Error(err)
			}

			acceptMsg, err := createPaymentAcceptMsg(recipientSeed, recipient)
			if err != nil {
				t.Error(err)
			}

			u.C = sender
			u.Seed = senderSeed
			err = u.handlePaymentAcceptMsg(acceptMsg)
			if err != nil {
				t.Error(err)
			}

			completeMsg, err := createPaymentCompleteMsg(senderSeed, sender)
			if err != nil {
				t.Error(err)
			}
			u.C = recipient
			u.Seed = recipientSeed
			err = u.handlePaymentCompleteMsg(completeMsg)

			m, err := createCloseMsg(senderSeed, sender)
			if err != nil {
				t.Error(err)
			}

			if c.chFunc != nil {
				c.chFunc(recipient)
			}
			if c.msgFunc != nil {
				c.msgFunc(m.CloseMsg)
			}

			u.C = recipient
			u.Seed = recipientSeed
			err = u.handleCloseMsg(m)
			if errors.Root(err) != c.wantErr {
				t.Errorf("got error %v, want %v", err, c.wantErr)
			}
			if recipient.State != c.wantState {
				t.Errorf("got state %v, want %v", recipient.State, c.wantState)
			}
		})
	}
}

func TestMessageAuthentication(t *testing.T) {
	recipient, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	recipient.KeyIndex = 0
	recipient.Role = Guest
	u := &Updater{
		C:    recipient,
		O:    ono{},
		Seed: []byte(guestSeed),
	}
	sender, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	sender.Role = Host
	h := createTestHost()
	m, err := createChannelProposeMsg([]byte(hostSeed), sender, h)
	if err != nil {
		t.Fatal(err)
	}
	err = u.verifyMsg(m)
	if err != nil {
		t.Fatal(err)
	}
	m, err = createChannelProposeMsg([]byte(guestSeed), sender, h)
	if err != nil {
		t.Fatal(err)
	}
	err = u.verifyMsg(m)
	if err != keypair.ErrInvalidSignature {
		t.Fatalf("got %s, want %s", err, keypair.ErrInvalidSignature)
	}
}
