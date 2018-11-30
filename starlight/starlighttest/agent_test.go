package starlighttest

import (
	"context"
	"path/filepath"
	"testing"

	bolt "github.com/coreos/bbolt"

	"github.com/interstellar/starlight/starlight"
)

// StartTestnetAgent starts an agent for testing
// purposes, but with requests made to a live
// testnet Horizon.
func StartTestnetAgent(ctx context.Context, t *testing.T, dbpath string) *starlight.Agent {
	db, err := bolt.Open(filepath.Join(dbpath), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := starlight.StartAgent(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	return g
}
