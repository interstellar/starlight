// Package enclaveagent implements a client for communicating
// with the HTTP server exposed on each Enclave device.
package enclaveagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/interstellar/starlight/errors"
)

type Client struct {
	Addr      string
	UserAgent string
	// TODO: auth ?
}

type Error struct {
	StatusCode int
	Message    string
}

func (err Error) Error() string {
	return fmt.Sprintf("%s: %s", http.StatusText(err.StatusCode), err.Message)
}

type errorResponse struct {
	Message string
}

func (c *Client) post(ctx context.Context, path string, body, respBody interface{}) error {
	marshalledBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.Addr+path, bytes.NewReader(marshalledBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		err = dec.Decode(&errResp)
		if err != nil {
			return errors.Wrapf(err, "error decoding %s error", resp.Status)
		}
		return Error{StatusCode: resp.StatusCode, Message: errResp.Message}
	}

	err = dec.Decode(respBody)
	return errors.Wrapf(err, "decoding %s response into %T", path, respBody)
}
