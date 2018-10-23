package fsm

import (
	"time"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/worizon/xlm"
)

type UserCommand string

const (
	CreateChannel UserCommand = "CreateChannel"
	CleanUp       UserCommand = "CleanUp"
	CloseChannel  UserCommand = "CloseChannel"
	TopUp         UserCommand = "TopUp"
	ChannelPay    UserCommand = "ChannelPay"
	ForceClose    UserCommand = "ForceClose"
	Pay           UserCommand = "Pay"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

// Command represents a user request,
// such as "force close" or "send payment".
type Command struct {
	UserCommand UserCommand
	Amount      xlm.Amount // for TopUp, ChannelPay, or Pay
	Time        time.Time
	Recipient   string // for Pay
}

var commandFuncs = map[UserCommand]func(*Command, *Updater) error{
	CreateChannel: createChannelFn,
	CleanUp:       cleanUpFn,
	CloseChannel:  closeChannelFn,
	TopUp:         topUpFn,
	ChannelPay:    channelPayFn,
	ForceClose:    forceCloseFn,
}

func createChannelFn(_ *Command, u *Updater) error {
	if u.C.State != Start {
		return errors.Wrapf(ErrUnexpectedState, "got %s, want %s", u.C.State, Start)
	}
	return u.transitionTo(SettingUp)
}
func cleanUpFn(_ *Command, u *Updater) error {
	if u.C.State != ChannelProposed {
		return errors.Wrapf(ErrUnexpectedState, "got %s, want %s", u.C.State, ChannelProposed)
	}
	u.H.Balance += u.C.SetupAndFundingReserveAmount()
	u.H.Seqnum++
	return u.transitionTo(AwaitingCleanup)
}

func closeChannelFn(_ *Command, u *Updater) error {
	if u.C.State != Open {
		return errors.Wrapf(ErrUnexpectedState, "got %s, want %s", u.C.State, Open)
	}
	return u.transitionTo(AwaitingClose)
}

func topUpFn(c *Command, u *Updater) error {
	if u.C.State != Open {
		return errors.Wrapf(ErrUnexpectedState, "got %s, want %s", u.C.State, Open)
	}
	if u.C.Role != Host {
		return errors.New("only host can top up")
	}
	if u.C.TopUpAmount != 0 {
		return errors.New("top-up currently being submitted")
	}
	if c.Amount > u.H.Balance {
		return errors.Wrapf(ErrInsufficientFunds, "balance %d", u.C.HostAmount)
	}
	u.C.TopUpAmount = c.Amount

	u.H.Balance -= c.Amount
	u.H.Balance -= u.C.HostFeerate

	u.H.Seqnum++
	return u.transitionTo(Open)
}

func channelPayFn(c *Command, u *Updater) error {
	if u.C.State != Open {
		return errors.Wrapf(ErrUnexpectedState, "got %s, want %s", u.C.State, Open)
	}
	u.C.PendingAmountSent = c.Amount
	if u.C.PaymentTime.After(c.Time) {
		u.C.PendingPaymentTime = u.C.PaymentTime
	} else {
		u.C.PendingPaymentTime = c.Time
	}
	switch u.C.Role {
	case Guest:
		if u.C.GuestAmount < c.Amount {
			return errors.Wrapf(ErrInsufficientFunds, "balance %d", u.C.GuestAmount)
		}
	case Host:
		if u.C.HostAmount < c.Amount {
			return errors.Wrapf(ErrInsufficientFunds, "balance %d", u.C.HostAmount)
		}
	}
	u.C.RoundNumber++
	return u.transitionTo(PaymentProposed)
}

func forceCloseFn(_ *Command, u *Updater) error {
	if isSetupState(u.C.State) || isForceCloseState(u.C.State) {
		return errors.Wrapf(ErrUnexpectedState, "got %s, want non-starting, non-force close state", u.C.State)
	}
	return u.setForceCloseState()
}
