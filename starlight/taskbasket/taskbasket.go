// Package taskbasket is a persistent task manager.
package taskbasket

import (
	"context"
	"crypto/rand"
	"sync"
	"time"

	bolt "github.com/coreos/bbolt"

	"github.com/interstellar/starlight/net"
)

// Task is an item in a TB.
// The TB runs the task via its Run method.
// If that returns an error,
// it is retried after an interval.
type Task interface {
	// Run runs the task once.
	// It is called repeatedly by a running taskbasket,
	// with exponential backoff and jitter,
	// until it returns nil.
	Run(context.Context) error
}

// Codec converts Tasks to and from byte slices.
type Codec interface {
	Encode(Task) ([]byte, error)
	Decode([]byte) (Task, error)
}

type pair struct {
	k []byte
	t Task
}

// TB is a taskbasket,
// a collection of abstract tasks that are persisted to a database.
// When the TB is running
// (see TB.Run)
// it launches each task in a goroutine that retries until the task succeeds,
// at which point it is removed from the database.
type TB struct {
	db     *bolt.DB
	bucket []byte
	codec  Codec
	ch     chan pair
	wg     *sync.WaitGroup
}

// New creates a new taskbasket.
// It launches goroutines for any tasks already existing in the db.
func New(ctx context.Context, db *bolt.DB, bucket []byte, codec Codec) (*TB, error) {
	var tb *TB
	err := db.Update(func(tx *bolt.Tx) error {
		var err error
		tb, err = NewTx(ctx, tx, db, bucket, codec)
		return err
	})
	return tb, err
}

// NewTx creates a new taskbasket in the context of an existing bolt Update transaction.
// It launches goroutines for any tasks already exiting in the db.
func NewTx(ctx context.Context, tx *bolt.Tx, db *bolt.DB, bucket []byte, codec Codec) (*TB, error) {
	tb := &TB{
		db:     db,
		bucket: bucket,
		codec:  codec,
		ch:     make(chan pair),
		wg:     new(sync.WaitGroup),
	}

	var tasks []pair

	bu, err := tx.CreateBucketIfNotExists(tb.bucket)
	if err != nil {
		return nil, err
	}

	err = bu.ForEach(func(k, v []byte) error {
		t, err := tb.codec.Decode(v)
		if err != nil {
			return err
		}
		tasks = append(tasks, pair{k: k, t: t})
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Initial launch of existing tasks.
	// Note,
	// this previously lived in TB.Run,
	// as a preamble to the main loop reading tasks from the channel.
	// But there was a race condition:
	// a goroutine that got in a call to Add _before_ the existing tasks were launched would find its task launched twice:
	// once via persistent storage and once via the channel.
	for _, p := range tasks {
		tb.wg.Add(1)
		go tb.runTask(ctx, p.k, p.t)
	}

	return tb, nil
}

// Add adds a task to the taskbasket.
// It is persisted to the database and then processed immediately.
// Note that if TB.Run has not been called,
// this function can block.
func (tb *TB) Add(t Task) error {
	return tb.db.Update(func(tx *bolt.Tx) error {
		return tb.AddTx(tx, t)
	})
}

// AddTx adds a task to the taskbasket in the context of an existing bolt Update transaction.
// It is persisted to the database and processed when the transaction commits.
// Note that if TB.Run has not been called,
// this function can block.
func (tb *TB) AddTx(tx *bolt.Tx, t Task) error {
	bits, err := tb.codec.Encode(t)
	if err != nil {
		return err
	}
	bu, err := tx.CreateBucketIfNotExists(tb.bucket)
	if err != nil {
		return err
	}
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		return err
	}
	err = bu.Put(key, bits)
	if err != nil {
		return err
	}
	tx.OnCommit(func() {
		tb.ch <- pair{k: key, t: t}
	})
	return nil
}

// Run runs forever, processing the tasks in a taskbasket.
// When it starts,
// it reads all existing tasks from persistent storage and launches a goroutine for each.
// Thereafter it waits for new tasks to arrive via Add.
// It returns if its context is canceled.
func (tb *TB) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case p := <-tb.ch:
			tb.wg.Add(1)
			go tb.runTask(ctx, p.k, p.t)
		}
	}
}

// Function runTask runs forever,
// retrying t.Run until it returns nil or until the context is canceled.
// If t.Run succeeds,
// it is removed from tb's task bucket.
func (tb *TB) runTask(ctx context.Context, key []byte, t Task) {
	defer tb.wg.Done()

	backoff := net.Backoff{Base: time.Second}
	for {
		err := t.Run(ctx)
		if err != nil {
			// Start this timer first,
			// so timing is as right as possible even if the db update takes long.
			timer := time.NewTimer(backoff.Next())

			// Write the possibly updated task back to the db.
			bits, err := tb.codec.Encode(t)
			if err != nil {
				panic(err)
			}
			// TODO(bobg): Test whether bits actually have changed and skip db write if so.
			err = tb.db.Update(func(tx *bolt.Tx) error {
				bu := tx.Bucket(tb.bucket)
				return bu.Put(key, bits)
			})
			if err != nil {
				panic(err)
			}

			select {
			case <-ctx.Done():
				timer.Stop()
				return

			case <-timer.C:
				continue
			}
		}
		err = tb.db.Update(func(tx *bolt.Tx) error {
			bu := tx.Bucket(tb.bucket)
			return bu.Delete(key)
		})
		if err != nil {
			panic(err)
		}
		return
	}
}
