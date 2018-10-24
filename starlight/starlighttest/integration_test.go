package starlighttest

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/interstellar/starlight/worizon/xlm"
)

// itest runs an alice-and-bob integration test.
func itest(t *testing.T, f func(ctx context.Context, alice, bob *Starlightd)) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()

	alice := start(ctx, t, testdir, "alice")
	defer alice.Close()

	bob := start(ctx, t, testdir, "bob")
	defer bob.Close()

	f(ctx, alice, bob)
}

func TestChannelCreation(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := channelCreationSteps(alice, bob, 0, 0)
		for _, s := range steps {
			testStep(ctx, t, s, nil)
		}
	})
}

func TestChannelOpenCloseNoPayment(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 0, 0), aliceCoopCloseSteps(alice, bob, 0)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestChannelPayment(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 0, 0), bobChannelPayAliceSteps(alice, bob)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestChannelSettleWithGuestCoopClose(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 0, 0), bobChannelPayAliceSteps(alice, bob)...)
		steps = append(steps, aliceCoopCloseSteps(alice, bob, 1000*xlm.Stroop)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestDuplicateChannelCreation(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 0, 0), step{
			name:  "bob create channel with alice",
			agent: bob,
			path:  "/api/do-create-channel",
			body: fmt.Sprintf(`{
				"GuestAddr": "alice*%s", 
				"HostAmount": 10000000000
			}`, alice.address),
			wantCode: http.StatusResetContent,
		})
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestChannelNoPaymentForceClose(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 1, 1), bobForceCloseStepsNoGuestBalance(alice, bob)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestSettleWithGuestForceClose(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 1, 1), bobChannelPayAliceSteps(alice, bob)...)
		steps = append(steps, bobForceCloseStepsSettleWithGuest(alice, bob, 1000*xlm.Stroop)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestHostTopUp(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 0, 0), hostTopUpSteps(alice, bob)...)

		var channelID string
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}
	})
}

func TestLogoutLogin(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 0, 0), logoutSteps(alice, bob)...)
		steps = append(steps, loginSteps(alice, bob)...)

		var channelID string
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}
	})
}

func TestCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	alice := TestServer("alice")
	defer alice.Close()

	bob := start(ctx, t, testdir, "bob")
	defer bob.server.Close()

	steps := cleanupSteps(alice, bob, 0, 0)
	var channelID string
	for _, s := range steps {
		testStep(ctx, t, s, &channelID)
	}
}

func TestPaymentMerge(t *testing.T) {
	itest(t, func(ctx context.Context, alice, bob *Starlightd) {
		steps := append(channelCreationSteps(alice, bob, 0, 0), bobChannelPayAliceSteps(alice, bob)...)

		// Create channel and do one payment
		var channelID string
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}

		address := bob.address
		bob.server.Close()

		hostBalance := 1000*xlm.Lumen - 1000*xlm.Stroop
		guestBalance := 1000 * xlm.Stroop

		steps = mergingPaymentSteps(alice, bob, hostBalance, guestBalance)
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}

		bob.server = httptest.NewUnstartedServer(bob.handler)
		l, err := net.Listen("tcp", address)
		if err != nil {
			t.Fatal(err)
		}
		bob.server.Listener.Close()
		bob.server.Listener = l
		bob.server.Start()

		steps = paymentMergeResolutionSteps(alice, bob, hostBalance, guestBalance)
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}
	})

}
