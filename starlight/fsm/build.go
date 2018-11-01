package fsm

import (
	"math"
	"time"

	b "github.com/stellar/go/build"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/math/checked"
	"github.com/interstellar/starlight/worizon/xlm"
)

func (ch *Channel) buildWalletTx(seqnum xdr.SequenceNumber, m ...b.TransactionMutator) (*b.TransactionBuilder, error) {
	return ch.buildTx(ch.HostAcct, seqnum, ch.HostFeerate, m...)
}

func (ch *Channel) buildEscrowTx(seqnum xdr.SequenceNumber, m ...b.TransactionMutator) (*b.TransactionBuilder, error) {
	return ch.buildTx(ch.EscrowAcct, seqnum, ch.ChannelFeerate, m...)
}

func (ch *Channel) buildTx(acct AccountID, seqnum xdr.SequenceNumber, basefee xlm.Amount, m ...b.TransactionMutator) (*b.TransactionBuilder, error) {
	args := []b.TransactionMutator{
		b.Network{Passphrase: ch.Passphrase},
		b.SourceAccount{AddressOrSeed: acct.Address()},
		b.Sequence{Sequence: uint64(seqnum)},
		b.BaseFee{Amount: uint64(basefee)},
	}
	args = append(args, m...)
	return b.Transaction(args...)
}

/* *** */

func buildSetupAccountTx(ch *Channel, account AccountID, seqnum xdr.SequenceNumber) (*b.TransactionBuilder, error) {
	tb, err := ch.buildWalletTx(
		seqnum,
		b.CreateAccount(
			b.Destination{AddressOrSeed: account.Address()},
			b.NativeAmount{Amount: xlm.Lumen.HorizonString()},
		),
	)
	return tb, err
}

func buildSettleOnlyWithHostTx(ch *Channel, paymentTime time.Time) (*b.TransactionBuilder, error) {
	minTime := uint64(paymentTime.Add(2 * ch.FinalityDelay).Add(ch.MaxRoundDuration).Unix())
	if minTime > math.MaxInt64 {
		return nil, checked.ErrOverflow
	}
	return ch.buildEscrowTx(
		ch.roundSeqNum()+2,
		b.Timebounds{MinTime: minTime},
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.GuestRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.HostRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
	)
}

func buildRatchetTx(ch *Channel, ledgerTime time.Time, acct AccountID, seqnum xdr.SequenceNumber) (*b.TransactionBuilder, error) {
	maxTime := uint64(ledgerTime.Add(ch.FinalityDelay).Add(ch.MaxRoundDuration).Unix())
	if maxTime > math.MaxInt64 {
		return nil, checked.ErrOverflow
	}
	return ch.buildTx(
		acct,
		seqnum+1,
		ch.ChannelFeerate,
		b.Timebounds{MaxTime: maxTime},
		b.BumpSequence(
			b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
			b.BumpTo(ch.roundSeqNum()+1),
		),
	)
}

func buildFundingTx(ch *Channel, h *WalletAcct) (*b.TransactionBuilder, error) {
	maxTime := uint64(ch.FundingTime.Add(ch.MaxRoundDuration).Add(ch.FinalityDelay).Unix())
	if maxTime > math.MaxInt64 {
		return nil, checked.ErrOverflow
	}
	tb, err := ch.buildWalletTx(
		h.Seqnum,
		b.Timebounds{MaxTime: maxTime},
		b.Payment(
			b.SourceAccount{AddressOrSeed: ch.HostAcct.Address()},
			b.Destination{AddressOrSeed: ch.EscrowAcct.Address()},
			b.NativeAmount{Amount: (ch.HostAmount + 500*xlm.Millilumen + 8*ch.ChannelFeerate).HorizonString()},
		),
		b.SetOptions(
			b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
			b.SetLowThreshold(2),
			b.SetMediumThreshold(2),
			b.SetHighThreshold(2),
			b.AddSigner(ch.GuestAcct.Address(), 1),
		),
		b.Payment(
			b.SourceAccount{AddressOrSeed: ch.HostAcct.Address()},
			b.Destination{AddressOrSeed: ch.GuestRatchetAcct.Address()},
			b.NativeAmount{Amount: (xlm.Lumen + ch.ChannelFeerate).HorizonString()},
		),
		b.SetOptions(
			b.SourceAccount{AddressOrSeed: ch.GuestRatchetAcct.Address()},
			b.MasterWeight(0),
			b.SetLowThreshold(2),
			b.SetMediumThreshold(2),
			b.SetHighThreshold(2),
			b.AddSigner(ch.GuestAcct.Address(), 1),
		),
		b.SetOptions(
			b.SourceAccount{AddressOrSeed: ch.GuestRatchetAcct.Address()},
			b.AddSigner(ch.EscrowAcct.Address(), 1),
		),
		b.Payment(
			b.SourceAccount{AddressOrSeed: ch.HostAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostRatchetAcct.Address()},
			b.NativeAmount{Amount: (500*xlm.Millilumen + ch.ChannelFeerate).HorizonString()},
		),
		b.SetOptions(
			b.SourceAccount{AddressOrSeed: ch.HostRatchetAcct.Address()},
			b.MasterWeight(0),
			b.AddSigner(ch.EscrowAcct.Address(), 1),
		),
	)
	return tb, err
}

func buildSettleWithGuestTx(ch *Channel, paymentTime time.Time) (*b.TransactionBuilder, error) {
	minTime := uint64(paymentTime.Add(2 * ch.FinalityDelay).Add(ch.MaxRoundDuration).Unix())
	if minTime > math.MaxInt64 {
		return nil, checked.ErrOverflow
	}
	return ch.buildEscrowTx(
		ch.roundSeqNum()+2,
		b.Timebounds{MinTime: minTime},
		b.Payment(
			b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
			b.Destination{AddressOrSeed: ch.GuestAcct.Address()},
			b.NativeAmount{Amount: ch.GuestAmount.HorizonString()},
		),
	)
}

func buildSettleWithHostTx(ch *Channel, paymentTime time.Time) (*b.TransactionBuilder, error) {
	minTime := uint64(paymentTime.Add(2 * ch.FinalityDelay).Add(ch.MaxRoundDuration).Unix())
	if minTime > math.MaxInt64 {
		return nil, checked.ErrOverflow
	}
	return ch.buildEscrowTx(
		ch.roundSeqNum()+3,
		b.Timebounds{MinTime: minTime},
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.GuestRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.HostRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
	)
}

func buildCooperativeCloseTx(ch *Channel) (*b.TransactionBuilder, error) {
	tb, err := ch.buildEscrowTx(ch.BaseSequenceNumber + 1)
	if err != nil {
		return nil, err
	}
	if ch.GuestAmount > 0 {
		err = tb.Mutate(
			b.Payment(
				b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
				b.Destination{AddressOrSeed: ch.GuestAcct.Address()},
				b.NativeAmount{Amount: ch.GuestAmount.HorizonString()},
			),
		)
		if err != nil {
			return nil, err
		}
	}
	err = tb.Mutate(
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.EscrowAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.GuestRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
		b.AccountMerge(
			b.SourceAccount{AddressOrSeed: ch.HostRatchetAcct.Address()},
			b.Destination{AddressOrSeed: ch.HostAcct.Address()},
		),
	)
	if err != nil {
		return nil, err
	}
	err = tb.Mutate(b.Defaults{})
	return tb, err
}

func buildCleanupTx(ch *Channel, h *WalletAcct) (*b.TransactionBuilder, error) {
	var seqnum xdr.SequenceNumber
	if ch.FundingTimedOut {
		seqnum = ch.FundingTxSeqnum
	} else {
		seqnum = h.Seqnum
	}
	return ch.buildWalletTx(
		seqnum,
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
}

func buildTopUpTx(ch *Channel, h *WalletAcct) (*b.TransactionBuilder, error) {
	return ch.buildWalletTx(
		h.Seqnum,
		b.Payment(
			b.SourceAccount{AddressOrSeed: ch.HostAcct.Address()},
			b.Destination{AddressOrSeed: ch.EscrowAcct.Address()},
			b.NativeAmount{Amount: ch.TopUpAmount.HorizonString()},
		),
	)
}
