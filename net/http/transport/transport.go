package transport

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// WithTLS is a helper function to combine helpful defaults with a
// tls.Config from the caller
func WithTLS(tls *tls.Config) *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSClientConfig:       tls,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
