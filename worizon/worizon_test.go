package worizon

import (
	"testing"

	"github.com/interstellar/starlight/worizon/worizontest"
)

func TestNow(t *testing.T) {
	wor := NewClient(nil, &worizontest.FakeHorizonClient{})
	got := wor.Now()
	if got.IsZero() {
		t.Error("got zero time")
	}
}
