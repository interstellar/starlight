package logstatus

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/interstellar/starlight/log"
)

func TestLogStatus(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stdout)

	emptyHandlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
	})
	handler := Handler(emptyHandlerFunc)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	status := strconv.Itoa(w.Result().StatusCode)
	path := "/foo"
	got := buf.String()

	if !strings.Contains(got, status) {
		t.Errorf("Log did not contain string:\ngot:  %s\nwant: %s", got, status)
	}

	if !strings.Contains(got, path) {
		t.Errorf("Log did not contain string:\ngot:  %s\nwant: %s", got, status)
	}

	t.Logf("got log: %v\n", got)
}
