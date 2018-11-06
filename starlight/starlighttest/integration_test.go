package starlighttest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/interstellar/starlight/starlight/log"
	"github.com/interstellar/starlight/worizon/xlm"
)

const (
	channelFundingAmount = 1000 * xlm.Lumen
	paymentAmount        = 1000 * xlm.Stroop
	topUpAmount          = 500 * xlm.Stroop
)

// itest runs an guest-and-host integration test.
func itest(t *testing.T, f func(ctx context.Context, guest, host *Starlightd)) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Infof("running %s", t.Name())
	testdir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()

	guest := start(ctx, t, testdir, "guest")
	defer guest.Close()

	host := start(ctx, t, testdir, "host")
	defer host.Close()

	f(ctx, guest, host)
}

func TestChannelCreation(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := channelCreationSteps(guest, host, 0, 0, channelFundingAmount)
		for _, s := range steps {
			testStep(ctx, t, s, nil)
		}
	})
}

func TestChannelOpenCloseNoPayment(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 0, 0, channelFundingAmount), guestCoopCloseSteps(guest, host, channelFundingAmount, 0)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestChannelPayment(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 0, 0, channelFundingAmount), hostChannelPayGuestSteps(guest, host, paymentAmount)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestChannelSettleWithGuestCoopClose(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 0, 0, channelFundingAmount), hostChannelPayGuestSteps(guest, host, paymentAmount)...)
		steps = append(steps, guestCoopCloseSteps(guest, host, channelFundingAmount, paymentAmount)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestDuplicateChannelCreation(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 0, 0, channelFundingAmount), step{
			name:  "host create channel with guest",
			agent: host,
			path:  "/api/do-create-channel",
			body: fmt.Sprintf(`{
				"GuestAddr": "guest*%s", 
				"HostAmount": %d
			}`, guest.address, channelFundingAmount),
			wantCode: http.StatusBadRequest,
		})
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestChannelNoPaymentForceClose(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 1, 1, channelFundingAmount), hostForceCloseStepsNoGuestBalance(guest, host, channelFundingAmount)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestSettleWithGuestForceClose(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 1, 1, channelFundingAmount), hostChannelPayGuestSteps(guest, host, paymentAmount)...)
		steps = append(steps, hostForceCloseStepsSettleWithGuest(guest, host, channelFundingAmount, paymentAmount)...)
		var channelID string
		for _, step := range steps {
			testStep(ctx, t, step, &channelID)
		}
	})
}

func TestHostTopUp(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 0, 0, channelFundingAmount), hostTopUpSteps(guest, host, topUpAmount)...)

		var channelID string
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}
	})
}

func TestLogoutLogin(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 0, 0, channelFundingAmount), logoutSteps(guest, host)...)
		steps = append(steps, loginSteps(guest, host)...)

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
	log.Infof("running %s", t.Name())
	testdir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	guest := testServer("guest")
	defer guest.Close()

	host := start(ctx, t, testdir, "host")
	defer host.server.Close()

	steps := cleanupSteps(guest, host, 0, 0, channelFundingAmount)
	var channelID string
	for _, s := range steps {
		testStep(ctx, t, s, &channelID)
	}
}

func TestPaymentMerge(t *testing.T) {
	itest(t, func(ctx context.Context, guest, host *Starlightd) {
		steps := append(channelCreationSteps(guest, host, 0, 0, channelFundingAmount), hostChannelPayGuestSteps(guest, host, paymentAmount)...)

		// Create channel and do one payment
		var channelID string
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}

		address := host.address
		host.server.Close()

		hostPayment := paymentAmount + 1*xlm.Lumen
		guestPayment := paymentAmount
		steps = mergingPaymentSteps(guest, host, guestPayment, hostPayment)
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}

		host.server = httptest.NewUnstartedServer(host.handler)
		l, err := net.Listen("tcp", address)
		if err != nil {
			t.Fatal(err)
		}
		host.server.Listener.Close()
		host.server.Listener = l
		host.server.Start()

		steps = paymentMergeResolutionSteps(guest, host, hostPayment-guestPayment)
		for _, s := range steps {
			testStep(ctx, t, s, &channelID)
		}
	})

}
