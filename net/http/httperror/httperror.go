// Package httperror defines the format for HTTP error responses
// from Chain services.
package httperror

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"net/http"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/interstellar/starlight/database/pg"
	"github.com/interstellar/starlight/errors"
	"github.com/interstellar/starlight/log"
	"github.com/interstellar/starlight/net/http/httpjson"
)

var (
	ErrNotAuthenticated  = errors.New("not authenticated")
	ErrNotAuthorized     = errors.New("request is not authorized")
	ErrMalformedHeader   = errors.New("malformed http header value")
	ErrNotFound          = errors.New("not found")
	ErrTooManyRequests   = errors.New("concurrency limit exceeded")
	ErrDisconnect        = errors.New("client disconnected")
	ErrGenericDoNotRetry = errors.New("Chain API Error")
	ErrGenericRetry      = errors.New("Chain API Error")
	ErrV1                = errors.New("request version is not supported")
)

var expvarErrcodes *expvar.Map // count of responses for each CH??? error code

func init() {
	expvarErrcodes = expvar.NewMap("httperror-errcodes")
}

// RetryFlag indicates whether an error is an
// error that should be retried by the HTTP client.
type RetryFlag func(error) bool

var (
	// Retry indicates this error is temporary. The
	// client may retry an identical request in the
	// future and may not receive this same error.
	Retry RetryFlag = func(error) bool { return true }
	// DoNotRetry indicates that this error must be raised
	// to the application, because it is not expected
	// to resolve on its own or is otherwise unsafe
	// to automatically retry.
	DoNotRetry RetryFlag = func(error) bool { return false }
)

// ForwardChainError is a special error type that doesn't require
// any custom formatting. Its content is directly returned
// to the client.
type ForwardChainError struct {
	Content Response
}

func (e *ForwardChainError) Error() string {
	return e.Content.Info.Message
}

// Info contains a set of error codes to send to the user.
type Info struct {
	HTTPStatus  int       `json:"-"`
	ChainCode   string    `json:"code"`
	SeqCode     string    `json:"seq_code"`
	Message     string    `json:"message"`
	IsRetriable RetryFlag `json:"-"`
}

// Response defines the error response for a Chain error.
type Response struct {
	Info
	Detail    string                 `json:"detail,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Retriable bool                   `json:"retriable"`
	// Temporary is deprecated, superseded by Retriable. New
	// clients should use "retriable". For now, "retriable"
	// and "temporary" are identical.
	Temporary bool `json:"temporary"` // deprecated
}

// Parse reads an error Response from the provided reader.
func Parse(r io.Reader) (*Response, bool) {
	var resp Response
	err := json.NewDecoder(r).Decode(&resp)
	if err != nil || resp.SeqCode == "" {
		return nil, false
	}
	return &resp, true
}

type ReportFunc func(context.Context, error)

// Formatter defines rules for mapping errors to the Chain error
// response format.
type Formatter struct {
	Default    Info
	Errors     map[error]Info
	reportFunc ReportFunc
}

func NewFormatter(fn ReportFunc) *Formatter {
	f := new(Formatter)
	f.Default = Info{500, "CH000", "SEQ000", "Chain API Error", Retry}

	f.Errors = make(map[error]Info)
	f.Add(ErrGenericDoNotRetry, 500, 0, "Chain API Error", DoNotRetry)
	f.Add(ErrGenericRetry, 500, 0, "Chain API Error", Retry)
	f.Add(context.DeadlineExceeded, 500, 1, "Request timed out", Retry)
	f.Add(pg.ErrUserInputNotFound, 404, 2, "Not found", DoNotRetry)
	f.Add(ErrNotFound, 404, 2, "Not found", DoNotRetry)

	// ErrNotAuthenticated and ErrNotAuthorized return 404 errors
	// so that unauthenticated requests do not leak information,
	// for instance, whether or not a customer uses Sequence.
	f.Add(ErrNotAuthenticated, 404, 2, "Not found", DoNotRetry)
	f.Add(ErrNotAuthorized, 404, 2, "Not found", DoNotRetry)

	f.Add(httpjson.ErrBadRequest, 400, 3, "Invalid request body", DoNotRetry)
	f.Add(ErrMalformedHeader, 400, 4, "Invalid request header", DoNotRetry)
	f.Add(ErrV1, 400, 5, "SDK v2.0 or greater is required. Details: https://seq.com/docs/changelog", DoNotRetry)
	f.Add(ErrTooManyRequests, 429, 7, "Concurrency limit exceeded", DoNotRetry)
	f.Add(ErrDisconnect, 400, 10, "Client disconnected before response", Retry)

	f.reportFunc = fn
	return f
}

func (f *Formatter) Add(key error, httpStatus int, appCode int, msg string, retry RetryFlag) {
	f.Errors[key] = Info{
		HTTPStatus:  httpStatus,
		ChainCode:   fmt.Sprintf("CH%03d", appCode),
		SeqCode:     fmt.Sprintf("SEQ%03d", appCode),
		Message:     msg,
		IsRetriable: retry,
	}
}

// Format builds an error Response body describing err by consulting
// the f.Errors lookup table. If no entry is found, it returns f.Default.
func (f *Formatter) Format(ctx context.Context, err error) (body Response) {
	root := errors.Root(err)

	// Right now, the parent contexts are not canceled by any middleware.
	// The only way that the context is canceled is by the net/http package,
	// which occurs when the client's connection closes before the response
	// is returned.
	if root == context.Canceled {
		root = ErrDisconnect
	}

	// Some types cannot be used as map keys, for example slices.
	// If an error's underlying type is one of these, don't panic.
	// Just treat it like any other missing entry.
	defer func() {
		if err := recover(); err != nil {
			log.Printkv(ctx, log.Error, err)
			body = Response{f.Default, "", nil, true, true}
		}
	}()

	if err, ok := err.(*ForwardChainError); ok {
		return err.Content
	}

	info, ok := f.Errors[root]
	if !ok {
		info = f.Default
	}

	retriable := info.IsRetriable(err)
	body = Response{
		Info:      info,
		Detail:    errors.Detail(err),
		Data:      errors.Data(err),
		Retriable: retriable,
		Temporary: retriable,
	}
	return body
}

// Write writes a json encoded Response to the ResponseWriter.
// It uses the status code associated with the error.
//
// Write may be used as an ErrorWriter in the httpjson package.
func (f *Formatter) Write(ctx context.Context, w http.ResponseWriter, err error) {
	log.Helper()
	f.Log(ctx, err)
	resp := f.Format(ctx, err)
	if resp.SeqCode == "SEQ000" && f.reportFunc != nil {
		f.reportFunc(ctx, err)
	}
	if span, ok := tracer.SpanFromContext(ctx); ok {
		if resp.HTTPStatus/100 == 5 {
			span.SetTag("error", err)
		} else {
			span.SetTag("user-error", "yes")
		}
	}

	expvarErrcodes.Add(resp.SeqCode, 1)

	httpjson.Write(ctx, w, resp.HTTPStatus, resp)
}

// Log writes a structured log entry to the chain/log logger with
// information about the error and the HTTP response.
func (f *Formatter) Log(ctx context.Context, err error) {
	log.Helper()
	var errorMessage string
	if err != nil {
		// strip the stack trace, if there is one
		errorMessage = err.Error()
	}

	resp := f.Format(ctx, err)
	keyvals := []interface{}{
		"status", resp.HTTPStatus,
		"chaincode", resp.ChainCode,
		"seqcode", resp.SeqCode,
		log.KeyError, errorMessage,
	}
	if resp.HTTPStatus == 500 {
		keyvals = append(keyvals, log.KeyStack, errors.Stack(err))
	}
	log.Printkv(ctx, keyvals...)

	// Print the error Data, too
	// TODO(tessr): format this more nicely
	for k, v := range resp.Data {
		log.Printkv(ctx, k, v)
	}
}

// RecoverHandler wraps an http.Handler, adding
// recovery of panics. If a panic is recovered
// the error is formatted by f and written to the
// response writer.
//
// If the recovered value is http.ErrAbortHandler,
// the err is panicked again without writing.
func RecoverHandler(f *Formatter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", r)
			}
			if errors.Root(err) == http.ErrAbortHandler {
				panic(err)
			}
			log.Printkv(req.Context(), "message", "panic", "error", errors.Wrap(err))
			f.Write(req.Context(), rw, err)
		}()

		next.ServeHTTP(rw, req)
	})
}
