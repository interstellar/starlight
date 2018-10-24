package starlighttest

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/interstellar/starlight/starlight/internal/update"
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
			// TODO(vniu): replace Horizon URL with non-rate limited Horizon.
			// WARNING: this software is not compatible with Stellar mainnet.
			body: `
			{
				"Username":"vicki",
				"Password":"password",
				"DemoServer":true,
				"HorizonURL":"https://horizon-testnet.stellar.org"
			}`,
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
				Type:      update.AccountType,
				UpdateNum: 3,
			},
		}, {
			name:  "check wallet pay txsuccess update",
			agent: alice,
			update: &update.Update{
				Type:      update.TxSuccessType,
				UpdateNum: 4,
			},
		}, {
			name:  "check wallet payment received account update",
			agent: alice,
			update: &update.Update{
				Type:      update.AccountType,
				UpdateNum: 5,
			},
		},
	}
	for _, s := range steps {
		testStep(ctx, t, s, nil)
	}
}
