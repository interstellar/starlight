package starlight

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net"
)

// FindAccount looks up the account ID and Starlight URL
// for the Stellar account named by target.
//
// The target must be a valid Stellar address, e.g.
//
//   kr*mywallet.example.com
//
// It will use the Stellar TOML file
// and the Stellar federation server protocol
// to look up the additional info.
func (g *Agent) FindAccount(target string) (accountID, starlightURL string, err error) {
	var host string
	federation := true

	i := strings.Index(target, "*")
	if i < 0 {
		var guest xdr.AccountId
		err := guest.SetAddress(target)
		if err != nil {
			err = errors.Sub(errBadAddress, err)
			return "", "", errors.Wrap(err, target)
		}
		acct, err := g.wclient.LoadAccount(target)
		if err != nil {
			err = errors.Sub(errBadAddress, err)
			return "", "", errors.Wrapf(err, "loading account %s", target)
		}
		if acct.HomeDomain == "" {
			return "", "", errors.Wrap(errBadAddress, "no home domain set")
		}
		host = acct.HomeDomain
		federation = false
	} else {
		host = target[i+1:]
	}

	// Get URLs from Stellar TOML configuration.
	// See https://www.stellar.org/developers/guides/concepts/stellar-toml.html.
	resp, err := g.httpclient.Get(protocol(host) + host + "/.well-known/stellar.toml")
	if err != nil {
		return "", "", errors.Sub(errBadHTTPRequest, err)
	}
	if resp.StatusCode/100 != 2 {
		return "", "", errors.Wrapf(errBadHTTPStatus, "got http status %d looking up TOML", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Sub(errBadHTTPRequest, err)
		return "", "", errors.Wrap(err, "reading TOML")
	}
	var stellarTOML struct {
		FedURL       string `toml:"FEDERATION_SERVER"`
		StarlightURL string `toml:"STARLIGHT_SERVER"`
	}
	err = toml.Unmarshal(body, &stellarTOML)
	if err != nil {
		err = errors.Sub(errDecoding, err)
		return "", "", errors.Wrap(err, "unmarshaling TOML")
	}
	if !federation {
		return target, stellarTOML.StarlightURL, nil
	}

	// Get account ID from federation server.
	// See https://www.stellar.org/developers/guides/concepts/federation.html.
	q := url.Values{
		"q":    {target},
		"type": {"name"},
	}
	resp, err = g.httpclient.Get(stellarTOML.FedURL + "?" + q.Encode())
	if err != nil {
		err = errors.Sub(errBadHTTPRequest, err)
		return "", "", errors.Wrapf(err, "getting account ID from %s", stellarTOML.FedURL)
	}
	if resp.StatusCode/100 != 2 {
		return "", "", errors.Wrapf(errBadHTTPStatus, "got http status %d", resp.StatusCode)
	}
	var acct struct {
		ID string `json:"account_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&acct)
	if err != nil {
		err = errors.Sub(errDecoding, err)
		return "", "", errors.Wrapf(err, "decoding account ID from %s", stellarTOML.FedURL)
	}
	return acct.ID, stellarTOML.StarlightURL, nil
}

// protocol returns the protocol identifier to be used for the
// given host. If it is a loopback, or localhost, address, then
// we use HTTP. Otherwise, we use HTTPS.
func protocol(host string) string {
	if net.IsLoopback(host) {
		return "http://"
	}
	return "https://"
}

// validateUsername returns whether the username matches the Stellar
// name requirements.
// See: https://www.stellar.org/developers/guides/concepts/federation.html
func validateUsername(username string) bool {
	f := func(r rune) bool {
		return unicode.IsSpace(r)
	}
	return utf8.ValidString(username) && !strings.ContainsAny(username, "<*,>") &&
		strings.IndexFunc(username, f) == -1
}
