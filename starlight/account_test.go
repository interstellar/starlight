package starlight

import (
	"net/http"
	"testing"

	"github.com/interstellar/starlight/errors"
)

func TestValidateUsername(t *testing.T) {
	cases := []struct {
		username string
		want     bool
	}{{
		username: "alice",
		want:     true,
	}, {
		username: "世界",
		want:     true,
	}, {
		username: "alice bob",
		want:     false,
	}, {
		username: string([]byte{0xff, 0xfe, 0xfd}),
		want:     false,
	}, {
		username: "alice*example.com",
		want:     false,
	}, {
		username: "<alice>",
		want:     false,
	}}

	for _, c := range cases {
		got := validateUsername(c.username)
		if got != c.want {
			t.Errorf("validateUsername(%s) = %t, want %t", c.username, got, c.want)
		}
	}
}

func TestFindAccount(t *testing.T) {
	type want struct {
		accountID    string
		starlightURL string
		err          error
	}
	cases := []struct {
		target string
		want   want
	}{{
		target: "alice*starlight.com",
		want: want{
			accountID:    "GDSRO6H2YM6MC6ZO7KORPJXSTUMBMT3E7MZ66CFVNMUAULFG6G2OP32I",
			starlightURL: "https://starlight.com/",
			err:          nil,
		},
	}, {
		target: "doesnotexist*starlight.com",
		want: want{
			accountID:    "",
			starlightURL: "",
			err:          errBadHttpStatus,
		},
	}, {
		target: "alice",
		want: want{
			accountID:    "",
			starlightURL: "",
			err:          errBadAddress,
		},
	}, {
		target: "invalid@address.com",
		want: want{
			accountID:    "",
			starlightURL: "",
			err:          errBadAddress,
		},
	}}
	client := &http.Client{
		Transport: agentHTTP{},
	}
	for _, c := range cases {
		accountID, starlightURL, err := findAccount(client, c.target)
		if errors.Root(err) != c.want.err {
			t.Errorf("Error finding %s: got %s, want %s", c.target, err, c.want.err)
			continue
		}
		if accountID != c.want.accountID {
			t.Errorf("Error finding %s account ID: got %s, want %s", c.target, accountID, c.want.accountID)
		}
		if starlightURL != c.want.starlightURL {
			t.Errorf("Error finding %s starlight URL: got %s, want %s", c.target, starlightURL, c.want.starlightURL)
		}
	}
}
