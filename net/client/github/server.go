package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/interstellar/starlight/log"
)

// Hook returns a handler that verifies
// each request has a valid X-Hub-Signature header
// for secret, before calling next.
//   http.Handle("/pr", github.Hook(secret, handler))
// Note that this handler buffers the entire request
// body into memory to validate the signature.
func Hook(secret string, next http.Handler) http.Handler {
	secretBytes := []byte(secret)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req = validateSig(req, secretBytes)
		if req == nil {
			// NOTE(kr): return 202 here, not 401 or any other non-2xx code.
			// See https://pubsubhubbub.github.io/PubSubHubbub/pubsubhubbub-core-0.4.html#authednotify.
			http.Error(w, "bad signature", http.StatusAccepted)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func validateSig(req *http.Request, secret []byte) *http.Request {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(req.Context(), err)
		return nil
	}

	s := req.Header.Get("X-Hub-Signature")
	if !strings.HasPrefix(s, "sha1=") {
		return nil
	}
	s = s[5:] // strip sha1= prefix

	got, _ := hex.DecodeString(s)
	h := hmac.New(sha1.New, secret)
	h.Write(body)
	if !hmac.Equal(h.Sum(nil), got) {
		return nil
	}

	req = copyReq(req)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	return req

}
