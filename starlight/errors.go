package starlight

import (
	"errors"
	"net/http"
)

// Defines errors returned by the agent.
var (
	errAcctsSame           = errors.New("same host and guest acct address")
	errAlreadyConfigured   = errors.New("already configured")
	errBadAddress          = errors.New("bad address")
	errBadHTTPStatus       = errors.New("bad http status")
	errBadHTTPRequest      = errors.New("bad http request")
	errBadRequest          = errors.New("bad request")
	errDecoding            = errors.New("error decoding")
	errEmptyAddress        = errors.New("destination address not set")
	errEmptyAmount         = errors.New("amount not set")
	errEmptyConfigEdit     = errors.New("config edit fields not set")
	errExists              = errors.New("channel exists")
	errFetchingAccounts    = errors.New("error fetching accounts")
	errInsufficientBalance = errors.New("insufficient balance")
	errInvalidAddress      = errors.New("invalid address")
	errInvalidChannelID    = errors.New("invalid channel ID")
	errInvalidEdit         = errors.New("can only update password and horizon URL")
	errInvalidPassword     = errors.New("invalid password")
	errInvalidUsername     = errors.New("invalid username")
	errNoChannelSpecified  = errors.New("channel not specified")
	errNoCommandSpecified  = errors.New("command not specified")
	errNotConfigured       = errors.New("not configured")
	errNotFunded           = errors.New("primary acct not funded")
	errPasswordsDontMatch  = errors.New("old password doesn't match")
)

// WriteError formats an error with the correct message and status from
// the starlight error formatter and writes the result to w.
func WriteError(req *http.Request, w http.ResponseWriter, err error) {
	errorFormatter.write(req, w, err)
}
