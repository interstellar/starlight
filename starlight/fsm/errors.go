package fsm

import "errors"

// Defines the FSM package errors
var (
	// Command errors
	ErrInsufficientFunds = errors.New("insufficient funds")
	errTopUpInProgress   = errors.New("top-up currently being submitted")
	errUnexpectedRole    = errors.New("unexpected role")

	// Message errors
	ErrChannelExists            = errors.New("received channel propose message for channel that already exists")
	ErrInvalidVersion           = errors.New("invalid version number")
	ErrUnusedSettleWithGuestSig = errors.New("unused settle with guest sig")

	// Tx errors
	errRatchetTxFailed = errors.New("ratchet tx failed")
	ErrUnexpectedState = errors.New("unexpected state")
	errNoMatch         = errors.New("did not recognize transaction")
	errNoSeed          = errors.New("fsm: seed required")
)
