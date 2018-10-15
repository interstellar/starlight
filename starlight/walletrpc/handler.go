package walletrpc

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kr/session"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/starlight"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/xlm"
)

type wallet struct {
	agent *starlight.Agent
	sess  session.Config
}

// Handler returns a handler that serves a React/JavaScript app
// for user interaction with g.
// The returned handler also forwards requests as appropriate
// to g's peer handler.
func Handler(g *starlight.Agent) http.Handler {
	wt := &wallet{agent: g}
	wt.sess.Secure = true
	wt.sess.HTTPOnly = true
	wt.sess.MaxAge = 14 * 24 * time.Hour

	// NOTE(kr): don't persist the session key across restarts.
	// We need the user to enter their password on startup to
	// enable private-key operations (such as executing rounds
	// of the protocol) that need to happen automatically to
	// keep channels open. So just generate a fresh session key
	// in memory for each new process.
	wt.sess.Keys = append(wt.sess.Keys, genKey())

	mux := new(http.ServeMux)
	mux.HandleFunc("/", index)
	mux.Handle("/starlight/", g.PeerHandler())
	mux.Handle("/federation", g.PeerHandler())
	mux.Handle("/.well-known/stellar.toml", g.PeerHandler())

	// Wallet RPCs. Add more here as necessary.
	mux.Handle("/api/updates", wt.auth(wt.updates))
	mux.Handle("/api/config-edit", wt.auth(wt.configEdit))
	mux.Handle("/api/logout", wt.auth(wt.logout))
	mux.Handle("/api/do-create-channel", wt.auth(wt.doCreateChannel))
	mux.Handle("/api/do-wallet-pay", wt.auth(wt.doWalletPay))
	mux.Handle("/api/do-command", wt.auth(wt.doCommand))
	mux.HandleFunc("/api/login", wt.login)
	mux.HandleFunc("/api/config-init", wt.configInit)
	mux.HandleFunc("/api/status", wt.status)

	return mux
}

func index(w http.ResponseWriter, req *http.Request) {
	// TODO(kr): serve index.html react app
	io.WriteString(w, "hello! this is where index.html goes!")
}

func (wt *wallet) updates(w http.ResponseWriter, req *http.Request) {
	// This is a handler for a standard long-polling event loop in
	// the client. The frontend will call GET /updates repeatedly, each
	// time supplying the index of the next event it's waiting for.
	// E.g. the first time it'll be From=1, then if it gets 3 updates
	// in reply, the next time it'll be From=4, and so on.
	var v struct{ From uint64 }
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		httperror(req, w, fmt.Sprintf("bad request: %s", err.Error()), 400)
		return
	}
	ctx := req.Context()

	// must be lower than the global write timeout (15s)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	wt.agent.WaitUpdate(ctx, v.From)
	ev := wt.agent.Updates(v.From, v.From+100)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ev)
}

func (wt *wallet) configEdit(w http.ResponseWriter, req *http.Request) {
	var config starlight.Config
	err := json.NewDecoder(req.Body).Decode(&config)
	if err != nil {
		httperror(req, w, fmt.Sprintf("bad request: %s", err.Error()), 400)
		return
	}
	err = wt.agent.ConfigEdit(&config)
	if err != nil {
		// TODO(kr): distinguish 5xx/4xx.
		// For now, just blame everything on the client.
		httperror(req, w, err.Error(), 400)
		return
	}
}

func (wt *wallet) configInit(w http.ResponseWriter, req *http.Request) {
	var config starlight.Config
	err := json.NewDecoder(req.Body).Decode(&config)
	if err != nil {
		httperror(req, w, fmt.Sprintf("bad request: %s", err.Error()), 400)
		return
	}
	err = wt.agent.ConfigInit(context.TODO(), &config)
	if err != nil {
		// TODO(kr): distinguish 5xx/4xx.
		// For now, just blame everything on the client.
		httperror(req, w, err.Error(), 400)
		return
	}
	session.Set(w, &struct{}{}, &wt.sess)
}

func (wt *wallet) doCreateChannel(w http.ResponseWriter, req *http.Request) {
	var v struct {
		GuestAddr  string
		HostAmount xlm.Amount
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		httperror(req, w, fmt.Sprintf("bad request: %s", err.Error()), 400)
		return
	}
	err = wt.agent.DoCreateChannel(v.GuestAddr, v.HostAmount, req.Host)
	switch errors.Root(err) {
	case nil:
	case starlight.ErrExists:
		// StatusResetContent is used to designate non-retriable errors.
		// TODO(debnil): Find a more suitable status code if possible.
		httperror(req, w, err.Error(), http.StatusResetContent)
	default:
		httperror(req, w, err.Error(), 500)
	}
}

func (wt *wallet) doWalletPay(w http.ResponseWriter, req *http.Request) {
	var v struct {
		Dest   string
		Amount uint64
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		httperror(req, w, fmt.Sprintf("bad request: %s", err.Error()), 400)
		return
	}
	xlmAmount := xlm.Amount(v.Amount)
	err = wt.agent.DoWalletPay(v.Dest, xlmAmount)
	if err != nil {
		httperror(req, w, err.Error(), 500)
	}
}

func (wt *wallet) doCommand(w http.ResponseWriter, req *http.Request) {
	var v struct {
		ChannelID string
		Command   fsm.Command
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		httperror(req, w, fmt.Sprintf("bad request: %s", err.Error()), 400)
		return
	}
	err = wt.agent.DoCommand(v.ChannelID, &v.Command)
	if err != nil {
		httperror(req, w, err.Error(), 500)
	}
}

func httperror(req *http.Request, w http.ResponseWriter, err string, code int) {
	log.Printf("request to %s returned internal error %s, returned status code %d", req.URL.Path, err, code)
	http.Error(w, err, code)
	return
}

func (wt *wallet) status(w http.ResponseWriter, req *http.Request) {
	var status struct {
		IsConfigured bool
		IsLoggedIn   bool
	}
	status.IsConfigured = wt.agent.Configured()
	status.IsLoggedIn = session.Get(req, &struct{}{}, &wt.sess) == nil
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (wt *wallet) login(w http.ResponseWriter, req *http.Request) {
	var cred struct{ Username, Password string }
	err := json.NewDecoder(req.Body).Decode(&cred)
	if err != nil {
		httperror(req, w, "bad request", 400)
		return
	}

	// This also enables private-key operations as a side effect.
	// (The private key material is encrypted with the user's password.)
	ok := wt.agent.Authenticate(cred.Username, cred.Password)

	if !ok {
		httperror(req, w, "unauthorized", 401)
		return
	}
	wt.sess.MaxAge = 14 * 24 * time.Hour
	session.Set(w, &struct{}{}, &wt.sess)
}

func (wt *wallet) logout(w http.ResponseWriter, req *http.Request) {
	wt.sess.MaxAge = -1
	session.Set(w, &struct{}{}, &wt.sess)
}

func (wt *wallet) auth(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := session.Get(req, &struct{}{}, &wt.sess)
		if err != nil {
			httperror(req, w, "unauthorized", 401)
			return
		}
		f(w, req)
	})
}

func genKey() *[32]byte {
	b := new([32]byte)
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err) // don't try to operate with a bad RNG
	}
	return b
}
