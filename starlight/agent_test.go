package starlight

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stellar/go/protocols/horizon"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/starlight/db"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/internal/update"
	"github.com/interstellar/starlight/worizon/xlm"
)

// WARNING: this software is not compatible with Stellar mainnet.
var testHorizonURL = "https://horizon-testnet.stellar.org"

func TestConfigInit(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	config := Config{
		Username:   "alice",
		Password:   "password",
		HorizonURL: testHorizonURL,
	}
	err := g.ConfigInit(&config, "")
	if err != nil {
		t.Fatal(err)
	}
	err = g.ConfigInit(&config, "")
	if err != errAlreadyConfigured {
		t.Errorf("got %s, want %s", err, errAlreadyConfigured)
	}
}

func TestConfigured(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	config := Config{
		Username:   "alice",
		Password:   "password",
		HorizonURL: testHorizonURL,
	}
	if g.Configured() {
		t.Errorf("g.Configured() = true, want false")
	}

	err := g.ConfigInit(&config, "")
	if err != nil {
		t.Fatal(err)
	}
	if !g.Configured() {
		t.Errorf("g.Configured() = false, want true")
	}
}

func TestUnconfiguredUpdates(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	got, err := json.Marshal(g.Updates(100, 200))
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("[]")
	if !bytes.Equal(got, want) {
		t.Errorf("json.Marshal(g.Updates(100, 200)) = %s, want %s", got, want)
	}
}

func TestUpdatesNull(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	config := Config{
		Username:   "alice",
		Password:   "password",
		HorizonURL: testHorizonURL,
	}
	err := g.ConfigInit(&config, "")
	if err != nil {
		t.Fatal(err)
	}

	got, err := json.Marshal(g.Updates(100, 200))
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("[]")
	if !bytes.Equal(got, want) {
		t.Errorf("json.Marshal(g.Updates(100, 200)) = %s, want %s", got, want)
	}
}

func TestInitConfigUpdate(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	config := Config{
		Username:   "alice",
		Password:   "password",
		HorizonURL: testHorizonURL,
	}
	err := g.ConfigInit(&config, "")
	if err != nil {
		t.Fatal(err)
	}

	gotUpdates := g.Updates(1, 100)

	got, err := json.Marshal(gotUpdates)
	if err != nil {
		t.Fatal(err)
	}

	wantUpdate := []Update{{
		Type:      update.InitType,
		UpdateNum: 1,
		Config: &update.Config{
			Username:          "alice",
			Password:          "[redacted]",
			HorizonURL:        testHorizonURL,
			MaxRoundDurMins:   60,
			FinalityDelayMins: 60,
			ChannelFeerate:    100000,
			HostFeerate:       100,
			KeepAlive:         true,
		},
		Account: &update.Account{
			ID:      "", // account ID is non-deterministic and checked below
			Balance: 0,
		},
	}}

	gotAccountID := gotUpdates[0].Account.ID
	keyType, err := horizon.KeyTypeFromAddress(gotAccountID)
	if err != nil || keyType != "ed25519_public_key" {
		t.Errorf("g.Updates(0, 100)[0].Account.ID = %s, want a valid Account ID", gotAccountID)
	}

	wantUpdate[0].Account.ID = gotUpdates[0].Account.ID
	wantUpdate[0].UpdateLedgerTime = gotUpdates[0].UpdateLedgerTime

	want, err := json.Marshal(wantUpdate)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("json.Marshal(g.Updates(0, 100)) = %s, want %s", got, want)
	}
}

func TestConfigEdit(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	config := Config{
		Username:   "alice",
		Password:   "password",
		HorizonURL: testHorizonURL,
	}
	err := g.ConfigEdit(&Config{
		Password:    "new password",
		OldPassword: "password",
	})
	if err != errNotConfigured {
		t.Errorf("got %s, want %s", err, errNotConfigured)
	}

	err = g.ConfigEdit(&Config{
		Password:    "",
		OldPassword: "new password",
	})
	if err != errEmptyConfigEdit {
		t.Errorf("got %s, want %s", err, errEmptyConfigEdit)
	}
	err = g.ConfigInit(&config, "")
	if err != nil {
		t.Fatal(err)
	}

	// WARNING: this software is not compatible with Stellar mainnet.
	newHorizonURL := "https://new-horizon-testnet.stellar.org/"
	newFinalityDelayMins := int64(30)
	edit := Config{
		Password:          "new password",
		OldPassword:       "password",
		HorizonURL:        newHorizonURL,
		FinalityDelayMins: newFinalityDelayMins,
	}
	err = g.ConfigEdit(&edit)
	if err != nil {
		t.Fatal(err)
	}

	var acctID string
	db.View(g.db, func(root *db.Root) error {
		url := root.Agent().Config().HorizonURL()
		if url != newHorizonURL {
			t.Errorf("got %s horizon url, want %s", url, newHorizonURL)
		}
		FinalityDelayMins := root.Agent().Config().FinalityDelayMins()
		if FinalityDelayMins != newFinalityDelayMins {
			t.Errorf("got %d finality delay, want %d", FinalityDelayMins, newFinalityDelayMins)
		}
		acctID = root.Agent().PrimaryAcct().Address()
		return nil
	})

	gotUpdates := g.Updates(2, 4)
	wantUpdates := []Update{{
		Type:      update.ConfigType,
		UpdateNum: 2,
		Config: &update.Config{
			Password:          "[redacted]",
			HorizonURL:        newHorizonURL,
			FinalityDelayMins: newFinalityDelayMins,
		},
		Account: &update.Account{
			Balance: 0,
			ID:      acctID,
		},
	}}
	if len(gotUpdates) != len(wantUpdates) {
		t.Errorf("g.Updates(2, 4): got %d updates, want %d", len(gotUpdates), len(wantUpdates))
	}

	for i, u := range gotUpdates {
		wantUpdates[i].UpdateLedgerTime = u.UpdateLedgerTime
	}

	gotUpdatesJSON, err := json.Marshal(gotUpdates)
	if err != nil {
		t.Fatal(err)
	}

	wantUpdatesJSON, err := json.Marshal(wantUpdates)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(gotUpdatesJSON, wantUpdatesJSON) {
		t.Errorf("json.Marshal(g.Updates(2, 4)) = %s, want %s", gotUpdatesJSON, wantUpdatesJSON)
	}

	var olddigest []byte
	db.View(g.db, func(root *db.Root) error {
		olddigest = root.Agent().Config().PwHash()
		return nil
	})

	err = g.ConfigEdit(&Config{
		Password:    "new password",
		OldPassword: "wrong old password",
	})
	if errors.Root(err) != errPasswordsDontMatch {
		t.Errorf("got %s, want %s", err, errPasswordsDontMatch)
	}

	db.View(g.db, func(root *db.Root) error {
		newdigest := root.Agent().Config().PwHash()
		if !bytes.Equal(olddigest, newdigest) {
			t.Errorf("got %x pw hash, want %x", newdigest, olddigest)
		}
		return nil
	})

	err = g.ConfigEdit(&Config{
		Username: "bob",
	})
	if err != errInvalidEdit {
		t.Errorf("got %s, want %s", err, errInvalidEdit)
	}
}

func TestAuthenticate(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	config := Config{
		Username:   "alice",
		Password:   "password",
		HorizonURL: testHorizonURL,
	}

	err := g.ConfigInit(&config, "")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		username string
		password string
		want     bool
	}{{
		username: config.Username,
		password: config.Password,
		want:     true,
	}, {
		username: config.Username,
		password: "not the right password",
		want:     false,
	}, {
		username: "incorrect username",
		password: config.Password,
		want:     false,
	}}

	for _, c := range cases {
		got := g.Authenticate(c.username, c.password)
		if got != c.want {
			t.Errorf("g.Authenticate(%s, %s) = %t, want %t", c.username, c.password, got, c.want)
		}
	}
}

func TestAgentCreateChannel(t *testing.T) {
	successGuestAddr := "bob*starlight.com"
	successHostAddr := "starlight.com"
	cases := []struct {
		name       string
		guestAddr  string
		hostAmount xlm.Amount
		host       string
		want       error
		agentFunc  func(g *Agent)
	}{
		{
			name:       "success",
			guestAddr:  successGuestAddr,
			hostAmount: 1 * xlm.Lumen,
			host:       successHostAddr,
			want:       nil,
		},
		{
			name:       "same host guest addresses",
			guestAddr:  "alice*starlight.com",
			hostAmount: 1 * xlm.Lumen,
			host:       successHostAddr,
			agentFunc: func(g *Agent) {
				db.Update(g.db, func(root *db.Root) error {
					guestAcctStr, _, _ := g.FindAccount("alice*starlight.com")
					var guestAcct fsm.AccountID
					err := guestAcct.SetAddress(guestAcctStr)
					if err != nil {
						return err
					}
					root.Agent().PutPrimaryAcct(&guestAcct)
					return nil
				})
			},
			want: errAcctsSame,
		},
		{
			name:       "agent not funded",
			guestAddr:  successGuestAddr,
			hostAmount: 1 * xlm.Lumen,
			host:       successHostAddr,
			agentFunc: func(g *Agent) {
				db.Update(g.db, func(root *db.Root) error {
					h := root.Agent().Wallet()
					h.Seqnum = 0
					root.Agent().PutWallet(h)
					return nil
				})
			},
			want: errNotFunded,
		},
		{
			name:       "insufficient balance",
			guestAddr:  successGuestAddr,
			hostAmount: 1 * xlm.Lumen,
			host:       successHostAddr,
			agentFunc: func(g *Agent) {
				db.Update(g.db, func(root *db.Root) error {
					h := root.Agent().Wallet()
					h.NativeBalance = xlm.Amount(2 * xlm.Lumen)
					root.Agent().PutWallet(h)
					return nil
				})
			},
			want: errInsufficientBalance,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := startTestAgent(t)
			defer g.CloseWait()
			config := Config{
				Username:   "alice",
				Password:   "password",
				HorizonURL: testHorizonURL,
			}

			err := g.ConfigInit(&config, c.host)
			if err != nil {
				t.Fatal(err)
			}

			// Initialize Host wallet.
			err = db.Update(g.db, func(root *db.Root) error {
				h := root.Agent().Wallet()
				h.Seqnum = 1
				h.NativeBalance = 50 * xlm.Lumen
				root.Agent().PutWallet(h)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if c.agentFunc != nil {
				c.agentFunc(g)
			}
			_, got := g.DoCreateChannel(c.guestAddr, c.hostAmount)
			if errors.Root(got) != c.want {
				t.Errorf("g.DoCreateChannel(%s, %s) = %s, want %s", c.guestAddr, c.hostAmount, got, c.want)
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()

	config := Config{
		Username:   "alice",
		Password:   "password",
		HorizonURL: testHorizonURL,
	}

	err := g.ConfigInit(&config, "")
	if err != nil {
		t.Fatal(err)
	}

	// Initialize Host wallet.
	err = db.Update(g.db, func(root *db.Root) error {
		h := root.Agent().Wallet()
		h.Seqnum = 1
		h.NativeBalance = 50 * xlm.Lumen
		root.Agent().PutWallet(h)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.DoCreateChannel("alice*starlight.com", xlm.Lumen)
	if err != nil {
		t.Fatal(err)
	}

	timedOut := make(chan struct{})
	finished := make(chan struct{})
	timer := time.AfterFunc(10*time.Second, func() {
		close(timedOut)
	})

	go func() {
		g.CloseWait()
		close(finished)
	}()

	select {
	case <-finished:
		timer.Stop()
	case <-timedOut:
		t.Error("timed out")
	}
}
