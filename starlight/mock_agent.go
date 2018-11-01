package starlight

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	bolt "github.com/coreos/bbolt"
	"github.com/stellar/go/network"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/worizon"
)

var testDB = "./testdb"

// StartTestnetAgent starts an agent for testing
// purposes, but with requests made to a live
// testnet Horizon.
// TODO(bobg): The main starlight package should not export testing code.
func StartTestnetAgent(ctx context.Context, t *testing.T, dbpath string) (*Agent, *worizon.Client) {
	db, err := bolt.Open(filepath.Join(dbpath), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := StartAgent(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	return g, &g.wclient
}

func startTestAgent(t *testing.T) *Agent {
	err := os.RemoveAll(testDB)
	if err != nil {
		t.Fatal(err)
	}
	db, err := bolt.Open(filepath.Join(testDB), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := StartAgent(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	g.wclient = *worizon.NewClient(horizonHTTP{}, &worizon.FakeHorizonClient{})
	g.httpclient.Transport = agentHTTP{}
	return g
}

type agentHTTP struct{}

func (a agentHTTP) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.Host {
	case "starlight.com", "localhost:7000":
		if req.URL.Path == "/.well-known/stellar.toml" {
			return mockToml(req)
		}
		if req.URL.Path == "/federation" {
			return mockFederation(req)
		}
		if req.URL.Path == "/starlight/message" {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBufferString("ok")),
			}, nil
		}
	case "friendbot.stellar.org":
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
		}, nil
	}
	return &http.Response{
		StatusCode: 404,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBufferString("not found")),
	}, nil
}

func mockToml(req *http.Request) (*http.Response, error) {
	var bytes bytes.Buffer
	v := struct{ Origin string }{protocol(req.Host) + req.Host}
	tomlTemplate.Execute(&bytes, v)
	header := make(http.Header)
	header.Add("Access-Control-Allow-Origin", "*")
	header.Add("Content-Type", "text/plain")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(&bytes),
	}, nil
}

func mockFederation(req *http.Request) (*http.Response, error) {
	if req.URL.Query().Get("type") != "name" {
		return &http.Response{
			StatusCode: http.StatusNotImplemented,
			Body:       ioutil.NopCloser(bytes.NewBufferString("not implemented")),
		}, nil
	}
	q := req.URL.Query().Get("q")
	switch q {
	case "alice*starlight.com":
		var acct xdr.AccountId
		acct.SetAddress("GDSRO6H2YM6MC6ZO7KORPJXSTUMBMT3E7MZ66CFVNMUAULFG6G2OP32I")
		buf, err := json.Marshal(map[string]string{
			"stellar_address": q + "*" + req.Host,
			"account_id":      acct.Address(),
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
		}, nil
	case "bob*starlight.com":
		var acct xdr.AccountId
		acct.SetAddress("GB7YAPZ43APNVOVF5RMDGFWMNUB6ACBMSVVBSZDFXBL6MIFKAMOOYP65")
		buf, err := json.Marshal(map[string]string{
			"stellar_address": q + "*" + req.Host,
			"account_id":      acct.Address(),
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
		}, nil
	case "bob*localhost:7000":
		var acct xdr.AccountId
		acct.SetAddress("GAIPBPU6OC4JGYLQ4WI6LFYECMN43RVK3EI7N3TL3CVVM6MBIC2QART2")
		buf, err := json.Marshal(map[string]string{
			"stellar_address": q + "*" + req.Host,
			"account_id":      acct.Address(),
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
		}, nil
	default:
		return &http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewBufferString("not found")),
		}, nil
	}
}

type horizonHTTP struct{}

func (h horizonHTTP) RoundTrip(req *http.Request) (*http.Response, error) {
	// WARNING: this software is not compatible with Stellar mainnet.
	if req.Host == "horizon-testnet.stellar.org" || req.Host == "new-horizon-testnet.stellar.org" {
		if (req.URL.Path == "" && req.Method == "GET") || (req.URL.Path == "/transactions" && req.Method == "POST") {
			buf, err := json.Marshal(map[string]string{
				"horizon_version":    "test",
				"network_passphrase": network.TestNetworkPassphrase,
			})
			if err != nil {
				panic(err)
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBuffer(buf)),
			}, nil
		}
	}
	return nil, errors.New("not implemented")
}
