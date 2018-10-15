package idempotency

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Header is the http header field name used to extract
// idempotency keys from requests.
const Header = "Idempotency-Key"

// contextKey is an unexported type for context keys defined in
// this package. This prevents collisions with context keys
// defined in other packages.
type contextKey int

const (
	// idempotencyKey is the key for idempotency keys in
	// Contexts. It is unexported; clients use Key and
	// WithKey instead of using this context key directly.
	idempotencyKey contextKey = iota

	// userIdempotencyKey is the key for the unmodified,
	// user-supplied idempotency key in Contexts.
	userIdempotencyKey
)

// WithKey returns a new Context that carries key.
func WithKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKey, key)
}

// WithUserKey returns a new Context that carries userKey.
func WithUserKey(ctx context.Context, userKey string) context.Context {
	return context.WithValue(ctx, userIdempotencyKey, userKey)
}

// Key returns the idempotency key carried
// in ctx, if any.
func Key(ctx context.Context) string {
	k, _ := ctx.Value(idempotencyKey).(string)
	return k
}

// UserKey returns the original, user-supplied idempotency
// key carried in ctx, if any. It should not be used directly
// for the purpose of idempotence—only for propagating to
// other processes. Most packages should use Key instead.
func UserKey(ctx context.Context) string {
	k, _ := ctx.Value(userIdempotencyKey).(string)
	return k
}

// ContextHandler wraps a Handler, reading idempotency keys
// from the 'Idempotency-Key' http header field and adding
// it to the request's context. If the request does not
// provide an idempotency key, a random one is generated.
//
// ContextHandler performs an HMAC over the idempotency key
// before adding it to the context. It uses the value returned
// by commitment as the HMAC key. This is used to scope
// idempotency keys to the client credentials, to mitigate
// attacks from compromised idempotency keys.
func ContextHandler(h http.Handler, commitment func(*http.Request) []byte) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		k := req.Header.Get(Header)
		body := req.Body // set on the new http.Request

		if k == "" {
			// Some older SDKs send an idempotency key
			// as a "client_token" field within a JSON
			// request body. Buffer the request body
			// and try to extract the client token.
			// TODO(howard): delete this block once all our clients adopt
			// the SDKs that send client token in the request header
			b, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}
			body = ioutil.NopCloser(bytes.NewReader(b))
			dec := json.NewDecoder(bytes.NewReader(b))
			dec.UseNumber()
			var v interface{}
			err = dec.Decode(&v)
			if err == nil {
				k = getClientToken(v)
			}
		}
		if k == "" {
			// If we still didn't find a client-provided
			// token, generate one.
			// 16 bytes is slightly more entropy
			// than a v4 uuid
			var generatedKey [16]byte
			_, err := rand.Read(generatedKey[:])
			if err != nil {
				panic(err)
			}
			k = base64.RawURLEncoding.EncodeToString(generatedKey[:])
		}

		hsh := hmac.New(sha256.New, commitment(req))
		io.WriteString(hsh, k)
		boundKey := base64.StdEncoding.EncodeToString(hsh.Sum(nil))

		ctx := req.Context()
		ctx = WithKey(ctx, boundKey)
		ctx = WithUserKey(ctx, k)
		req2 := req.WithContext(ctx)
		req2.Body = body
		h.ServeHTTP(rw, req2)
	})
}

func getClientToken(v interface{}) (token string) {
	switch t := v.(type) {
	case bool, string, json.Number, nil:
		return ""
	case []interface{}:
		for _, v := range t {
			token = getClientToken(v)
			if token != "" {
				return token
			}
		}
		return token
	case map[string]interface{}:
		for k, v := range t {
			if k == "client_token" {
				token, ok := v.(string)
				if ok && token != "" {
					return token
				}
			} else {
				token = getClientToken(v)
				if token != "" {
					return token
				}
			}
		}
		return token
	default:
		panic(fmt.Errorf("unknown json type %T", v))
	}
}
