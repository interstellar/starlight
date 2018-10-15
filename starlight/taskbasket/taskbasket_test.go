package taskbasket

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"
	"time"

	bolt "github.com/coreos/bbolt"
)

const testBucket = "testbucket"

type testCodec struct {
	ch       chan<- *testTask
	failflag *bool
}

func (c testCodec) Encode(t Task) ([]byte, error) {
	return json.Marshal(t)
}

func (c testCodec) Decode(b []byte) (Task, error) {
	tt := &testTask{
		ch:       c.ch,
		failflag: c.failflag,
	}
	err := json.Unmarshal(b, tt)
	return tt, err
}

type testTask struct {
	Failures  int32
	Succeeded bool
	ch        chan<- *testTask
	failflag  *bool
}

func (tt *testTask) Run(ctx context.Context) error {
	defer func() { tt.ch <- tt }()

	if *tt.failflag {
		failures := atomic.AddInt32(&tt.Failures, 1)
		return fmt.Errorf("failure %d", failures)
	}
	tt.Succeeded = true
	return nil
}

func TestTaskbasket(t *testing.T) {
	f, err := ioutil.TempFile("", "TestTaskbasket")
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	f.Close()
	defer os.Remove(f.Name())

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}

	forceFailures := true
	ch := make(chan *testTask)
	codec := &testCodec{
		ch:       ch,
		failflag: &forceFailures,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tb, err := New(ctx, db, []byte(testBucket), codec)
	if err != nil {
		t.Fatal(err)
	}
	go tb.Run(ctx)

	task := &testTask{
		ch:       ch,
		failflag: &forceFailures,
	}

	err = tb.Add(task)
	if err != nil {
		t.Fatal(err)
	}

	timer := time.AfterFunc(10*time.Second, func() {
		t.Fatal("timed out waiting for task to process")
	})
	<-ch
	timer.Stop()

	timer = time.AfterFunc(10*time.Second, func() {
		t.Fatal("timed out waiting for canceled tasks to stop")
	})
	cancel()
	tb.wg.Wait()
	timer.Stop()

	if atomic.LoadInt32(&task.Failures) == 0 {
		t.Fatal("got 0 failures, want 1")
	}

	err = db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket([]byte(testBucket))
		if bu == nil {
			t.Fatalf("bucket %s does not exist", testBucket)
		}
		var found Task
		err := bu.ForEach(func(k, v []byte) error {
			if found != nil {
				t.Fatal("found unexpected item in bucket")
			}
			var err error
			found, err = codec.Decode(v)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		if found == nil {
			t.Fatal("persisted task not found")
		}
		ft := found.(*testTask)
		if ft.Failures != task.Failures {
			t.Errorf("got %d failure(s) in persisted task, want %d", ft.Failures, task.Failures)
		}
		if ft.Succeeded {
			t.Errorf("persisted task succeeded unexpectedly")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	forceFailures = false

	tb, err = New(ctx, db, []byte(testBucket), codec)
	if err != nil {
		t.Fatal(err)
	}
	go tb.Run(ctx)

	timer = time.AfterFunc(10*time.Second, func() {
		t.Fatal("timed out waiting for task to process")
	})
	newtt := <-ch
	timer.Stop()
	if newtt.Failures != task.Failures {
		t.Errorf("got failures=%d in resumed task, want %d", newtt.Failures, task.Failures)
	}
	if !newtt.Succeeded {
		t.Error("resumed task did not succeed")
	}

	timer = time.AfterFunc(10*time.Second, func() {
		t.Fatal("timed out waiting for canceled tasks to stop")
	})
	cancel()
	tb.wg.Wait()
	timer.Stop()

	err = db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket([]byte(testBucket))
		if bu == nil {
			t.Fatalf("bucket %s does not exist", testBucket)
		}
		var count int
		err := bu.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if count > 0 {
			t.Errorf("found %d item(s) in bucket, want 0", count)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
