// Package starlighttest contains agent-level integration tests for starlight.
package starlighttest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/interstellar/starlight/starlight"
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
