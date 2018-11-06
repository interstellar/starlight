package starlighttest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/internal/update"
	"github.com/interstellar/starlight/starlight/log"
	"github.com/interstellar/starlight/worizon"
	"github.com/interstellar/starlight/worizon/xlm"
)

const (
	friendbotAmount = 10000 * xlm.Lumen
	channelFeerate  = 10 * xlm.Millilumen
	hostFeerate     = 100 * xlm.Stroop
)

type step struct {
	name           string
	agent          *Starlightd
	path           string
	body           string
	wantCode       int
	injectChanID   bool
	injectHostAcct bool
	update         *update.Update
	outOfOrder     bool
	checkLedger    bool
	walletDelta    xlm.Amount
	hostDelta      xlm.Amount
	guestDelta     xlm.Amount
}

// WalletPaySelf executes a do-wallet-pay API action.
func WalletPaySelf(ctx context.Context, self *Starlightd) error {
	steps := []step{
		{
			name:  "do wallet pay",
			agent: self,
			path:  "/api/do-wallet-pay",
			body: `{
				"Dest":"%s",
				"Amount": 1000000000
			}`,
			injectHostAcct: true,
		},
	}
	var channelID string
	for _, s := range steps {
		err := handleStep(ctx, s, &channelID)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateChannel executes the API steps for creating a channel.
func CreateChannel(ctx context.Context, guest, host *Starlightd, channelFundingAmount xlm.Amount) (string, error) {
	steps := channelCreationSteps(guest, host, 0, 0, channelFundingAmount)
	var channelID string
	for _, s := range steps {
		err := handleStep(ctx, s, &channelID)
		if err != nil {
			return "", err
		}
	}
	return channelID, nil
}

// HostChannelPayGuest executes the API steps for a host-to-guest channel payment.
func HostChannelPayGuest(ctx context.Context, guest, host *Starlightd, channelID string, payment xlm.Amount) error {
	steps := hostChannelPayGuestSteps(guest, host, payment)
	for _, s := range steps {
		err := handleStep(ctx, s, &channelID)
		if err != nil {
			return err
		}
	}
	return nil
}

// HostTopUp executes the API steps for a host top-up.
func HostTopUp(ctx context.Context, guest, host *Starlightd, channelID string, topUpAmount xlm.Amount) error {
	steps := hostTopUpSteps(guest, host, topUpAmount)
	for _, s := range steps {
		err := handleStep(ctx, s, &channelID)
		if err != nil {
			return err
		}
	}
	return nil
}

func channelCreationSteps(guest, host *Starlightd, maxRoundDurMins, finalityDelayMins int, channelFundingAmount xlm.Amount) []step {
	// WARNING: this software is not compatible with Stellar mainnet.
	return []step{
		{
			name:  "guest config init",
			agent: guest,
			path:  "/api/config-init",
			// WARNING: this software is not compatible with Stellar mainnet.
			body: fmt.Sprintf(`
			{
				"Username":"guest",
				"Password":"password",
				"DemoServer":true,
				"HorizonURL":"%s",
				"KeepAlive":false,
				"MaxRoundDurMins": %d,
				"FinalityDelayMins": %d
			}`, *HorizonURL, maxRoundDurMins, finalityDelayMins),
		}, {
			name:  "guest config init update",
			agent: guest,
			update: &update.Update{
				Type:      update.InitType,
				UpdateNum: 1,
			},
		}, {
			name:  "host config init",
			agent: host,
			path:  "/api/config-init",
			// WARNING: this software is not compatible with Stellar mainnet.
			body: fmt.Sprintf(`
			{
				"Username":"host",
				"Password":"password",
				"DemoServer":true,
				"HorizonURL":"%s",
				"KeepAlive":false,
				"MaxRoundDurMins": %d,
				"FinalityDelayMins": %d,
				"HostFeerate": %d,
				"ChannelFeerate":%d
			}`, *HorizonURL, maxRoundDurMins, finalityDelayMins, hostFeerate, channelFeerate),
		}, {
			name:  "host config init update",
			agent: host,
			update: &update.Update{
				Type:      update.InitType,
				UpdateNum: 1,
			},
		}, {
			name:  "guest wallet funding update",
			agent: guest,
			update: &update.Update{
				Type:      update.AccountType,
				UpdateNum: 2,
			},
			walletDelta: friendbotAmount - hostFeerate,
			checkLedger: true,
		}, {
			name:  "host wallet funding update",
			agent: host,
			path:  "/api/updates",
			body:  `{"From": 2}`,
			update: &update.Update{
				Type:      update.AccountType,
				UpdateNum: 2,
			},
			walletDelta: friendbotAmount - hostFeerate,
			checkLedger: true,
		}, {
			name:  "host create channel with guest",
			agent: host,
			path:  "/api/do-create-channel",
			body: fmt.Sprintf(`{
				"GuestAddr": "guest*%s",
				"HostAmount": %d
			}`, guest.address, channelFundingAmount),
		}, {
			name:  "host channel creation setting up update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.SettingUp,
				},
			},
			// Host balance should decrease by:
			// channelFundingAmount
			// 3 * Lumen + 3 * hostFee to create the three accounts
			// 7 * hostFee in funding tx fees for 7 operations
			// 0.5*Lumen + 8 * channelFee in initial funding for the escrow account
			// 1 * Lumen + 1 * channelFee in initial funding for the guest ratchet
			// 0.5*Lumen + 1 * channelFee in initial funding for the host ratchet account
			walletDelta: -(channelFundingAmount + 5*xlm.Lumen + 10*hostFeerate + 10*channelFeerate),
			// Host amount should increase by channelFundingAmount
			hostDelta: channelFundingAmount,
		}, {
			name:  "host channel creation channel proposed update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.ChannelProposed,
				},
			},
		}, {
			name:  "guest channel creation awaiting funding update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingFunding,
				},
			},
			hostDelta: channelFundingAmount,
		}, {
			name:  "guest channel creation awaiting funding tx update",
			agent: guest,
			update: &update.Update{
				Type:    update.ChannelType,
				InputTx: &worizon.Tx{},
				Channel: &fsm.Channel{
					State: fsm.AwaitingFunding,
				},
			},
		}, {
			name:  "host channel creation awaiting funding update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingFunding,
				},
			},
		}, {
			name:  "guest channel creation open update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
		}, {
			name:  "host channel creation open update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
		},
	}
}

func hostChannelPayGuestSteps(guest, host *Starlightd, payment xlm.Amount) []step {
	return []step{
		{
			name:  "host channel pay to guest",
			agent: host,
			path:  "/api/do-command",
			body: fmt.Sprintf(`
			{
				"ChannelID": "%%s",
				"Command": {
					"Name": "ChannelPay",
					"Amount": %d,
					"Time": "2018-10-02T10:26:43.511Z"
				}
			}`, payment),
			injectChanID: true,
		}, {
			name:  "host channel pay to guest payment proposed update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.PaymentProposed,
				},
			},
		}, {
			name:  "host channel pay to guest payment accept update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.PaymentAccepted,
				},
			},
		}, {
			name:  "host channel pay to guest payment host channel open update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
			hostDelta:  -payment,
			guestDelta: payment,
		}, {
			name:  "host channel pay to guest payment guest channel open update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
			hostDelta:  -payment,
			guestDelta: payment,
		},
	}
}

func guestCoopCloseSteps(guest, host *Starlightd, channelFundingAmount, settleWithGuestAmount xlm.Amount) []step {
	var guestFee xlm.Amount
	if settleWithGuestAmount != 0 {
		// Add additional channelFeerate for additional operation cost to
		// settle with guest
		guestFee = channelFeerate
	}
	return []step{
		{
			name:  "guest cooperative close",
			agent: guest,
			path:  "/api/do-command",
			body: `
				{
					"ChannelID": "%s",
					"Command": {
						"Name": "CloseChannel",
						"Time": "2018-10-02T10:26:43.511Z"
					}
				}`,
			injectChanID: true,
		}, {
			name:  "guest cooperative close awaiting close update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingClose,
				},
			},
		}, {
			name:  "host cooperative close awaiting close update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingClose,
				},
			},
		}, {
			name:  "host cooperative close channel closed update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Closed,
				},
			},
			outOfOrder: true,
		}, {
			name:  "host cooperative close coop close tx merge escrow account",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// Balance should increase by:
			// channelFundingAmount, initial funding
			// 1 Lumen minimum escrow account balance
			// 0.5 Lumen multisig escrow account balance
			// 8 * channelFeerate initial funding
			// less 3 * channelFeerate coop close tx fees (one added above if settling with guest)
			// less payment amount
			walletDelta: channelFundingAmount + 1*xlm.Lumen + 500*xlm.Millilumen + 8*channelFeerate - 3*channelFeerate - settleWithGuestAmount - guestFee,
		}, {
			name:  "host cooperative close coop close tx merge guest ratchet account",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// Balance should increase by
			// 2 Lumens initial funding + channelFeerate
			walletDelta: 2*xlm.Lumen + channelFeerate,
		}, {
			name:  "host cooperative close coop close tx merge host ratchet account",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// Balance should increase by
			// 1.5 Lumens initial funding + channelFeerate
			walletDelta: 1*xlm.Lumen + 500*xlm.Millilumen + channelFeerate,
			// Account should be up to date with ledger at this point
			checkLedger: true,
		}, {
			name:  "guest cooperative close channel closed update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Closed,
				},
			},
		},
	}
}

func hostForceCloseStepsNoGuestBalance(guest, host *Starlightd, channelFundingAmount xlm.Amount) []step {
	return []step{
		{
			name:  "host force close command",
			agent: host,
			path:  "/api/do-command",
			body: `{
				"ChannelID": "%s",
				"Command": {
					"Name": "ForceClose"
				}
			}`,
			injectChanID: true,
		}, {
			// host transitions to AwaitingRatchet state,
			// outputs current ratchet tx
			name:  "host force close state transition awaiting ratchet",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingRatchet,
				},
			},
		}, {
			// host sees current ratchet tx hit the ledger,
			// transitions to awaiting settlement mintime
			name:  "host force close state transition awaiting settlement mintime",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingSettlementMintime,
				},
			},
		}, {
			// guest sees current ratchet tx hit the ledger,
			// transitions to closed
			name:  "guest force close state transition closed",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Closed,
				},
			},
		}, {
			// host waits until settlement mintime, then transitions to
			// awaiting settlement
			name:  "host force close state transition awaiting settlement",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingSettlement,
				},
			},
		}, {
			// host sees current settlement tx hit the ledger, transitions to
			// closed state
			name:  "host force close state transition closed",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Closed,
				},
			},
			walletDelta: 0,
			outOfOrder:  true,
		}, {
			name:  "host force close merge escrow account",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// Balance should increase by:
			// channelFundingAmount, initial funding
			// 1 Lumen minimum escrow account balance
			// 0.5 Lumen multisig escrow account balance
			// 8 * channelFeerate initial funding
			// less 3 * channelFeerate ratchet / settlement tx fees
			walletDelta: channelFundingAmount + 1*xlm.Lumen + 500*xlm.Millilumen + 8*channelFeerate - 3*channelFeerate,
		}, {
			name:  "host force close merge guest ratchet",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// merge back 2 Lumens guest ratchet balance
			walletDelta: 2*xlm.Lumen + channelFeerate,
		}, {
			name:  "host force close merge host ratchet",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// merge back 1.5 Lumens host ratchet balance
			walletDelta: 1*xlm.Lumen + 500*xlm.Millilumen,
			checkLedger: true,
		},
	}
}

func hostForceCloseStepsSettleWithGuest(guest, host *Starlightd, channelFundingAmount, payment xlm.Amount) []step {
	return []step{
		{
			name:  "host force close command",
			agent: host,
			path:  "/api/do-command",
			body: `{
				"ChannelID": "%s",
				"Command": {
					"Name": "ForceClose"
				}
			}`,
			injectChanID: true,
		}, {
			// host sees current ratchet tx hit the ledger,
			// transitions to awaiting settlement mintime
			name:  "host force close state transition awaiting settlement mintime",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingSettlementMintime,
				},
			},
		}, {
			// guest sees current ratchet tx hit the ledger,
			// transitions to awaiting settlement mintime
			name:  "guest force close state transition closed",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingSettlementMintime,
				},
			},
		}, {
			// host waits until settlement mintime, then transitions to
			// awaiting settlement
			name:  "host force close state transition awaiting settlement",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingSettlement,
				},
			},
		}, {
			// guest waits until settlement mintime, then transitions to
			// awaiting settlement
			name:  "guest force close state transition awaiting settlement",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingSettlement,
				},
			},
		}, {
			// host sees current settlement tx hit the ledger, transitions to
			// closed state
			name:  "host force close state transition closed",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Closed,
				},
			},
			// state closed update could come before or after merge tx updates
			outOfOrder: true,
		}, {
			// guest sees current settlement tx hit the ledger, transitions to
			// closed state
			name:  "guest force close state transition closed",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Closed,
				},
			},
		}, {
			name:  "host force close merge escrow account",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// Balance should increase by:
			// channelFundingAmount initial funding
			// 1 Lumen minimum escrow account balance
			// 0.5 Lumen multisig escrow account balance
			// 8 * channelFeerate initial funding
			// less 4 * channelFeerate ratchet / settlement tx fees
			// less 1000 stroops sent to Guest
			walletDelta: channelFundingAmount + 1*xlm.Lumen + 500*xlm.Millilumen + 8*channelFeerate - 4*channelFeerate - payment,
		}, {
			name:  "host force close merge guest ratchet",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// merge back 2 Lumens guest ratchet balance
			walletDelta: 2 * xlm.Lumen,
		}, {
			name:  "host force close merge host ratchet",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			// merge back 1.5 Lumens host ratchet balance
			walletDelta: 1*xlm.Lumen + 500*xlm.Millilumen + channelFeerate,
			checkLedger: true,
		},
	}
}

func hostTopUpSteps(guest, host *Starlightd, topUpAmount xlm.Amount) []step {
	return []step{
		{
			name:  "host top-up command",
			agent: host,
			path:  "/api/do-command",
			body: fmt.Sprintf(`{
				"ChannelID": "%%s",
				"Command": {
					"Name": "TopUp",
					"Amount":%d
				}
			}`, topUpAmount),
			injectChanID: true,
		}, {
			name:  "host top-up channel update",
			agent: host,
			update: &update.Update{
				Type:    update.ChannelType,
				InputTx: &worizon.Tx{},
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
			// balance decreases by top-up amount + host fee rate
			walletDelta: -(topUpAmount + hostFeerate),
			hostDelta:   topUpAmount,
		}, {
			name:  "guest top-up tx update",
			agent: guest,
			update: &update.Update{
				Type:    update.ChannelType,
				InputTx: &worizon.Tx{},
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
			hostDelta: topUpAmount,
		},
	}
}

// mergingPaymentSteps sends payments from Guest and Host that, if one of the agents is not
// receiving messages, will eventually create a payment merge state when the agent comes
// back online.
func mergingPaymentSteps(guest, host *Starlightd, guestPayment, hostPayment xlm.Amount) []step {
	return []step{
		{
			name:  "guest channel pay host",
			agent: guest,
			path:  "/api/do-command",
			body: fmt.Sprintf(`
			{
				"ChannelID": "%%s",
				"Command": {
					"Name": "ChannelPay",
					"Amount": %d,
					"Time": "2018-10-02T10:26:43.511Z"
				}
			}`, guestPayment),
			injectChanID: true,
		}, {
			name:  "host channel pay guest",
			agent: host,
			path:  "/api/do-command",
			body: fmt.Sprintf(`
			{
				"ChannelID": "%%s",
				"Command": {
					"Name": "ChannelPay",
					"Amount": %d,
					"Time": "2018-10-02T10:26:43.511Z"
				}
			}`, hostPayment),
			injectChanID: true,
		},
	}
}

// paymentMergeResolutionSteps represents the state transitions that Guest and Host go through to merge
// the conflict payment, sending a merged payment from Host to Guest.
func paymentMergeResolutionSteps(guest, host *Starlightd, mergePayment xlm.Amount) []step {
	return []step{
		{
			name:  "host channel pay guest payment proposed update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.PaymentProposed,
				},
			},
		}, {
			name:  "guest channel pay host payment proposed update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.PaymentProposed,
				},
			},
		}, {
			name:  "guest channel pay awaiting payment merge update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingPaymentMerge,
				},
			},
		}, {
			name:  "host channel pay payment proposed update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.PaymentProposed,
				},
			},
		}, {
			name:  "guest channel pay payment accepted update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.PaymentAccepted,
				},
			},
		}, {
			name:  "host channel pay payment complete state open update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
			hostDelta:  -mergePayment,
			guestDelta: mergePayment,
		}, {
			name:  "guest channel pay open update",
			agent: guest,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Open,
				},
			},
			hostDelta:  -mergePayment,
			guestDelta: mergePayment,
		},
	}
}

func logoutSteps(guest, host *Starlightd) []step {
	return []step{
		{
			name:  "guest logout command",
			agent: guest,
			path:  "/api/logout",
		}, {
			name:     "guest logout update",
			agent:    guest,
			path:     "/api/updates",
			body:     `{"From":1}`,
			wantCode: 401,
		},
	}
}

func loginSteps(guest, host *Starlightd) []step {
	return []step{
		{
			name:  "guest login command",
			agent: guest,
			path:  "/api/login",
			body:  `{"Username":"guest","Password":"password"}`,
		}, {
			name:  "guest login update",
			agent: guest,
			path:  "/api/updates",
			body:  `{"From":2}`,
		},
	}
}

func cleanupSteps(guest *httptest.Server, host *Starlightd, maxRoundDurMins, finalityDelayMins int, channelFundingAmount xlm.Amount) []step {
	return []step{
		{
			name:  "host config init",
			agent: host,
			path:  "/api/config-init",
			// WARNING: this software is not compatible with Stellar mainnet.
			body: fmt.Sprintf(`
			{
				"Username":"host",
				"Password":"password",
				"DemoServer":true,
				"HorizonURL":"%s",
				"KeepAlive":false,
				"MaxRoundDurMins": %d,
				"FinalityDelayMins": %d,
				"HostFeerate": %d,
				"ChannelFeerate":%d
			}`, *HorizonURL, maxRoundDurMins, finalityDelayMins, hostFeerate, channelFeerate),
		}, {
			name:  "host config init update",
			agent: host,
			update: &update.Update{
				Type:      update.InitType,
				UpdateNum: 1,
			},
		},
		{
			name:  "host wallet funding update",
			agent: host,
			path:  "/api/updates",
			body:  `{"From": 2}`,
			update: &update.Update{
				Type:      update.AccountType,
				UpdateNum: 2,
			},
			walletDelta: friendbotAmount - hostFeerate,
			checkLedger: true,
		}, {
			name:  "host create channel with guest",
			agent: host,
			path:  "/api/do-create-channel",
			body: fmt.Sprintf(`{
				"GuestAddr": "guest*%s",
				"HostAmount": %d
			}`, strings.TrimPrefix(guest.URL, "http://"), channelFundingAmount),
		}, {
			name:  "host channel creation setting up update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.SettingUp,
				},
			},
			// Host balance should decrease by:
			// channelFundingAmount
			// 3 * Lumen + 3 * hostFee to create the three accounts
			// 7 * hostFee in funding tx fees for 7 operations
			// 0.5*Lumen + 8 * channelFee in initial funding for the escrow account
			// 1 * Lumen + 1 * channelFee in initial funding for the guest ratchet
			// 0.5*Lumen + 1 * channelFee in initial funding for the host ratchet account
			walletDelta: -(channelFundingAmount + 5*xlm.Lumen + 10*hostFeerate + 10*channelFeerate),
		},
		{
			name:  "host channel creation channel proposed update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.ChannelProposed,
				},
			},
		},
		{
			name:  "host cleanup command",
			agent: host,
			path:  "/api/do-command",
			body: `{
				"ChannelID": "%s",
				"Command": {
					"Name": "CleanUp"
				}
			}`,
			injectChanID: true,
		},
		{
			name:  "host cleanup awaiting cleanup update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.AwaitingCleanup,
				},
			},
			// Balance should increase by
			// channelFundingAmount
			// 4 * hostFee in (7 ops in returned funding tx - 3 ops in cleanup tx)
			// 0.5*Lumen + 8 * channelFee in initial funding for the escrow account
			// 1 * Lumen + 1 * channelFee in initial funding for the guest ratchet
			// 0.5*Lumen + 1 * channelFee in initial funding for the host ratchet account
			walletDelta: channelFundingAmount + 2*xlm.Lumen + 4*hostFeerate + 10*channelFeerate,
		}, {
			name:  "host cleanup channel closed update",
			agent: host,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State: fsm.Closed,
				},
			},
			outOfOrder: true,
		},
		{
			name:  "host cleanup escrow merge tx update",
			agent: host,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
			walletDelta: 1 * xlm.Lumen,
		},
		{
			name:  "host cleanup host ratchet merge tx update",
			agent: host,
			update: &update.Update{
				Type: update.AccountType,
			},
			walletDelta: 1 * xlm.Lumen,
		},
		{
			name:  "host cleanup guest ratchet merge tx update",
			agent: host,
			update: &update.Update{
				Type: update.AccountType,
			},
			walletDelta: 1 * xlm.Lumen,
			// Account should be up to date with ledger at this point
			checkLedger: true,
		},
	}
}

func checkUpdate(ctx context.Context, s step, channelID *string) error {
	backoff := net.Backoff{Base: 10 * time.Second}
	found := false
	updateNum := s.agent.nextUpdateNum
	for i := 0; i < 10 && !found; i++ {
		body := fmt.Sprintf(`{"From": %d}`, updateNum)
		log.Debugf("%s: polling /api/updates %s\n", s.name, body)
		resp := post(ctx, s.agent.handler, s.agent.address, "/api/updates", body, s.agent.cookie)
		if resp.Code != 200 {
			return errors.New(fmt.Sprintf("%s: got http response: %d, want: 200", s.name, resp.Code))
		}
		var updates []update.Update
		err := json.Unmarshal(resp.Body.Bytes(), &updates)
		if err != nil {
			return err
		}
		for _, u := range updates {
			if !s.outOfOrder {
				s.agent.nextUpdateNum = u.UpdateNum + 1
			}
			if found = updateMatches(u, *s.update); found {
				if s.walletDelta != 0 {
					newBalance := s.agent.balance + s.walletDelta
					if uint64(newBalance) != u.Account.Balance {
						return errors.New(fmt.Sprintf("%s: got balance %s, want %s", s.name, xlm.Amount(u.Account.Balance), newBalance))
					}
					s.agent.balance = newBalance
				}
				if s.hostDelta != 0 {
					newHostAmount := s.agent.hostAmount + s.hostDelta
					if newHostAmount != u.Channel.HostAmount {
						return errors.New(fmt.Sprintf("%s: got host amount %s, want %s", s.name, u.Channel.HostAmount, newHostAmount))
					}
					s.agent.hostAmount = newHostAmount
				}
				if s.guestDelta != 0 {
					newGuestAmount := s.agent.guestAmount + s.guestDelta
					if newGuestAmount != u.Channel.GuestAmount {
						return errors.New(fmt.Sprintf("%s: got guest amount %s, want %s", s.name, u.Channel.GuestAmount, newGuestAmount))
					}
					s.agent.guestAmount = newGuestAmount
				}
				if s.checkLedger {
					if err = checkAcctBalance(s.agent.wclient, u.Account.ID, s.agent.balance, s.agent.wclient.Now()); err != nil {
						return errors.Wrapf(err, "step %s", s.name)
					}
				}
			}
			if found {
				break
			}
		}
		updateCookie(s.agent, resp.Header())
		if channelID != nil {
			*channelID = updateChannelID(*channelID, resp.Body.Bytes())
		}
		s.agent.accountID = updateHostAddr(resp.Body.Bytes())
		if !found {
			dur := backoff.Next()
			time.Sleep(dur)
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("%s: timed out polling for update %+v", s.name, *s.update))
	}
	return nil
}

func checkAcctBalance(wclient *worizon.Client, acctID string, want xlm.Amount, ledgerTime time.Time) error {
	return checkAcctBalanceHelper(wclient, acctID, want, ledgerTime, 4)
}

func checkAcctBalanceHelper(wclient *worizon.Client, acctID string, want xlm.Amount, ledgerTime time.Time, retries int) error {
	acc, err := wclient.LoadAccount(acctID)
	if err != nil {
		return err
	}
	balStr, err := acc.GetNativeBalance()
	if err != nil {
		return err
	}
	bal, err := xlm.Parse(balStr)
	if err != nil {
		return err
	}
	if bal != want {
		if retries > 0 {
			done := make(chan struct{})
			wclient.AfterFunc(ledgerTime.Add(time.Second), func() {
				err = checkAcctBalanceHelper(wclient, acctID, want, wclient.Now(), retries-1)
				close(done)
			})
			<-done
			return err
		}
		return fmt.Errorf("account %s: got balance %s, want %s", acctID, bal, want)
	}
	return nil
}

func updateMatches(got, want update.Update) bool {
	if got.Type != want.Type {
		return false
	}
	if want.UpdateNum != 0 && got.UpdateNum != want.UpdateNum {
		return false
	}
	if want.InputTx != nil && got.InputTx == nil {
		return false
	}
	if want.Channel != nil {
		if got.Channel.State != want.Channel.State {
			return false
		}
	}
	return true
}

func testStep(ctx context.Context, t *testing.T, s step, channelID *string) {
	t.Helper()
	err := handleStep(ctx, s, channelID)
	if err != nil {
		t.Fatal(err)
	}
}

func handleStep(ctx context.Context, s step, channelID *string) error {
	if s.update != nil {
		err := checkUpdate(ctx, s, channelID)
		if err != nil {
			return err
		}
		return nil
	}
	if s.injectChanID {
		s.body = fmt.Sprintf(s.body, *channelID)
		log.Debugf("Body: %s", s.body)
	}
	if s.injectHostAcct {
		s.body = fmt.Sprintf(s.body, s.agent.accountID)
	}
	resp := post(ctx, s.agent.handler, s.agent.address, s.path, s.body, s.agent.cookie)
	wantCode := s.wantCode
	if wantCode == 0 {
		wantCode = 200
	}
	if wantCode != resp.Code {
		return errors.New(fmt.Sprintf("%s: got http response: %d, want: %d", s.name, resp.Code, wantCode))
	}
	updateCookie(s.agent, resp.Header())
	return nil
}

func updateChannelID(orig string, body []byte) string {
	var updates []update.Update
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return orig
	}
	for _, u := range updates {
		if u.Channel != nil {
			log.Debugf("updating channel ID to: %s", u.Channel.ID)
			return u.Channel.ID
		}
	}
	return orig
}

func logWrapper(handler http.Handler, dest string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("%s: %s %s %s\n", r.Host, r.Method, r.URL.Path, dest)
		handler.ServeHTTP(w, r)
	})
}

// updateHostAddr updates the host address from the response body
// if the body is a update.Update type. Otherwise, does nothing.
func updateHostAddr(body []byte) string {
	var updates []update.Update
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return ""
	}
	if len(updates) == 0 {
		return ""
	}
	update := updates[0]
	if update.Account != nil {
		return update.Account.ID
	}
	return ""
}

func updateCookie(agent *Starlightd, header http.Header) {
	cookie := header.Get("set-cookie")
	if cookie != "" {
		agent.cookie = cookie[0:strings.Index(cookie, ";")]
	}
}

func post(ctx context.Context, h http.Handler, host, path, body, cookie string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, bytes.NewReader([]byte(body)))
	req.Host = host
	req.Header.Set("Cookie", cookie)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}
