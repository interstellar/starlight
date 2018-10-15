package github

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestValidateSig(t *testing.T) {
	secret := []byte("hunter2")
	body := "hello"
	sig := "sha1=8e84af330dc7ad3cb2f29e832db2475af8ba07c6"

	got := validateSig(&http.Request{
		Body: ioutil.NopCloser(strings.NewReader(body)),
		Header: http.Header{
			"X-Hub-Signature": {sig},
		},
	}, secret)

	if got == nil {
		t.Error("sig check failed, want pass")
	}
}
