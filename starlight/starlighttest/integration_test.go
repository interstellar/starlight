package starlighttest

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/interstellar/starlight/starlight/xlm"
)

func TestChannelCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "channel-creation")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := channelCreationSteps(alice, bob, 0, 0)
	for _, s := range steps {
		testStep(t, ctx, s, nil)
	}
}

func TestChannelOpenCloseNoPayment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "coop-close")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := append(channelCreationSteps(alice, bob, 0, 0), aliceCoopCloseSteps(alice, bob, 0)...)
	var channelID string
	for _, step := range steps {
		testStep(t, ctx, step, &channelID)
	}
}

func TestChannelPayment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "channel-pay")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := append(channelCreationSteps(alice, bob, 0, 0), bobChannelPayAliceSteps(alice, bob)...)
	var channelID string
	for _, step := range steps {
		testStep(t, ctx, step, &channelID)
	}
}

func TestChannelSettleWithGuestCoopClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "settle-with-guest-coop-close")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := append(channelCreationSteps(alice, bob, 0, 0), bobChannelPayAliceSteps(alice, bob)...)
	steps = append(steps, aliceCoopCloseSteps(alice, bob, 1000*xlm.Stroop)...)
	var channelID string
	for _, step := range steps {
		testStep(t, ctx, step, &channelID)
	}
}

func TestDuplicateChannelCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "duplicate-channel")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

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
		testStep(t, ctx, step, &channelID)
	}
}

func TestChannelNoPaymentForceClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "no-payment-force-close")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := append(channelCreationSteps(alice, bob, 1, 1), bobForceCloseStepsNoGuestBalance(alice, bob)...)
	var channelID string
	for _, step := range steps {
		testStep(t, ctx, step, &channelID)
	}
}

func TestSettleWithGuestForceClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "settle-guest-force-close")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := append(channelCreationSteps(alice, bob, 1, 1), bobChannelPayAliceSteps(alice, bob)...)
	steps = append(steps, bobForceCloseStepsSettleWithGuest(alice, bob, 1000*xlm.Stroop)...)
	var channelID string
	for _, step := range steps {
		testStep(t, ctx, step, &channelID)
	}
}

func TestHostTopUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "host-top-up")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := append(channelCreationSteps(alice, bob, 0, 0), hostTopUpSteps(alice, bob)...)

	var channelID string
	for _, s := range steps {
		testStep(t, ctx, s, &channelID)
	}
}

func TestLogoutLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.Printf("running %s", t.Name())
	testdir, err := ioutil.TempDir("", "logout-login")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alice := start(t, ctx, testdir, "alice")
	defer alice.server.Close()
	bob := start(t, ctx, testdir, "bob")
	defer bob.server.Close()

	steps := append(channelCreationSteps(alice, bob, 0, 0), logoutSteps(alice, bob)...)
	steps = append(steps, loginSteps(alice, bob)...)

	var channelID string
	for _, s := range steps {
		testStep(t, ctx, s, &channelID)
	}
}
