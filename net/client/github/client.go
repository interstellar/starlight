// Package github makes interacting with the GitHub API
// more convenient.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Client is a GitHub API client.
type Client struct {
	client *http.Client
}

// Open returns a new Client object
// with behavior determined by Option values.
func Open(opt ...Option) *Client {
	return &Client{
		client: &http.Client{Transport: RoundTripper(opt...)},
	}
}

// Deletef performs a DELETE request to the given URL.
func (c *Client) Deletef(format string, arg ...interface{}) error {
	_, err := c.rpc("DELETE", fmt.Sprintf(format, arg...), nil, nil)
	return err
}

// Getf performs a GET request to the given URL.
//
// It treats the response body according to the type of resp:
//   nil        discard
//   io.Writer  write body to resp
//   (other)    decode JSON into resp
func (c *Client) Getf(resp interface{}, format string, arg ...interface{}) error {
	_, err := c.rpc("GET", fmt.Sprintf(format, arg...), nil, resp)
	return err
}

// GetAllf performs a series of one or more GET requests
// to the given URL, to fetch all pages of results.
//
// It treats the response body for each page
// according to the type of resp:
//   nil        discard
//   io.Writer  write each page to resp
//   *slice     decode JSON and append to *resp
//   (other)    error
// If resp is JSON, it must be a pointer to a slice,
// and pages after the first will be appended to it.
func (c *Client) GetAllf(resp interface{}, format string, arg ...interface{}) error {
	var page interface{}
	appendPage := func() {}
	switch resp.(type) {
	case nil:
	case io.Writer:
		page = resp
	default:
		if rt := reflect.TypeOf(resp); rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice {
			return fmt.Errorf("bad resp type %T", resp)
		}
		rv := reflect.ValueOf(resp).Elem()
		page = reflect.New(rv.Type()).Interface()
		appendPage = func() {
			rv.Set(reflect.AppendSlice(rv, reflect.ValueOf(page).Elem()))
			page = reflect.New(rv.Type()).Interface()
		}
	}
	u, err := url.Parse(fmt.Sprintf(format, arg...))
	if err != nil {
		return err
	}
	for {
		hresp, err := c.rpc("GET", u.String(), nil, page)
		if err != nil {
			return err
		}
		appendPage()
		if !hasNextPage(hresp) {
			return nil
		}
		err = nextPage(u)
		if err != nil {
			return err
		}
	}
}

func hasNextPage(resp *http.Response) bool {
	// We look at the Link header field to determine
	// if there's a next page, but don't bother actually
	// parsing the (surprisingly complicated) format.
	// See https://tools.ietf.org/search/rfc8288 for a
	// description of what we're not doing.
	for _, s := range resp.Header["Link"] {
		if strings.Contains(s, `; rel=next`) || strings.Contains(s, `; rel="next"`) {
			return true
		}
	}
	return false
}

func nextPage(u *url.URL) error {
	// See https://developer.github.com/v3/#pagination
	// for specification of GitHub "page" query param.
	//
	// NOTE(kr): the doc linked above says not to do this:
	// "It's important to form calls with Link header values
	// instead of constructing your own URLs." However, it
	// offers no evidence or reasoning to justify this position,
	// and its suggested alternative is two orders of magnitude
	// more complicated with no discernable advantage.
	// (See both https://tools.ietf.org/search/rfc8288
	// and https://tools.ietf.org/html/rfc6570 to get
	// a sense of the complexity involved.)
	// It explicitly documents the format and meaning of the
	// "page" query param as a stable part of the API, so
	// there's no reason to think using it would be unsafe.
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	q.Set("page", strconv.Itoa(page+1))
	u.RawQuery = q.Encode()
	return nil
}

// Postf performs a POST request to the given URL.
//
// It sends the request body according to req's type:
//   nil         empty
//   io.Reader   req
//   url.Values  encode req as form data & set Content-Type
//   (other)     encode req as JSON      & set Content-Type
// It treats the response body according to resp's type:
//   nil        discard
//   io.Writer  write body to resp
//   (other)    decode JSON into resp
func (c *Client) Postf(req, resp interface{}, format string, arg ...interface{}) error {
	_, err := c.rpc("POST", fmt.Sprintf(format, arg...), req, resp)
	return err
}

// Putf performs a PUT request to the given URL.
//
// It sends the request body according to req's type:
//   nil         empty
//   io.Reader   req
//   url.Values  encode req as form data & set Content-Type
//   (other)     encode req as JSON      & set Content-Type
// It treats the response body according to resp's type:
//   nil        discard
//   io.Writer  write body to resp
//   (other)    decode JSON into resp
func (c *Client) Putf(req, resp interface{}, format string, arg ...interface{}) error {
	_, err := c.rpc("PUT", fmt.Sprintf(format, arg...), req, resp)
	return err
}

// Note that rpc closes the response body before returning.
func (c *Client) rpc(method, u string, req, resp interface{}) (*http.Response, error) {
	var r io.Reader
	var contentType string
	switch req := req.(type) {
	case nil:
	case io.Reader:
		r = req
	case url.Values:
		r = strings.NewReader(req.Encode())
		contentType = "application/x-www-form-urlencoded"
	default:
		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
		contentType = "application/json"
	}

	hreq, err := http.NewRequest(method, u, r)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		hreq.Header.Set("Content-Type", contentType)
	}

	hresp, err := c.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer hresp.Body.Close()
	if hresp.StatusCode/100 != 2 {
		return nil, StatusError(hresp.StatusCode)
	}
	switch body := resp.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(body, hresp.Body)
	default:
		err = json.NewDecoder(hresp.Body).Decode(body)
	}
	return hresp, err
}

// Do calls Do on the underlying HTTP client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
