package starlight

import (
	"errors"
	"net/http"
)

// Defines errors returned by the agent.
var (
	errAcctsSame         = errors.New("same host and guest acct address")
	errAgentClosing      = errors.New("agent in closing state: cannot process new commands")
	errAlreadyConfigured = errors.New("already configured")
	errBadAddress        = errors.New("bad address")
	errBadHTTPStatus     = errors.New("bad http status")
	errBadHTTPRequest    = errors.New("bad http request")
	errBadRequest        = errors.New("bad request")
	errDecoding          = errors.New("error decoding")
	errEmptyAddress      = errors.New("destination address not set")
	errEmptyAmount       = errors.New("amount not set")
	errEmptyConfigEdit   = errors.New("config edit fields not set")
	errEmptyAsset        = errors.New("asset field not set")
	errEmptyIssuer       = errors.New("issuer field not set")
	errExists            = errors.New("channel exists")
	// Channel exists, but will be cleaned up and so the error is retriable
	errChannelExistsRetriable = errors.New("channel exists in a setup state")
	errFetchingAccounts       = errors.New("error fetching accounts")
	errInsufficientBalance    = errors.New("insufficient balance")
	errInvalidAddress         = errors.New("invalid address")
	errInvalidAsset           = errors.New("invalid asset")
	errInvalidChannelID       = errors.New("invalid channel ID")
	errInvalidEdit            = errors.New("can only update password and horizon URL")
	errInvalidInput           = errors.New("invalid input")
	errInvalidPassword        = errors.New("invalid password")
	errInvalidUsername        = errors.New("invalid username")
	errNoChannelSpecified     = errors.New("channel not specified")
	errNoCommandSpecified     = errors.New("command not specified")
	errNotConfigured          = errors.New("not configured")
	errNotFunded              = errors.New("primary acct not funded")
	errPasswordsDontMatch     = errors.New("old password doesn't match")
	errRemoteGuestMessage     = errors.New("received RPC message from guest")
)

// WriteError formats an error with the correct message and status from
// the starlight error formatter and writes the result to w.
func WriteError(req *http.Request, w http.ResponseWriter, err error) {
	errorFormatter.write(req, w, err)
}
