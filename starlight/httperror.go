package starlight

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/net/http/httpjson"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/log"
)

// TODO(vniu): refactor i10r.io/net/httperror to avoid
// code duplication.

// Defines the exported errors that are used by the Starlight
// wallet RPC handler.
var (
	ErrAuthFailed   = errors.New("authentication failed")
	ErrUnauthorized = errors.New("unauthorized")
	ErrUnmarshaling = errors.New("error unmarshaling input")
)

// response contains  a set of error codes to send to the user.
type response struct {
	HTTPStatus int    `json:"-"`
	Message    string `json:"message"`
	Retriable  bool   `json:"retriable"`
}

type formatter struct {
	Default response
	Errors  map[error]response
}

// errorFormatter takes error objects and formats them to be HTTP error
// codes with the correct status code and message set.
var errorFormatter formatter

func init() {
	errorFormatter.Default = response{
		HTTPStatus: 500,
		Message:    "Starlight internal server error",
		Retriable:  true,
	}
	errorFormatter.Errors = make(map[error]response)

	// Handler errors
	errorFormatter.add(ErrUnauthorized, 401, "invalid session cookie", true)
	errorFormatter.add(ErrAuthFailed, 401, "invalid login", false)
	errorFormatter.add(ErrUnmarshaling, 400, "invalid input", false)

	// General agent errors
	errorFormatter.add(errBadRequest, 400, "bad request", false)

	// Find account
	errorFormatter.add(errBadAddress, 400, "invalid Stellar address", false)
	errorFormatter.add(errBadHTTPRequest, 500, "failed http request", true)
	errorFormatter.add(errBadHTTPStatus, 500, "bad http status", true)
	errorFormatter.add(errDecoding, 500, "decoding error", true)

	// Commands
	errorFormatter.add(errNoChannelSpecified, 400, "no channel specified", false)
	errorFormatter.add(errNoCommandSpecified, 400, "no command specified", false)
	errorFormatter.add(errEmptyAddress, 400, "no address specified", false)
	errorFormatter.add(errEmptyAmount, 400, "no amount specified", false)
	errorFormatter.add(errEmptyAsset, 400, "no asset specified", false)
	errorFormatter.add(errInsufficientBalance, 400, "insufficient balance", true)
	errorFormatter.add(errEmptyIssuer, 400, "no issuer specified", false)
	errorFormatter.add(errAcctsSame, 400, "same host and guest accounts", false)
	errorFormatter.add(errNotFunded, 500, "agent not yet funded", true)
	errorFormatter.add(errInvalidAddress, 400, "invalid address", false)

	// Message errors
	errorFormatter.add(errExists, 400, "channel already exists", false)
	errorFormatter.add(errChannelExistsRetriable, 400, "channel already exists, in setting up state", true)
	errorFormatter.add(errInvalidChannelID, 400, "invalid channel ID", false)
	errorFormatter.add(errFetchingAccounts, 400, "error fetching sequence numbers for accounts", false)

	// Configuration
	errorFormatter.add(errAlreadyConfigured, 400, "already configured", false)
	errorFormatter.add(errInvalidAsset, 400, "invalid asset", false)
	errorFormatter.add(errInvalidInput, 400, "invalid input", false)
	errorFormatter.add(errInvalidPassword, 400, "invalid password", false)
	errorFormatter.add(errInvalidUsername, 400, "invalid username", false)
	errorFormatter.add(errInvalidEdit, 400, "invalid edit field", false)
	errorFormatter.add(errEmptyConfigEdit, 400, "empty configuration edit", false)
	errorFormatter.add(errNotConfigured, 500, "not configured", true)
	errorFormatter.add(errPasswordsDontMatch, 400, "passwords don't match", false)

	// FSM errors
	errorFormatter.add(fsm.ErrInvalidVersion, 400, "invalid message version", false)
	errorFormatter.add(fsm.ErrChannelExists, 400, "channel proposed already exists", false)
	errorFormatter.add(fsm.ErrUnusedSettleWithGuestSig, 400, "unused settle with guest sig", false)
	errorFormatter.add(fsm.ErrUnexpectedState, 400, "unexpected state", true)
	errorFormatter.add(fsm.ErrInsufficientFunds, 400, "insufficient funds", true)
}

func (f *formatter) write(req *http.Request, w http.ResponseWriter, err error) {
	f.log(req, err)

	resp := f.format(err)
	httpjson.Write(req.Context(), w, resp.HTTPStatus, resp)
}

func (f *formatter) format(err error) response {
	root := errors.Root(err)

	resp, ok := f.Errors[root]
	if !ok {
		resp = f.Default
	}
	return resp
}

// parse reads an error Response from the provided reader
func parse(r io.Reader) (*response, bool) {
	var resp response
	err := json.NewDecoder(r).Decode(&resp)
	if err != nil {
		return nil, false
	}
	return &resp, true
}

func (f *formatter) add(key error, httpStatus int, msg string, retry bool) {
	f.Errors[key] = response{
		HTTPStatus: httpStatus,
		Message:    msg,
		Retriable:  retry,
	}
}

func (f *formatter) log(req *http.Request, err error) {
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	// TODO(vniu): format the log message with more error details and data
	log.Debugf("request to %s returned internal error %s", req.URL.Path, errorMessage)
}
