package starlight

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/BurntSushi/toml"

	"github.com/interstellar/starlight/errors"
)

var (
	errBadAddress    = errors.New("bad address")
	errBadHttpStatus = errors.New("bad http status")
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
	i := strings.Index(target, "*")
	if i < 0 {
		// TODO(kr): check if target is a valid account pubkey and if so,
		// look up the "home domain" from the account's metadata on the
		// ledger.
		// See https://www.stellar.org/developers/guides/concepts/accounts.html#home-domain.
		return "", "", errors.Wrap(errBadAddress, target)
	}

	host := target[i+1:]

	// Get URLs from Stellar TOML configuration.
	// See https://www.stellar.org/developers/guides/concepts/stellar-toml.html.
	resp, err := g.httpclient.Get("https://" + host + "/.well-known/stellar.toml")
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode/100 != 2 {
		return "", "", errors.Wrapf(errBadHttpStatus, "http status %d looking up TOML", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "reading TOML")
	}
	var stellarTOML struct {
		FedURL       string `toml:"FEDERATION_SERVER"`
		StarlightURL string `toml:"STARLIGHT_SERVER"`
	}
	err = toml.Unmarshal(body, &stellarTOML)
	if err != nil {
		return "", "", errors.Wrap(err, "unmarshaling TOML")
	}

	// Get account ID from federation server.
	// See https://www.stellar.org/developers/guides/concepts/federation.html.
	q := url.Values{
		"q":    {target},
		"type": {"name"},
	}
	resp, err = g.httpclient.Get(stellarTOML.FedURL + "?" + q.Encode())
	if err != nil {
		return "", "", errors.Wrapf(err, "getting account ID from %s", stellarTOML.FedURL)
	}
	if resp.StatusCode/100 != 2 {
		return "", "", errors.Wrapf(errBadHttpStatus, "http status %d", resp.StatusCode)
	}
	var acct struct {
		ID string `json:"account_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&acct)
	if err != nil {
		return "", "", errors.Wrapf(err, "decoding account ID from %s", stellarTOML.FedURL)
	}
	return acct.ID, stellarTOML.StarlightURL, nil
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
