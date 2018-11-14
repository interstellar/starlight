package walletrpc

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/kr/session"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net"
	"github.com/interstellar/starlight/starlight"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/worizon/xlm"
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
	mux.Handle("/api/do-close-account", wt.auth(wt.doCloseAccount))
	mux.Handle("/api/do-command", wt.auth(wt.doCommand))
	mux.Handle("/api/find-account", wt.auth(wt.findAccount))
	// TODO(vniu): authenticate requests to the messages endpoint
	mux.HandleFunc("/api/messages", wt.messages)
	mux.HandleFunc("/api/login", wt.login)
	mux.HandleFunc("/api/config-init", wt.configInit)
	mux.HandleFunc("/api/status", wt.status)

	return mux
}

func index(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `
		<html>
			<script>
				let xmlhttp = new XMLHttpRequest();
				xmlhttp.onreadystatechange = function(){
					if (xmlhttp.readyState == 4 && xmlhttp.status == 200){
						document.open();
						document.write(xmlhttp.responseText);
						document.close();
					}
				}
				// If you want to make changes to the client and deploy them,
				// change this URL to point to where you've deployed your changes.
				// See ./sync-frontend.sh for how we're hosting the client.
				xmlhttp.open('GET', 'https://starlight-client.s3.amazonaws.com/v1/index.html', true);
				xmlhttp.send();
			</script>
		</html>
	`)
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
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	ctx := req.Context()

	// must be lower than the global write timeout (15s)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	wt.agent.WaitUpdate(ctx, v.From)
	// return max 100 updates at a time
	ev := wt.agent.Updates(v.From, v.From+100)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ev)
}

func (wt *wallet) configEdit(w http.ResponseWriter, req *http.Request) {
	var config starlight.Config
	err := json.NewDecoder(req.Body).Decode(&config)
	if err != nil {
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	err = wt.agent.ConfigEdit(&config)
	if err != nil {
		starlight.WriteError(req, w, err)
		return
	}
}

func (wt *wallet) configInit(w http.ResponseWriter, req *http.Request) {
	var config starlight.Config
	err := json.NewDecoder(req.Body).Decode(&config)
	if err != nil {
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	err = wt.agent.ConfigInit(&config, req.Host)
	if err != nil {
		starlight.WriteError(req, w, err)
		return
	}
	if net.IsLoopback(req.Host) {
		wt.sess.Secure = false
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
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	ch, err := wt.agent.DoCreateChannel(v.GuestAddr, v.HostAmount)
	switch errors.Root(err) {
	case nil:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ch)
	default:
		starlight.WriteError(req, w, err)
	}
}

func (wt *wallet) doWalletPay(w http.ResponseWriter, req *http.Request) {
	var v struct {
		Dest   string
		Amount uint64
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	xlmAmount := xlm.Amount(v.Amount)
	err = wt.agent.DoWalletPay(v.Dest, xlmAmount)
	if err != nil {
		starlight.WriteError(req, w, err)
	}
}

func (wt *wallet) doCloseAccount(w http.ResponseWriter, req *http.Request) {
	var v struct {
		Dest string
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	err = wt.agent.DoCloseAccount(v.Dest)
	if err != nil {
		starlight.WriteError(req, w, err)
	}
	wt.sess.MaxAge = -1
	session.Set(w, &struct{}{}, &wt.sess)
}

func (wt *wallet) doCommand(w http.ResponseWriter, req *http.Request) {
	var v struct {
		ChannelID string
		Command   fsm.Command
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	err = wt.agent.DoCommand(v.ChannelID, &v.Command)
	if err != nil {
		starlight.WriteError(req, w, err)
	}
}

func (wt *wallet) findAccount(w http.ResponseWriter, req *http.Request) {
	// TODO(debnil): Add unit test and needed framework for this and other wallet RPCs.
	var v struct {
		Addr string `json:"stellar_addr"`
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}
	var result struct {
		AcctID       string
		StarlightURL string
	}
	result.AcctID, result.StarlightURL, err = wt.agent.FindAccount(v.Addr)
	if err != nil {
		starlight.WriteError(req, w, err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (wt *wallet) messages(w http.ResponseWriter, req *http.Request) {
	var v struct {
		ChannelID string `json:"channel_id"`
		From      uint64
	}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}

	ctx := req.Context()

	// must be lower than the global write timeout (15s)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	wt.agent.WaitMsg(ctx, v.ChannelID, v.From)
	// return max 100 messages at a time
	msgs := wt.agent.Messages(v.ChannelID, v.From, v.From+100)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
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
		starlight.WriteError(req, w, errors.Sub(starlight.ErrUnmarshaling, err))
		return
	}

	// This also enables private-key operations as a side effect.
	// (The private key material is encrypted with the user's password.)
	ok := wt.agent.Authenticate(cred.Username, cred.Password)

	if !ok {
		starlight.WriteError(req, w, starlight.ErrAuthFailed)
		return
	}
	if net.IsLoopback(req.Host) {
		wt.sess.Secure = false
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
			starlight.WriteError(req, w, errors.Sub(starlight.ErrUnauthorized, err))
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
