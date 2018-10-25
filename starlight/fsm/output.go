package fsm

import (
	"time"

	b "github.com/stellar/go/build"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/starlight/key"
)

// Outputter accumulates side effects to be emitted to the outside world.
// This consists of zero or more
// messages to send to the peer,
// transactions to publish to the Stellar network,
// and timers to schedule.
type Outputter interface {
	// OutputMsg sends a message to the remote endpoint.
	OutputMsg(*Message)

	// OutputTx publishes a transaction to the ledger.
	OutputTx(xdr.TransactionEnvelope)
}

func publishFundingTx(seed []byte, ch *Channel, o Outputter, h *WalletAcct) error {
	tx, err := buildFundingTx(ch, h)
	if err != nil {
		return err
	}
	ch.FundingTxSeqnum = h.Seqnum
	env, err := txSig(tx, seed, key.PrimaryAccountIndex, ch.KeyIndex, ch.KeyIndex+1, ch.KeyIndex+2)
	if err != nil {
		return err
	}
	o.OutputTx(*env.E)

	return nil
}

func publishCleanupTx(seed []byte, ch *Channel, o Outputter, h *WalletAcct) error {
	var seqnum xdr.SequenceNumber
	if ch.FundingTimedOut {
		seqnum = ch.FundingTxSeqnum
	} else {
		seqnum = h.Seqnum
	}

	tx, err := b.Transaction(
		b.Network{Passphrase: ch.Passphrase},
		b.SourceAccount{AddressOrSeed: ch.HostAcct.Address()},
		b.Sequence{Sequence: uint64(seqnum)},
		b.BaseFee{Amount: uint64(ch.HostFeerate)},
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.HostRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.GuestRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
	)
	if err != nil {
		return err
	}
	env, err := txSig(tx, seed, key.PrimaryAccountIndex, ch.KeyIndex, ch.KeyIndex+1, ch.KeyIndex+2)
	if err != nil {
		return err
	}
	o.OutputTx(*env.E)
	return nil
}

func publishCoopCloseTx(seed []byte, ch *Channel, o Outputter, h *WalletAcct) error {
	coopCloseTx, err := buildCooperativeCloseTx(ch)
	if err != nil {
		return err
	}
	channelCoopCloseSig, err := detachedSig(coopCloseTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return err
	}
	env := xdr.TransactionEnvelope{
		Tx:         *coopCloseTx.TX,
		Signatures: []xdr.DecoratedSignature{channelCoopCloseSig, ch.CounterpartyCoopCloseSig}}
	o.OutputTx(env)
	return nil
}

func publishTopUpTx(seed []byte, ch *Channel, o Outputter, h *WalletAcct) error {
	tx, err := b.Transaction(
		b.Network{Passphrase: ch.Passphrase},
		b.SourceAccount{AddressOrSeed: ch.HostAcct.Address()},
		b.Sequence{Sequence: uint64(h.Seqnum)},
		b.BaseFee{Amount: uint64(ch.HostFeerate)},
		b.Payment(
			b.SourceAccount{AddressOrSeed: ch.HostAcct.Address()},
			b.Destination{AddressOrSeed: ch.EscrowAcct.Address()},
			b.NativeAmount{Amount: ch.TopUpAmount.HorizonString()},
		),
	)
	if err != nil {
		return err
	}
	env, err := txSig(tx, seed, key.PrimaryAccountIndex)
	if err != nil {
		return err
	}
	o.OutputTx(*env.E)
	return nil
}

func createPaymentCompleteMsg(seed []byte, ch *Channel) (*Message, error) {
	var ratchetAccount AccountID
	var ratchetSeqNum xdr.SequenceNumber
	switch ch.Role {
	case Guest:
		ratchetAccount = ch.GuestRatchetAcct
		ratchetSeqNum = ch.GuestRatchetAcctSeqNum
	case Host:
		ratchetAccount = ch.HostRatchetAcct
		ratchetSeqNum = ch.HostRatchetAcctSeqNum
	}
	senderRatchetTx, err := buildRatchetTx(ch, ch.PendingPaymentTime, ratchetAccount, ratchetSeqNum)
	if err != nil {
		return nil, err
	}
	senderRatchetSig, err := detachedSig(senderRatchetTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return nil, err
	}
	m := &Message{
		ChannelID: ch.ID,
		PaymentCompleteMsg: &PaymentCompleteMsg{
			RoundNumber:      ch.RoundNumber,
			SenderRatchetSig: senderRatchetSig,
		},
		Version: version,
	}
	return m.signMsg(seed, ch.KeyIndex)
}

func sendPaymentCompleteMsg(seed []byte, ch *Channel, o Outputter) error {
	m, err := createPaymentCompleteMsg(seed, ch)
	if err != nil {
		return err
	}
	o.OutputMsg(m)
	return nil
}

func publishSetupAccountTxes(seed []byte, ch *Channel, o Outputter, h *WalletAcct) error {
	hostRatchetTx, err := buildSetupAccountTx(ch, ch.HostRatchetAcct, h.Seqnum-2)
	if err != nil {
		return err
	}
	hostRatchet, err := txSig(hostRatchetTx, seed, key.PrimaryAccountIndex)
	if err != nil {
		return err
	}
	o.OutputTx(*hostRatchet.E)

	guestRatchetTx, err := buildSetupAccountTx(ch, ch.GuestRatchetAcct, h.Seqnum-1)
	if err != nil {
		return err
	}
	guestRatchet, err := txSig(guestRatchetTx, seed, key.PrimaryAccountIndex)
	if err != nil {
		return err
	}
	o.OutputTx(*guestRatchet.E)

	escrowTx, err := buildSetupAccountTx(ch, ch.EscrowAcct, h.Seqnum)
	if err != nil {
		return err
	}
	escrow, err := txSig(escrowTx, seed, key.PrimaryAccountIndex)
	if err != nil {
		return err
	}
	o.OutputTx(*escrow.E)

	return nil
}

func createChannelProposeMsg(seed []byte, ch *Channel, h *WalletAcct) (*Message, error) {
	m := &Message{
		ChannelID: ch.ID,
		ChannelProposeMsg: &ChannelProposeMsg{
			HostAcct:            ch.HostAcct, // TODO(bobg): spec wants the federation address, can it be derived from this?
			GuestAcct:           ch.GuestAcct,
			HostRatchetAcct:     ch.HostRatchetAcct,
			GuestRatchetAcct:    ch.GuestRatchetAcct,
			MaxRoundDuration:    ch.MaxRoundDuration,
			FinalityDelay:       ch.FinalityDelay,
			HostAmount:          ch.HostAmount,
			FundingTime:         ch.FundingTime,
			BaseSequenceNumber:  xdr.SequenceNumber(ch.BaseSequenceNumber),
			Feerate:             ch.ChannelFeerate,
			CounterpartyAddress: h.Address,
		},
		Version: version,
	}
	return m.signMsg(seed, ch.KeyIndex)
}

func sendChannelProposeMsg(seed []byte, ch *Channel, o Outputter, h *WalletAcct) error {
	m, err := createChannelProposeMsg(seed, ch, h)
	if err != nil {
		return err
	}
	o.OutputMsg(m)
	return nil
}

func createPaymentProposeMsg(seed []byte, ch *Channel) (*Message, error) {
	// We copy the Channel to construct signatures with updated PaymentAmount values.
	ch2 := *ch
	switch ch.Role {
	case Guest:
		ch2.GuestAmount -= ch.PendingAmountSent
		ch2.HostAmount += ch.PendingAmountSent
	case Host:
		ch2.HostAmount -= ch.PendingAmountSent
		ch2.GuestAmount += ch.PendingAmountSent
	}

	var settleWithHostSig, settleWithGuestSig xdr.DecoratedSignature
	if ch2.GuestAmount == 0 {
		settleOnlyWithHostTx, err := buildSettleOnlyWithHostTx(&ch2)
		if err != nil {
			return nil, err
		}
		settleWithHostSig, err = detachedSig(settleOnlyWithHostTx.TX, seed, ch2.Passphrase, ch2.KeyIndex)
		if err != nil {
			return nil, err
		}
	} else {
		settleWithGuestTx, err := buildSettleWithGuestTx(&ch2, ch2.PendingPaymentTime)
		if err != nil {
			return nil, err
		}
		settleWithGuestSig, err = detachedSig(settleWithGuestTx.TX, seed, ch2.Passphrase, ch2.KeyIndex)
		if err != nil {
			return nil, err
		}
		settleWithHostTx, err := buildSettleWithHostTx(&ch2, ch2.PendingPaymentTime)
		if err != nil {
			return nil, err
		}
		settleWithHostSig, err = detachedSig(settleWithHostTx.TX, seed, ch2.Passphrase, ch2.KeyIndex)
		if err != nil {
			return nil, err
		}
	}
	m := &Message{
		ChannelID: ch2.ID,
		PaymentProposeMsg: &PaymentProposeMsg{
			RoundNumber:              uint64(ch2.RoundNumber),
			PaymentTime:              ch2.PendingPaymentTime,
			PaymentAmount:            ch2.PendingAmountSent,
			SenderSettleWithGuestSig: settleWithGuestSig,
			SenderSettleWithHostSig:  settleWithHostSig,
		},
		Version: version,
	}
	return m.signMsg(seed, ch.KeyIndex)
}

func sendPaymentProposeMsg(seed []byte, ch *Channel, o Outputter) error {
	m, err := createPaymentProposeMsg(seed, ch)
	if err != nil {
		return err
	}
	o.OutputMsg(m)
	return nil
}

func createPaymentAcceptMsg(seed []byte, ch *Channel) (*Message, error) {
	var ratchetAccount AccountID
	var ratchetSeqNum xdr.SequenceNumber
	switch ch.Role {
	case Guest:
		ratchetAccount = ch.GuestRatchetAcct
		ratchetSeqNum = ch.GuestRatchetAcctSeqNum
	case Host:
		ratchetAccount = ch.HostRatchetAcct
		ratchetSeqNum = ch.HostRatchetAcctSeqNum
	}
	ratchetTx, err := buildRatchetTx(ch, ch.PendingPaymentTime, ratchetAccount, ratchetSeqNum)
	if err != nil {
		return nil, err
	}
	ratchetTxSig, err := detachedSig(ratchetTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return nil, err
	}

	guestAmount := ch.GuestAmount
	switch ch.Role {
	case Guest:
		guestAmount += ch.PendingAmountReceived

	case Host:
		guestAmount -= ch.PendingAmountReceived
	}

	var settleWithGuestSig *xdr.DecoratedSignature
	if guestAmount != 0 {
		settleWithGuestSig = new(xdr.DecoratedSignature)
		*settleWithGuestSig, err = detachedSig(&ch.CounterpartyLatestSettleWithGuestTx.Tx, seed, ch.Passphrase, ch.KeyIndex)
		if err != nil {
			return nil, err
		}
	}

	settleWithHostSig, err := detachedSig(&ch.CounterpartyLatestSettleWithHostTx.Tx, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return nil, err
	}
	m := &Message{
		ChannelID: ch.ID,
		PaymentAcceptMsg: &PaymentAcceptMsg{
			RoundNumber:                 ch.RoundNumber,
			RecipientRatchetSig:         ratchetTxSig,
			RecipientSettleWithGuestSig: settleWithGuestSig,
			RecipientSettleWithHostSig:  settleWithHostSig,
		},
		Version: version,
	}
	return m.signMsg(seed, ch.KeyIndex)
}

func sendPaymentAcceptMsg(seed []byte, ch *Channel, o Outputter) error {
	m, err := createPaymentAcceptMsg(seed, ch)
	if err != nil {
		return err
	}
	o.OutputMsg(m)
	return nil
}

func createChannelAcceptMsg(seed []byte, ch *Channel, ledgerTime time.Time) (*Message, error) {
	settleOnlyWithHostTx, err := buildSettleOnlyWithHostTx(ch)
	if err != nil {
		return nil, err
	}
	settleOnlyWithHostSig, err := detachedSig(settleOnlyWithHostTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return nil, err
	}
	ratchetTx, err := buildRatchetTx(ch, ch.FundingTime, ch.HostRatchetAcct, ch.HostRatchetAcctSeqNum)
	if err != nil {
		return nil, err
	}
	ratchetTxSig, err := detachedSig(ratchetTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return nil, err
	}
	m := &Message{
		ChannelID: ch.ID,
		ChannelAcceptMsg: &ChannelAcceptMsg{
			GuestRatchetRound1Sig:      ratchetTxSig,
			GuestSettleOnlyWithHostSig: settleOnlyWithHostSig,
		},
		Version: version,
	}
	return m.signMsg(seed, ch.KeyIndex)
}

func sendChannelAcceptMsg(seed []byte, ch *Channel, o Outputter, ledgerTime time.Time) error {
	m, err := createChannelAcceptMsg(seed, ch, ledgerTime)
	if err != nil {
		return err
	}
	o.OutputMsg(m)
	return nil
}

func createCloseMsg(seed []byte, ch *Channel) (*Message, error) {
	coopCloseTx, err := buildCooperativeCloseTx(ch)
	if err != nil {
		return nil, err
	}
	coopCloseSig, err := detachedSig(coopCloseTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return nil, err
	}
	m := &Message{
		ChannelID: ch.ID,
		CloseMsg: &CloseMsg{
			CooperativeCloseSig: coopCloseSig,
		},
		Version: version,
	}
	return m.signMsg(seed, ch.KeyIndex)
}

func sendCloseMsg(seed []byte, ch *Channel, o Outputter) error {
	m, err := createCloseMsg(seed, ch)
	if err != nil {
		return err
	}
	o.OutputMsg(m)
	return nil
}
