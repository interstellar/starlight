package httptest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interstellar/starlight/net/http/httperror"
)

func Post(t testing.TB, ctx context.Context, h http.Handler, f *httperror.Formatter, path, body string, opts ...func(*http.Request)) *TestResponse {
	req := httptest.NewRequest("POST", path, bytes.NewReader([]byte(body)))
	req = req.WithContext(ctx)
	for _, o := range opts {
		o(req)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return &TestResponse{t, w, f}
}

func WithHeaders(h map[string]string) func(*http.Request) {
	return func(r *http.Request) {
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}

type TestResponse struct {
	t testing.TB
	w *httptest.ResponseRecorder
	f *httperror.Formatter
}

func (tr *TestResponse) Errored() bool {
	return tr.w.Code/100 != 2
}

func (tr *TestResponse) Error() httperror.Response {
	var resp httperror.Response
	err := json.NewDecoder(tr.w.Body).Decode(&resp)
	if err != nil {
		tr.t.Fatalf("error decoding response: %s", err)
	}
	return resp
}

func (tr *TestResponse) Decode(into interface{}) {
	err := json.NewDecoder(tr.Body()).Decode(into)
	if err != nil {
		tr.t.Fatalf("decoding response: %s", err)
	}
}

func (tr *TestResponse) Bytes() []byte {
	return tr.Body().Bytes()
}

func (tr *TestResponse) Body() *bytes.Buffer {
	if tr.Errored() {
		info := tr.Error().Info
		tr.t.Fatalf("unexpected error: %s: %s", info.SeqCode, info.Message)
	}
	return tr.w.Body
}

func (tr *TestResponse) CheckError(want, data error) {
	resp := tr.Error()
	wantInfo := tr.f.Errors[want]
	if resp.Info.SeqCode != wantInfo.SeqCode {
		tr.t.Fatalf("got error: %s, want error: %s", resp.Info.SeqCode, wantInfo.SeqCode)
	}
	if resp.Info.Message != wantInfo.Message {
		tr.t.Fatalf("got err msg: %s, want err msg: %s", resp.Info.Message, wantInfo.Message)
	}

	if data != nil {
		// TODO: make this generic
		wantInfo = tr.f.Errors[data]
		if resp.Data["actions"].([]interface{})[0].(map[string]interface{})["seq_code"] != wantInfo.SeqCode {
			tr.t.Fatalf("got data action error: %s, want data action error: %s", resp.Data["actions"].([]interface{})[0].(map[string]interface{})["seq_code"], wantInfo.SeqCode)
		}
	}
}
