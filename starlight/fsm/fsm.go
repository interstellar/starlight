// Package fsm implements the state transition logic
// in the Starlight spec.
package fsm

/*

NOTE(dan):
We should never be passing transactions around on the wire,
especially if it's something that's gonna get signed.
Always just pass around the information needed to construct
the transaction. We never want to be inspecting a tx,
decide "looks good", and then sign it. Because if there's a
mistake in the transaction inspection code, like it doesn't
check that there are no extra operations other than the
ones it was expecting, it could result in theft.

*/

import (
	"encoding/json"
	"log"
	"time"

	b "github.com/stellar/go/build"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/worizon/xlm"
)

// Role is the type of a role constant.
type Role string

// Roles. A participant in a channel is either the channel's host,
// if they created the channel, or the guest, if they accepted a channel proposal.
const (
	Host  Role = "Host"
	Guest Role = "Guest"
)

// Channel represents the point-in-time state
// of a single Starlight channel.
// It is pure data,
// and designed to be serializable to and from JSON.
type Channel struct {
	ID                     string
	Role                   Role
	State, PrevState       State
	CounterpartyAddress    string // Stellar federation address of counterparty
	RemoteURL              string
	Passphrase             string
	Cursor                 string // where we are in watching escrowacct txs on the ledger
	BaseSequenceNumber     xdr.SequenceNumber
	RoundNumber            uint64
	MaxRoundDuration       time.Duration
	FinalityDelay          time.Duration
	ChannelFeerate         xlm.Amount
	HostFeerate            xlm.Amount
	FundingTime            time.Time
	FundingTimedOut        bool
	FundingTxSeqnum        xdr.SequenceNumber
	HostAmount             xlm.Amount
	GuestAmount            xlm.Amount
	TopUpAmount            xlm.Amount
	PendingAmountSent      xlm.Amount
	PendingAmountReceived  xlm.Amount
	PaymentTime            time.Time
	PendingPaymentTime     time.Time
	HostAcct               AccountID
	GuestAcct              AccountID
	EscrowAcct             AccountID
	HostRatchetAcct        AccountID
	GuestRatchetAcct       AccountID
	KeyIndex               uint32
	HostRatchetAcctSeqNum  xdr.SequenceNumber
	GuestRatchetAcctSeqNum xdr.SequenceNumber

	// Ratchet transaction from the last completed round, including the
	// counterparty's signature.
	CurrentRatchetTx xdr.TransactionEnvelope

	// Latest settlement txes for which the counterparty has a valid
	// ratchet transaction and has provided their signature.
	// TODO(debnil): Convert SettleWithHostTx to a pointer, as it is
	// only set in some channel states, both here and below.
	CounterpartyLatestSettleWithGuestTx *xdr.TransactionEnvelope
	CounterpartyLatestSettleWithHostTx  xdr.TransactionEnvelope

	// Settlement transaction from the latest completed round, including
	// the counterparty's signature. This only differs from
	// CounterPartyLatestSettlementTxes when the channel has transitioned
	// into the PaymentAccepted state, but has yet to receive the
	// PaymentCompleteMsg from the counterparty.
	CurrentSettleWithGuestTx *xdr.TransactionEnvelope
	CurrentSettleWithHostTx  xdr.TransactionEnvelope

	// In a cooperative close, the counterparty's signature
	// is included in the Channel state so a Transaction Envelope
	// containing the transaction signed by each party can be submitted.
	CounterpartyCoopCloseSig xdr.DecoratedSignature
}

// MarshalJSON implements json.Marshaler. Required for genbolt.
func (ch *Channel) MarshalJSON() ([]byte, error) {
	type t Channel
	return json.Marshal((*t)(ch))
}

// UnmarshalJSON implements json.Unmarshaler. Required for genbolt.
func (ch *Channel) UnmarshalJSON(b []byte) error {
	type t Channel
	return json.Unmarshal(b, (*t)(ch))
}

// AccountID is a Stellar account ID.
type AccountID xdr.AccountId

// MarshalText implements the TextMarshaler interface for
// the accountID type, allowing us to serialize the account
// IDs to their string addresses, rather than the default
// xdr Uint256 slice.
func (id *AccountID) MarshalText() ([]byte, error) {
	return []byte(id.Address()), nil
}

// UnmarshalText implements the TextMarshaler interface, taking
// our custom-serialized JSON for Channel objects and converting
// the string addresses back into xdr.AccountId types.
func (id *AccountID) UnmarshalText(data []byte) error {
	return id.SetAddress(string(data))
}

// MarshalBinary satisfies interface BinaryMarshaler.
func (id *AccountID) MarshalBinary() ([]byte, error) {
	return (*xdr.AccountId)(id).MarshalBinary()
}

// UnmarshalBinary satisfies interface BinaryUnmarshaler.
func (id *AccountID) UnmarshalBinary(data []byte) error {
	return (*xdr.AccountId)(id).UnmarshalBinary(data)
}

// Address produces the Stellar address string for id.
func (id *AccountID) Address() string {
	if id == nil || id.Ed25519 == nil {
		return ""
	}
	return (*xdr.AccountId)(id).Address()
}

// SetAddress sets *id to the account ID corresponding to the given Stellar address.
func (id *AccountID) SetAddress(address string) error {
	if address == "" {
		return nil
	}
	return (*xdr.AccountId)(id).SetAddress(address)
}

// Equals tells whether two Stellar account IDs are the same.
func (id *AccountID) Equals(other AccountID) bool {
	return (*xdr.AccountId)(id).Equals(xdr.AccountId(other))
}

// XDR produces the XDR form of a Stellar account ID.
func (id *AccountID) XDR() *xdr.AccountId {
	return (*xdr.AccountId)(id)
}

func (ch *Channel) roundSeqNum() xdr.SequenceNumber {
	return ch.BaseSequenceNumber + xdr.SequenceNumber(ch.RoundNumber*4)
}

func (ch *Channel) setCounterpartySettlementTxes(guestTx, hostTx *b.TransactionBuilder, guestSig, hostSig xdr.DecoratedSignature, seed []byte) error {
	var counterpartySettleWithGuestTx *xdr.TransactionEnvelope
	if guestTx != nil {
		myGuestSig, err := detachedSig(guestTx.TX, seed, ch.Passphrase, ch.KeyIndex)
		if err != nil {
			return err
		}
		counterpartySettleWithGuestTx = &xdr.TransactionEnvelope{
			Tx:         *guestTx.TX,
			Signatures: []xdr.DecoratedSignature{guestSig, myGuestSig},
		}
	}
	ch.CounterpartyLatestSettleWithGuestTx = counterpartySettleWithGuestTx
	myHostSig, err := detachedSig(hostTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return err
	}
	ch.CounterpartyLatestSettleWithHostTx = xdr.TransactionEnvelope{
		Tx:         *hostTx.TX,
		Signatures: []xdr.DecoratedSignature{hostSig, myHostSig},
	}
	return nil
}

func (ch *Channel) setLatestSettlementTxes(
	guestTx, hostTx *b.TransactionBuilder,
	guestSig *xdr.DecoratedSignature,
	hostSig xdr.DecoratedSignature,
	seed []byte,
) error {
	var latestSettleWithGuestTx *xdr.TransactionEnvelope
	if guestTx != nil {
		myGuestSig, err := detachedSig(guestTx.TX, seed, ch.Passphrase, ch.KeyIndex)
		if err != nil {
			return err
		}
		Signatures := []xdr.DecoratedSignature{myGuestSig}
		if guestSig != nil {
			Signatures = append(Signatures, *guestSig)
		}
		latestSettleWithGuestTx = &xdr.TransactionEnvelope{
			Tx:         *guestTx.TX,
			Signatures: Signatures,
		}
	}

	myHostSig, err := detachedSig(hostTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return err
	}
	latestSettleWithHostTx := xdr.TransactionEnvelope{
		Tx:         *hostTx.TX,
		Signatures: []xdr.DecoratedSignature{hostSig, myHostSig},
	}

	ch.CounterpartyLatestSettleWithGuestTx = latestSettleWithGuestTx
	ch.CurrentSettleWithGuestTx = latestSettleWithGuestTx
	ch.CounterpartyLatestSettleWithHostTx = latestSettleWithHostTx
	ch.CurrentSettleWithHostTx = latestSettleWithHostTx
	return nil
}

func (ch *Channel) signRatchetTx(ratchetTx *b.TransactionBuilder, ratchetSig xdr.DecoratedSignature, seed []byte) error {
	myRatchetSig, err := detachedSig(ratchetTx.TX, seed, ch.Passphrase, ch.KeyIndex)
	if err != nil {
		return err
	}
	ch.CurrentRatchetTx = xdr.TransactionEnvelope{
		Tx:         *ratchetTx.TX,
		Signatures: []xdr.DecoratedSignature{ratchetSig, myRatchetSig},
	}
	return nil
}

// SetupAndFundingReserveAmount reports the amount in lumens needed to set up and fund the channel.
func (ch *Channel) SetupAndFundingReserveAmount() xlm.Amount {
	var result xlm.Amount
	result += setupMinBalanceAmount()
	result += ch.setupFeeAmount()
	result += ch.totalFundingTxAmount()
	return result
}

func setupMinBalanceAmount() xlm.Amount {
	// Escrow, host ratchet, guest ratchet have min balance of 1 XLM.
	return 3 * xlm.Lumen
}

func (ch *Channel) setupFeeAmount() xlm.Amount {
	// Escrow, host ratchet, guest ratchet have same fees.
	return 3 * ch.HostFeerate
}

func (ch *Channel) totalFundingTxAmount() xlm.Amount {
	// Combine funding tx balance, funding tx fees, and fee-amts for funded channel accts.
	var result xlm.Amount
	result += ch.fundingBalanceAmount()
	result += ch.fundingFeeAmount()
	result += ch.fundedAcctsTxFeeAmount()
	return result
}

func (ch *Channel) fundingBalanceAmount() xlm.Amount {
	// Guest ratchet has 2 additional signers, escrow and host ratchet 1 each.
	// Each additional signer adds .5 Lumen to the minimum reserve balance.
	return ch.HostAmount + 2*xlm.Lumen
}

func (ch *Channel) fundingFeeAmount() xlm.Amount {
	// Funding tx has 7 ops, from Host account.
	return 7 * ch.HostFeerate
}

func (ch *Channel) fundedAcctsTxFeeAmount() xlm.Amount {
	// Escrow fees are 8 * feerate XLM and ratchet accounts are 1 * feerate XLM each.
	return 10 * ch.ChannelFeerate
}

func isSetupState(state State) bool {
	switch state {
	case Start, SettingUp, ChannelProposed, AwaitingFunding:
		return true
	}
	return false
}

func isForceCloseState(state State) bool {
	switch state {
	case AwaitingRatchet, AwaitingSettlementMintime, AwaitingSettlement:
		return true
	}
	return false
}

func (u *Updater) setForceCloseState() error {
	// if we're already in a force close state, do nothing
	switch u.C.State {
	case AwaitingRatchet, AwaitingSettlement, AwaitingSettlementMintime, Closed:
		return nil
	}
	log.Print("entering force close")
	if u.C.Role == Guest && u.C.GuestAmount == 0 {
		// doesn't care about settlement
		// and may not even have a ratchet tx
		return u.transitionTo(Closed)
	}
	return u.transitionTo(AwaitingRatchet)
}
