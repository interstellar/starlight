package fsm

import (
	"math"
	"time"

	"github.com/interstellar/starlight/math/checked"
)

// TimerTime returns the time at which a timer for the current state should fire,
// or nil if there is no timer associated with this state.
func (ch *Channel) TimerTime() (*time.Time, error) {
	var t time.Time
	switch ch.State {
	case AwaitingFunding:
		if ch.Role == Host {
			return nil, nil
		}

		// PreFundTimeout
		t = ch.FundingTime.Add(ch.MaxRoundDuration + ch.FinalityDelay)

	case ChannelProposed:
		// ChannelProposedTimeout
		t = ch.FundingTime.Add(ch.MaxRoundDuration)

	case Open, PaymentProposed, PaymentAccepted, AwaitingClose:
		// RoundTimeout
		t = ch.PaymentTime.Add(ch.MaxRoundDuration)

	case AwaitingSettlementMintime:
		// SettlementMintimeTimeout
		var err error
		t, err = ch.settlementMinTime()
		if err != nil {
			return nil, err
		}

	default:
		return nil, nil
	}

	return &t, nil
}

// settlementMinTime returns the mintime of the current settlement txs
func (ch *Channel) settlementMinTime() (time.Time, error) {
	minTime := ch.CurrentSettleWithHostTx.Tx.TimeBounds.MinTime
	if minTime > math.MaxInt64 {
		return time.Time{}, checked.ErrOverflow
	}
	return time.Unix(int64(minTime), 0), nil
}
