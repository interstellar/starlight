package starlight

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"

	bolt "github.com/coreos/bbolt"

	"github.com/interstellar/starlight/starlight/db"
	"github.com/interstellar/starlight/starlight/internal/update"
	"github.com/interstellar/starlight/starlight/log"
)

var debug io.Writer = os.Stderr

// Update describes a change made to an agent's state.
type Update = update.Update

// WaitUpdate returns after an update at index i has occurred,
// or the ctx becomes done, whichever happens first.
func (g *Agent) WaitUpdate(ctx context.Context, i uint64) {
	go func() {
		<-ctx.Done()
		g.evcond.Broadcast()
	}()
	g.evcond.L.Lock()
	defer g.evcond.L.Unlock()
	for lastUpdateNum(g.db) < i && ctx.Err() == nil {
		g.evcond.Wait()
	}
}

// Updates returns all updates in the half-open interval [a, b).
// The returned slice will have length less than b-a
// if a or b is out of range.
func (g *Agent) Updates(a, b uint64) []*Update {
	updates := make([]*Update, 0) // we want json "[]" not "null"
	err := db.View(g.db, func(root *db.Root) error {
		bu := root.Agent().Updates().Bucket()
		if bu == nil {
			return nil
		}
		c := bu.Cursor()
		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, a)
		k, _ = c.Seek(k)
		for k != nil && binary.BigEndian.Uint64(k) < b {
			n := binary.BigEndian.Uint64(k)
			ev := root.Agent().Updates().Get(n)
			updates = append(updates, ev)
			k, _ = c.Next()
		}
		return nil
	})
	if err != nil {
		panic(err) // only errors here are bugs
	}
	return updates
}

// putUpdate assigns new values to ev.UpdateNum and ev.UpdateLedgerTime.
func (g *Agent) putUpdate(root *db.Root, ev *Update) {
	if ev.Account == nil {
		ev.Account = &update.Account{
			Balance: uint64(root.Agent().Wallet().Balance),
			ID:      root.Agent().PrimaryAcct().Address(),
		}
	}
	ev.UpdateLedgerTime = g.wclient.Now()
	root.Agent().Updates().Add(ev, &ev.UpdateNum)
	root.Tx().OnCommit(g.evcond.Broadcast)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(ev)
	log.Debug(string(b.Bytes()))
}

func lastUpdateNum(boltDB *bolt.DB) (n uint64) {
	err := db.View(boltDB, func(root *db.Root) error {
		if bu := root.Agent().Updates().Bucket(); bu != nil {
			n = bu.Sequence()
		}
		return nil
	})
	if err != nil {
		panic(err) // this indicates a bug; crash
	}
	return n
}
