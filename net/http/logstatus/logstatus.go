package logstatus

import (
	"context"
	"net/http"

	"github.com/interstellar/starlight/log"
)

// Handler prints a log message for any HTTP response.
func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w = &logResult{ResponseWriter: w, ctx: req.Context(), path: req.URL.Path}
		h.ServeHTTP(w, req)
	})
}

type logResult struct {
	http.ResponseWriter // embedded for other methods
	wroteHeader         bool
	ctx                 context.Context
	path                string
}

func (lr *logResult) WriteHeader(code int) {
	lr.ResponseWriter.WriteHeader(code)
	log.Printkv(lr.ctx,
		"event", "response",
		"code", code,
		"path", lr.path,
	)
	lr.wroteHeader = true
}

func (lr *logResult) Write(p []byte) (int, error) {
	if !lr.wroteHeader {
		lr.WriteHeader(200)
	}
	return lr.ResponseWriter.Write(p)
}
