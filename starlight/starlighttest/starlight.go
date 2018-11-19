// Package starlighttest contains agent-level integration tests for starlight.
package starlighttest

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/interstellar/starlight/starlight"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/walletrpc"
	"github.com/interstellar/starlight/worizon"
	"github.com/interstellar/starlight/worizon/xlm"
)

// Starlightd is an in-memory starlight agent with HTTP endpoints for protocol messages and UI commands.
type Starlightd struct {
	g             *starlight.Agent
	wclient       *worizon.Client
	handler       http.Handler
	server        *httptest.Server
	cookie        string
	address       string
	accountID     string
	nextUpdateNum uint64
	balance       xlm.Amount
	hostAmount    xlm.Amount
	guestAmount   xlm.Amount
}

// StartServer starts a Startlightd instance.
func StartServer(ctx context.Context, testdir, name string) *Starlightd {
	return start(ctx, nil, testdir, name)
}

// Address returns the (trimmed) URL of the Starlightd server.
func (s *Starlightd) Address() string {
	return s.address
}

// Close releases the resources associated with s.
func (s *Starlightd) Close() {
	s.g.Close() // TODO(bobg): This should be CloseWait, but that's much slower. Figure out why!
	s.server.Close()
}

func start(ctx context.Context, t *testing.T, testdir, name string) *Starlightd {
	g, wclient := starlight.StartTestnetAgent(ctx, t, fmt.Sprintf("%s/testdb_%s", testdir, name))
	s := &Starlightd{
		g:             g,
		wclient:       wclient,
		nextUpdateNum: 1,
	}
	s.handler = logWrapper(walletrpc.Handler(s.g), name)
	s.server = httptest.NewServer(s.handler)
	s.address = strings.TrimPrefix(s.server.URL, "http://")
	return s
}

func testServer(name string) *httptest.Server {
	mux := new(http.ServeMux)
	mux.HandleFunc("/starlight/message", testHandleMsg)
	mux.HandleFunc("/federation", testHandleFed)
	mux.HandleFunc("/.well-known/stellar.toml", testHandleTOML)
	mux.HandleFunc("/api/messages", testPollMessages)
	handler := logWrapper(mux, name)
	server := httptest.NewServer(handler)
	return server
}

func testPollMessages(w http.ResponseWriter, req *http.Request) {
	time.Sleep(8 * time.Second)
	var msgs []*fsm.Message
	json.NewEncoder(w).Encode(msgs)
	return
}

func testHandleMsg(w http.ResponseWriter, req *http.Request) {
	m := new(fsm.Message)
	err := json.NewDecoder(req.Body).Decode(m)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(m.ChannelID) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
}

func testHandleFed(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"stellar_address": "alice*" + req.Host,
		"account_id":      "GBOJVRYHEQRGBQDUT6B5C6HJYHVSY2LP65DRYJRXZWR2QZHTXFS3W4KL",
	})
}

func testHandleTOML(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/plain")
	v := struct{ Origin string }{req.Host}
	tomlTemplate := template.Must(template.New("toml").Parse(`
	FEDERATION_SERVER="http://{{.Origin}}/federation"
	STARLIGHT_SERVER="http://{{.Origin}}/"
	`))
	tomlTemplate.Execute(w, v)
}
