package github

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type transport struct {
	prefix    string
	token     string
	accept    string
	transport http.RoundTripper
}

// RoundTripper returns a new round-tripper
// for GitHub API requests.
// It makes requests only to github.com and its subdomains.
// The behavior can be configured with Option values.
func RoundTripper(opt ...Option) http.RoundTripper {
	t := &transport{
		prefix:    "/",
		accept:    "application/vnd.github+json",
		transport: http.DefaultTransport,
	}
	for _, o := range opt {
		o(t)
	}
	return t
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = copyReq(req)
	req.URL = copyURL(req.URL)
	req.Header = copyHeader(req.Header)

	req.URL.Scheme = "https"
	if req.URL.Host == "" {
		req.URL.Host = "api.github.com"
	}
	if !isGitHubHost(req.URL.Host) {
		return nil, errors.New("bad host " + req.URL.Host)
	}
	if !path.IsAbs(req.URL.Path) {
		req.URL.Path = path.Join(t.prefix, req.URL.Path)
	}

	if _, ok := req.Header["Accept"]; !ok {
		req.Header.Set("Accept", t.accept)
	}
	if _, ok := req.Header["Authorization"]; !ok && t.token != "" {
		req.Header.Set("Authorization", "token "+t.token)
	}

	return t.transport.RoundTrip(req)
}

// Option values configure the behavior of objects
// created by this package.
// See New and RoundTripper.
type Option func(*transport)

// Prefix sets the default path prefix.
// For any request req with a relative URL path,
// the RoundTripper will use path.Join(s, req.URL.Path).
func Prefix(s string) Option {
	return func(t *transport) {
		t.prefix = s
	}
}

// Org sets the default path prefix to /repos/name.
// See also Prefix.
func Org(name string) Option {
	return Prefix(path.Join("/repos", name))
}

// Repo sets the default path prefix to /repos/org/repo.
// See also Prefix.
func Repo(org, repo string) Option {
	return Prefix(path.Join("/repos", org, repo))
}

// Token adds an HTTP "Authorization: token" header field
// to all outgoing requests that don't already have this
// header set.
func Token(s string) Option {
	return func(t *transport) {
		t.token = s
	}
}

// Accept sets the HTTP Accept header field to s for
// all outgoing requests that don't already have it set.
func Accept(s string) Option {
	return func(t *transport) {
		t.accept = s
	}
}

func isGitHubHost(s string) bool {
	return s == "github.com" || strings.HasSuffix(s, ".github.com")
}

func copyReq(req1 *http.Request) *http.Request {
	req2 := new(http.Request)
	*req2 = *req1
	return req2
}

func copyURL(url1 *url.URL) *url.URL {
	url2 := new(url.URL)
	*url2 = *url1
	return url2
}

func copyHeader(h1 http.Header) http.Header {
	h2 := make(http.Header)
	for k, v := range h1 {
		h2[k] = v
	}
	return h2
}
