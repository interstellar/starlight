package starlighttest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/internal/update"
	"github.com/interstellar/starlight/worizon"
)

func TestAgentRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testdir, err := ioutil.TempDir("", "TestAgentRequest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()

	alice := start(ctx, t, testdir, "test")
	defer alice.Close()

	steps := []step{
		{
			name:  "config init",
			agent: alice,
			path:  "/api/config-init",
			// WARNING: this software is not compatible with Stellar mainnet.
			body: fmt.Sprintf(`
			{
				"Username":"starlight",
				"Password":"password",
				"DemoServer":true,
				"HorizonURL":"%s"
			}`, *HorizonURL),
		}, {
			name:  "get init update",
			agent: alice,
			update: &update.Update{
				Type:      update.InitType,
				UpdateNum: 1,
			},
		}, {
			name:  "get wallet funding update",
			agent: alice,
			update: &update.Update{
				Type:      update.AccountType,
				UpdateNum: 2,
			},
		}, {
			name:           "do wallet pay",
			agent:          alice,
			path:           "/api/do-wallet-pay",
			injectHostAcct: true,
			body: `
			{
				"Dest":"%s",
				"Amount":1000000000
			}`,
		}, {
			name:  "check wallet pay update",
			agent: alice,
			update: &update.Update{
				Type: update.AccountType,
				InputCommand: &fsm.Command{
					Name: "WalletPay",
				},
			},
		}, {
			name:  "check wallet pay txsuccess update",
			agent: alice,
			update: &update.Update{
				Type: update.TxSuccessType,
			},
		}, {
			name:  "check wallet payment received account update",
			agent: alice,
			update: &update.Update{
				Type:    update.AccountType,
				InputTx: &worizon.Tx{},
			},
		},
	}
	for _, s := range steps {
		testStep(ctx, t, s, nil)
	}
}
