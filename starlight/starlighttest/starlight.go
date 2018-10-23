// Package starlighttest contains agent-level integration tests for starlight

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

func StartServer(ctx context.Context, testdir, name string) *Starlightd {
	return start(nil, ctx, testdir, name)
}

func (s *Starlightd) Address() string {
	return s.address
}

func (s *Starlightd) Close() {
	s.g.Close() // TODO(bobg): This should be CloseWait, but that's much slower. Figure out why!
	s.server.Close()
}

func start(t *testing.T, ctx context.Context, testdir, name string) *Starlightd {
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
