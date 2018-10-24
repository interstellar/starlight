package fsm

// State is the type of a channel-state constant.
type State string

// Channel-state constants.
const (
	// Start indicates a channel that does not (yet) exist.
	//
	// Note that no state transition enters the state Start,
	// so if a Channel has this state after running an Update
	// function in this package, the channel is invalid.
	Start State = ""

	Closed State = "Closed"

	AwaitingCleanup           State = "AwaitingCleanup"
	AwaitingClose             State = "AwaitingClose"
	AwaitingFunding           State = "AwaitingFunding"
	AwaitingPaymentMerge      State = "AwaitingPaymentMerge"
	AwaitingRatchet           State = "AwaitingRatchet"
	AwaitingSettlement        State = "AwaitingSettlement"
	AwaitingSettlementMintime State = "AwaitingSettlementMintime"
	ChannelProposed           State = "ChannelProposed"
	Open                      State = "Open"
	PaymentAccepted           State = "PaymentAccepted"
	PaymentProposed           State = "PaymentProposed"
	SettingUp                 State = "SettingUp"
)

func (u *Updater) transitionTo(newState State) error {
	u.C.State, u.C.PrevState = newState, u.C.State

	switch newState {
	case AwaitingCleanup:
		return publishCleanupTx(u.Seed, u.C, u.O, u.H)

	case AwaitingClose:
		if u.C.CounterpartyCoopCloseSig.Signature != nil {
			return publishCoopCloseTx(u.Seed, u.C, u.O, u.H)
		}
		return sendCloseMsg(u.Seed, u.C, u.O)

	case AwaitingFunding:
		switch u.C.Role {
		case Guest:
			if u.C.PrevState != Start {
				return errUnexpectedState
			}
			err := sendChannelAcceptMsg(u.Seed, u.C, u.O, u.LedgerTime)
			if err != nil {
				return err
			}
			// timer gets set

		case Host:
			return publishFundingTx(u.Seed, u.C, u.O, u.H)
		}

	case AwaitingPaymentMerge:
		return nil // nothing to do

	case AwaitingRatchet:
		u.O.OutputTx(u.C.CurrentRatchetTx)

	case AwaitingSettlement:
		if u.C.CurrentSettleWithGuestTx != nil {
			u.O.OutputTx(*u.C.CurrentSettleWithGuestTx)
		}
		u.O.OutputTx(u.C.CurrentSettleWithHostTx)

	case AwaitingSettlementMintime:
		// timer gets set

	case ChannelProposed:
		return sendChannelProposeMsg(u.Seed, u.C, u.O, u.H)

	case Closed:
		return nil // nothing to do

	case Open:
		switch u.C.PrevState {
		case Open:
			if u.C.TopUpAmount != 0 && u.C.Role == Host {
				return publishTopUpTx(u.Seed, u.C, u.O, u.H)
			}

		case PaymentProposed:
			return sendPaymentCompleteMsg(u.Seed, u.C, u.O)
		}

	case PaymentAccepted:
		return sendPaymentAcceptMsg(u.Seed, u.C, u.O)

	case PaymentProposed:
		return sendPaymentProposeMsg(u.Seed, u.C, u.O)

	case SettingUp:
		return publishSetupAccountTxes(u.Seed, u.C, u.O, u.H)
	}
	return nil
}
