package starlighttest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/internal/update"
	"github.com/interstellar/starlight/worizon"
	"github.com/interstellar/starlight/worizon/xlm"
)

const (
	channelFeerate = 10 * xlm.Millilumen
	hostFeerate    = 100 * xlm.Stroop
	hostAmount     = 1000 * xlm.Lumen
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
	delta          xlm.Amount
}

func WalletPaySelf(ctx context.Context, self *Starlightd) error {
	steps := []step{
		{
			name:  "do wallet pay",
			agent: self,
			path:  "/api/do-wallet-pay",
			body: `{
				"Dest":"%s",
				"Amount":1000000000
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

func CreateChannel(ctx context.Context, alice, bob *Starlightd) (string, error) {
	steps := channelCreationSteps(alice, bob, 0, 0)
	var channelID string
	for _, s := range steps {
		err := handleStep(ctx, s, &channelID)
		if err != nil {
			return "", err
		}
	}
	return channelID, nil
}

func BobChannelPayAlice(ctx context.Context, alice, bob *Starlightd, channelID string) error {
	steps := bobChannelPayAliceSteps(alice, bob)
	for _, s := range steps {
		err := handleStep(ctx, s, &channelID)
		if err != nil {
			return err
		}
	}
	return nil
}

func BobTopUp(ctx context.Context, alice, bob *Starlightd, channelID string) error {
	steps := hostTopUpSteps(alice, bob)
	for _, s := range steps {
		err := handleStep(ctx, s, &channelID)
		if err != nil {
			return err
		}
	}
	return nil
}

func channelCreationSteps(alice, bob *Starlightd, maxRoundDurMin, finalityDelayMin int) []step {
	// WARNING: this software is not compatible with Stellar mainnet.
	return []step{
		{
			name:  "alice config init",
			agent: alice,
			path:  "/api/config-init",
			// WARNING: this software is not compatible with Stellar mainnet.
			body: `
			{
				"Username":"alice",
				"Password":"password",
				"DemoServer":true,
				"HorizonURL":"https://horizon-testnet.stellar.org/",
				"KeepAlive":false
			}`,
		}, {
			name:  "alice config init update",
			agent: alice,
			update: &update.Update{
				Type:      update.InitType,
				UpdateNum: 1,
			},
		}, {
			name:  "bob config init",
			agent: bob,
			path:  "/api/config-init",
			// WARNING: this software is not compatible with Stellar mainnet.
			body: fmt.Sprintf(`
			{
				"Username":"bob",
				"Password":"password",
				"DemoServer":true,
				"HorizonURL":"https://horizon-testnet.stellar.org/",
				"KeepAlive":false,
				"MaxRoundDurMin": %d,
				"FinalityDelayMin": %d,
				"HostFeerate": %d,
				"ChannelFeerate":%d
			}`, maxRoundDurMin, finalityDelayMin, hostFeerate, channelFeerate),
		}, {
			name:  "bob config init update",
			agent: bob,
			update: &update.Update{
				Type:      update.InitType,
				UpdateNum: 1,
			},
		}, {
			name:  "alice wallet funding update",
			agent: alice,
			update: &update.Update{
				Type:      update.AccountType,
				UpdateNum: 2,
			},
			delta:       10000 * xlm.Lumen,
			checkLedger: true,
		}, {
			name:  "bob wallet funding update",
			agent: bob,
			path:  "/api/updates",
			body:  `{"From": 2}`,
			update: &update.Update{
				Type:      update.AccountType,
				UpdateNum: 2,
			},
			delta:       10000 * xlm.Lumen,
			checkLedger: true,
		}, {
			name:  "bob create channel with alice",
			agent: bob,
			path:  "/api/do-create-channel",
			body: fmt.Sprintf(`{
				"GuestAddr": "alice*%s",
				"HostAmount": 10000000000
			}`, alice.address),
		}, {
			name:  "bob channel creation setting up update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.SettingUp,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
			// Host balance should decrease by:
			// 1000 Lumen for channel funding amount
			// 3 * Lumen + 3 * hostFee to create the three accounts
			// 7 * hostFee in funding tx fees for 7 operations
			// 0.5*Lumen + 8 * channelFee in initial funding for the escrow account
			// 1 * Lumen + 1 * channelFee in initial funding for the guest ratchet
			// 0.5*Lumen + 1 * channelFee in initial funding for the host ratchet account
			delta: -(1000*xlm.Lumen + 5*xlm.Lumen + 10*hostFeerate + 10*channelFeerate),
		}, {
			name:  "bob channel creation channel proposed update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.ChannelProposed,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			name:  "alice channel creation awaiting funding update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingFunding,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			name:  "alice channel creation awaiting funding tx update",
			agent: alice,
			update: &update.Update{
				Type:    update.ChannelType,
				InputTx: &fsm.Tx{},
				Channel: &fsm.Channel{
					State:       fsm.AwaitingFunding,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			name:  "bob channel creation awaiting funding update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingFunding,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			name:  "alice channel creation open update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			name:  "bob channel creation open update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		},
	}
}

func bobChannelPayAliceSteps(alice, bob *Starlightd) []step {
	return []step{
		{
			name:  "bob channel pay to alice",
			agent: bob,
			path:  "/api/do-command",
			body: `
			{
				"ChannelID": "%s",
				"Command": {
					"UserCommand": "ChannelPay",
					"Amount": 1000,
					"Time": "2018-10-02T10:26:43.511Z"
				}
			}`,
			injectChanID: true,
		}, {
			name:  "bob channel pay to alice payment proposed update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.PaymentProposed,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			name:  "bob channel pay to alice payment accept update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.PaymentAccepted,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			name:  "bob channel pay to alice payment bob channel open update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostAmount - 1000*xlm.Stroop,
					GuestAmount: 1000,
				},
			},
		}, {
			name:  "bob channel pay to alice payment alice channel open update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostAmount - 1000*xlm.Stroop,
					GuestAmount: 1000,
				},
			},
		},
	}
}

func aliceCoopCloseSteps(alice, bob *Starlightd, payment xlm.Amount) []step {
	var guestFee xlm.Amount
	if payment != 0 {
		// Add additional channelFeerate for additional operation cost to
		// settle with guest
		guestFee = channelFeerate
	}
	return []step{
		{
			name:  "alice cooperative close",
			agent: alice,
			path:  "/api/do-command",
			body: `
				{
					"ChannelID": "%s",
					"Command": {
						"UserCommand": "CloseChannel",
						"Time": "2018-10-02T10:26:43.511Z"
					}
				}`,
			injectChanID: true,
		}, {
			name:  "alice cooperative close awaiting close update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingClose,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		}, {
			name:  "bob cooperative close awaiting close update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingClose,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		}, {
			name:  "bob cooperative close channel closed update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Closed,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
			outOfOrder: true,
		}, {
			name:  "bob cooperative close coop close tx merge escrow account",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// Balance should increase by:
			// 1000 Lumen initial funding
			// 1 Lumen minimum escrow account balance
			// 0.5 Lumen multisig escrow account balance
			// 8 * channelFeerate initial funding
			// less 3 * channelFeerate coop close tx fees (one added above if settling with guest)
			// less payment amount
			delta: 1000*xlm.Lumen + 1*xlm.Lumen + 500*xlm.Millilumen + 8*channelFeerate - 3*channelFeerate - payment - guestFee,
		}, {
			name:  "bob cooperative close coop close tx merge guest ratchet account",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// Balance should increase by
			// 2 Lumens initial funding + channelFeerate
			delta: 2*xlm.Lumen + channelFeerate,
		}, {
			name:  "bob cooperative close coop close tx merge host ratchet account",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// Balance should increase by
			// 1.5 Lumens initial funding + channelFeerate
			delta: 1*xlm.Lumen + 500*xlm.Millilumen + channelFeerate,
		}, {
			name:  "bob cooperative close channel closed update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Closed,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
			delta: 0,
			// Account should be up to date with ledger at this point
			checkLedger: true,
		}, {
			name:  "alice cooperative close channel closed update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Closed,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		},
	}
}

func bobForceCloseStepsNoGuestBalance(alice, bob *Starlightd) []step {
	return []step{
		{
			name:  "bob force close command",
			agent: bob,
			path:  "/api/do-command",
			body: `{
				"ChannelID": "%s",
				"Command": {
					"UserCommand": "ForceClose"
				}
			}`,
			injectChanID: true,
		}, {
			// bob transitions to AwaitingRatchet state,
			// outputs current ratchet tx
			name:  "bob force close state transition awaiting ratchet",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingRatchet,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			// bob sees current ratchet tx hit the ledger,
			// transitions to awaiting settlement mintime
			name:  "bob force close state transition awaiting settlement mintime",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingSettlementMintime,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			// alice sees current ratchet tx hit the ledger,
			// transitions to closed
			name:  "alice force close state transition closed",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Closed,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			// bob waits until settlement mintime, then transitions to
			// awaiting settlement
			name:  "bob force close state transition awaiting settlement",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingSettlement,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
		}, {
			// bob sees current settlement tx hit the ledger, transitions to
			// closed state
			name:  "bob force close state transition closed",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Closed,
					HostAmount:  hostAmount,
					GuestAmount: 0,
				},
			},
			delta:      0,
			outOfOrder: true,
		}, {
			name:  "bob force close merge escrow account",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// Balance should increase by:
			// 1000 Lumen initial funding
			// 1 Lumen minimum escrow account balance
			// 0.5 Lumen multisig escrow account balance
			// 8 * channelFeerate initial funding
			// less 3 * channelFeerate ratchet / settlement tx fees
			delta: 1000*xlm.Lumen + 1*xlm.Lumen + 500*xlm.Millilumen + 8*channelFeerate - 3*channelFeerate,
		}, {
			name:  "bob force close merge guest ratchet",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// merge back 2 Lumens guest ratchet balance
			delta: 2*xlm.Lumen + channelFeerate,
		}, {
			name:  "bob force close merge host ratchet",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// merge back 1.5 Lumens host ratchet balance
			delta:       1*xlm.Lumen + 500*xlm.Millilumen,
			checkLedger: true,
		},
	}
}

func bobForceCloseStepsSettleWithGuest(alice, bob *Starlightd, payment xlm.Amount) []step {
	return []step{
		{
			name:  "bob force close command",
			agent: bob,
			path:  "/api/do-command",
			body: `{
				"ChannelID": "%s",
				"Command": {
					"UserCommand": "ForceClose"
				}
			}`,
			injectChanID: true,
		}, {
			// bob sees current ratchet tx hit the ledger,
			// transitions to awaiting settlement mintime
			name:  "bob force close state transition awaiting settlement mintime",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingSettlementMintime,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		}, {
			// alice sees current ratchet tx hit the ledger,
			// transitions to awaiting settlement mintime
			name:  "alice force close state transition closed",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingSettlementMintime,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		}, {
			// bob waits until settlement mintime, then transitions to
			// awaiting settlement
			name:  "bob force close state transition awaiting settlement",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingSettlement,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		}, {
			// alice waits until settlement mintime, then transitions to
			// awaiting settlement
			name:  "alice force close state transition awaiting settlement",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingSettlement,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		}, {
			// bob sees current settlement tx hit the ledger, transitions to
			// closed state
			name:  "bob force close state transition closed",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Closed,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
			// state closed update could come before or after merge tx updates
			outOfOrder: true,
		}, {
			// alice sees current settlement tx hit the ledger, transitions to
			// closed state
			name:  "alice force close state transition closed",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Closed,
					HostAmount:  hostAmount - payment,
					GuestAmount: payment,
				},
			},
		}, {
			name:  "bob force close merge escrow account",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// Balance should increase by:
			// 1000 Lumen initial funding
			// 1 Lumen minimum escrow account balance
			// 0.5 Lumen multisig escrow account balance
			// 8 * channelFeerate initial funding
			// less 4 * channelFeerate ratchet / settlement tx fees
			// less 1000 stroops sent to Alice
			delta: 1000*xlm.Lumen + 1*xlm.Lumen + 500*xlm.Millilumen + 8*channelFeerate - 4*channelFeerate - payment,
		}, {
			name:  "bob force close merge guest ratchet",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// merge back 2 Lumens guest ratchet balance
			delta: 2 * xlm.Lumen,
		}, {
			name:  "bob force close merge host ratchet",
			agent: bob,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &fsm.Tx{},
			},
			// merge back 1.5 Lumens host ratchet balance
			delta:       1*xlm.Lumen + 500*xlm.Millilumen + channelFeerate,
			checkLedger: true,
		},
	}
}

func hostTopUpSteps(alice, bob *Starlightd) []step {
	return []step{
		{
			name:  "bob top-up command",
			agent: bob,
			path:  "/api/do-command",
			body: `{
				"ChannelID": "%s",
				"Command": {
					"UserCommand": "TopUp",
					"Amount":500
				}
			}`,
			injectChanID: true,
		}, {
			name:  "bob top-up channel update",
			agent: bob,
			update: &update.Update{
				Type:    update.ChannelType,
				InputTx: &fsm.Tx{},
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostAmount + 500*xlm.Stroop,
					GuestAmount: 0,
				},
			},
			// balance decreases by top-up amount + host fee rate
			delta: -(500*xlm.Stroop + hostFeerate),
		}, {
			name:  "alice top-up tx update",
			agent: alice,
			update: &update.Update{
				Type:    update.ChannelType,
				InputTx: &fsm.Tx{},
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostAmount + 500*xlm.Stroop,
					GuestAmount: 0,
				},
			},
		},
	}
}

// mergingPaymentSteps sends payments from Alice and Bob that, if one of the agents is not
// receiving messages, will eventually create a payment merge state when the agent comes
// back online.
func mergingPaymentSteps(alice, bob *Starlightd, hostBalance, guestBalance xlm.Amount) []step {
	return []step{
		{
			name:  "alice channel pay bob",
			agent: alice,
			path:  "/api/do-command",
			body: `
			{
				"ChannelID": "%s",
				"Command": {
					"UserCommand": "ChannelPay",
					"Amount": 1000,
					"Time": "2018-10-02T10:26:43.511Z"
				}
			}`,
			injectChanID: true,
		}, {
			name:  "bob channel pay alice",
			agent: bob,
			path:  "/api/do-command",
			body: `
			{
				"ChannelID": "%s",
				"Command": {
					"UserCommand": "ChannelPay",
					"Amount": 1500,
					"Time": "2018-10-02T10:26:43.511Z"
				}
			}`,
			injectChanID: true,
		},
	}
}

// paymentMergeResolutionSteps represents the state transitions that Alice and Bob go through to merge
// the conflict payment, sending a merged payment from Bob to Alice.
func paymentMergeResolutionSteps(alice, bob *Starlightd, hostBalance, guestBalance xlm.Amount) []step {
	return []step{
		{
			name:  "bob channel pay alice payment proposed update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.PaymentProposed,
					HostAmount:  hostBalance,
					GuestAmount: guestBalance,
				},
			},
		}, {
			name:  "alice channel pay bob payment proposed update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.PaymentProposed,
					HostAmount:  hostBalance,
					GuestAmount: guestBalance,
				},
			},
		}, {
			name:  "alice channel pay awaiting payment merge update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.AwaitingPaymentMerge,
					HostAmount:  hostBalance,
					GuestAmount: guestBalance,
				},
			},
		}, {
			name:  "bob channel pay payment proposed upate",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.PaymentProposed,
					HostAmount:  hostBalance,
					GuestAmount: guestBalance,
				},
			},
		}, {
			name:  "alice channel pay payment accepted update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.PaymentAccepted,
					HostAmount:  hostBalance,
					GuestAmount: guestBalance,
				},
			},
		}, {
			name:  "bob channel pay payment complete state open update",
			agent: bob,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostBalance - 500*xlm.Stroop,
					GuestAmount: guestBalance + 500*xlm.Stroop,
				},
			},
		}, {
			name:  "alice channel pay open update",
			agent: alice,
			update: &update.Update{
				Type: update.ChannelType,
				Channel: &fsm.Channel{
					State:       fsm.Open,
					HostAmount:  hostBalance - 500*xlm.Stroop,
					GuestAmount: guestBalance + 500*xlm.Stroop,
				},
			},
		},
	}
}

func logoutSteps(alice, bob *Starlightd) []step {
	return []step{
		{
			name:  "alice logout command",
			agent: alice,
			path:  "/api/logout",
		}, {
			name:     "alice logout update",
			agent:    alice,
			path:     "/api/updates",
			body:     `{"From":1}`,
			wantCode: 401,
		},
	}
}

func loginSteps(alice, bob *Starlightd) []step {
	return []step{
		{
			name:  "alice login command",
			agent: alice,
			path:  "/api/login",
			body:  `{"Username":"alice","Password":"password"}`,
		}, {
			name:  "alice login update",
			agent: alice,
			path:  "/api/updates",
			body:  `{"From":2}`,
		},
	}
}

func checkUpdate(ctx context.Context, s step, channelID *string) error {
	backoff := net.Backoff{Base: 10 * time.Second}
	found := false
	updateNum := s.agent.nextUpdateNum
	for i := 0; i < 10 && !found; i++ {
		body := fmt.Sprintf(`{"From": %d}`, updateNum)
		log.Printf("%s: polling /api/updates %s\n", s.name, body)
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
				if s.delta != 0 {
					newBalance := s.agent.balance + s.delta
					if uint64(newBalance) != u.Account.Balance {
						return errors.New(fmt.Sprintf("%s: got balance %s, want %s", s.name, xlm.Amount(u.Account.Balance), newBalance))
					}
					s.agent.balance = newBalance
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
		if got.Channel.HostAmount != want.Channel.HostAmount {
			log.Fatalf("incorrect host amount: got %s, want %s", got.Channel.HostAmount, want.Channel.HostAmount)
			return false
		}
		if got.Channel.GuestAmount != want.Channel.GuestAmount {
			log.Fatalf("incorrect guest amount: got %s, want %s", got.Channel.GuestAmount, want.Channel.GuestAmount)
			return false
		}
	}
	return true
}

func testStep(t *testing.T, ctx context.Context, s step, channelID *string) {
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
		log.Printf("Body: %s", s.body)
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
			log.Printf("updating channel ID to: %s", u.Channel.ID)
			return u.Channel.ID
		}
	}
	return orig
}

func logWrapper(handler http.Handler, dest string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s %s %s\n", r.Host, r.Method, r.URL.Path, dest)
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
