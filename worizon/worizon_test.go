package worizon

import (
	"testing"
)

func TestNow(t *testing.T) {
	wor := NewClient(nil, &FakeHorizonClient{})
	got := wor.Now()
	if got.IsZero() {
		t.Error("got zero time")
	}
}
