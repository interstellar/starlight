package httpjson

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	wantRespHead := "baz"
	wantPath := "/foo"
	f := func(ctx context.Context) {
		if p := Path(ctx); p != wantPath {
			t.Errorf("Path(ctx) = %q, want %q", p, wantPath)
		}
		ResponseWriter(ctx).Header().Set("Test-Resp-Key", wantRespHead)
	}

	h, err := Handler(f, nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(resp, req)
	if g := resp.Header().Get("Test-Resp-Key"); g != wantRespHead {
		t.Errorf("header = %q want %q", g, wantRespHead)
	}
}
