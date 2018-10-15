package httpjson

import (
	"context"
	"net/http"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// Keys for HTTP objects in Contexts.
// They are unexported; clients use Path and ResponseWriter
// instead of using these keys directly.
const (
	respKey key = iota
	pathKey
)

// ResponseWriter returns the HTTP response writer stored in ctx.
// If there is none, it panics.
// The context given to a handler function
// registered in this package is guaranteed to have
// a response writer.
func ResponseWriter(ctx context.Context) http.ResponseWriter {
	return ctx.Value(respKey).(http.ResponseWriter)
}

// Path returns the HTTP request path stored in ctx.
// If there is none, it panics.
// The context given to a handler function
// registered in this package is guaranteed to have
// a request.
func Path(ctx context.Context) string {
	return ctx.Value(pathKey).(string)
}

// WithPath returns a context with a path stored in it.
func WithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, pathKey, path)
}
