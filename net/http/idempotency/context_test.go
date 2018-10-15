package idempotency

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testCommitment(*http.Request) []byte {
	return []byte{0x01, 0x02, 0x03, 0x04}
}

func TestGenerateKey(t *testing.T) {
	h := ContextHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if Key(req.Context()) == "" {
			t.Error("missing idempotency key")
		}
		if UserKey(req.Context()) == "" {
			t.Error("missing user idempotency key")
		}
	}), testCommitment)

	req, err := http.NewRequest("POST", "https://api.seq.com/foo/bar/build-transaction", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestClientToken(t *testing.T) {
	const requestBody = `{"client_token": "abc"}`

	h := ContextHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != requestBody {
			t.Errorf("req.Body = %q, want %q", string(b), requestBody)
		}
		if got := UserKey(req.Context()); got != "abc" {
			t.Errorf("UserKey(req.Context()) = %q, want %q", got, "abc")
		}
	}), testCommitment)

	req, err := http.NewRequest("POST", "https://api.seq.com/foo/bar/build-transaction", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGetClientToken(t *testing.T) {
	cases := []struct {
		blob string
		want string
	}{
		{`{"client_token":"key"}`, "key"},
		{`{"account_alias":"chain","client_token":"key"}`, "key"},
		{`{"account_alias":"chain","token":"key"}`, ""},
		{`{"token":"client_token"}`, ""},
		{`{"actions":{"type":"issue","client_token":"key"}}`, "key"},
		{`{"actions":{"type":"issue","amount":20,"client_token":"key"}}`, "key"},
	}

	for i, tc := range cases {
		var v interface{}
		dec := json.NewDecoder(strings.NewReader(tc.blob))
		dec.UseNumber()
		err := dec.Decode(&v)
		if err != nil {
			t.Errorf("test case %d got Read error: %v", i, err)
		}
		got := getClientToken(v)
		if got != tc.want {
			t.Errorf("getClientToken(%v) = %s, want %s", v, got, tc.want)
		}
	}
}
