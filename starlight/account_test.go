package starlight

import (
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
		name   string
		target string
		want   want
	}{{
		name:   "success",
		target: "alice*starlight.com",
		want: want{
			accountID:    "GDSRO6H2YM6MC6ZO7KORPJXSTUMBMT3E7MZ66CFVNMUAULFG6G2OP32I",
			starlightURL: "https://starlight.com/",
			err:          nil,
		},
	}, {
		name:   "localhost success",
		target: "bob*localhost:7000",
		want: want{
			accountID:    "GAIPBPU6OC4JGYLQ4WI6LFYECMN43RVK3EI7N3TL3CVVM6MBIC2QART2",
			starlightURL: "http://localhost:7000/",
			err:          nil,
		},
	}, {
		name:   "federation address does not exist",
		target: "doesnotexist*starlight.com",
		want: want{
			accountID:    "",
			starlightURL: "",
			err:          errBadHttpStatus,
		},
	}, {
		name:   "federation address ill-formed ",
		target: "alice",
		want: want{
			accountID:    "",
			starlightURL: "",
			err:          errBadAddress,
		},
	}, {
		name:   "federation address email",
		target: "invalid@address.com",
		want: want{
			accountID:    "",
			starlightURL: "",
			err:          errBadAddress,
		},
	}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := startTestAgent(t)
			defer g.CloseWait()
			config := Config{
				Username:   "alice",
				Password:   "password",
				HorizonURL: testHorizonURL,
			}

			err := g.ConfigInit(&config)
			if err != nil {
				t.Error(err)
			}

			accountID, starlightURL, err := g.FindAccount(c.target)
			if errors.Root(err) != c.want.err {
				t.Errorf("Error finding %s: got %s, want %s", c.target, err, c.want.err)
			}
			if accountID != c.want.accountID {
				t.Errorf("Error finding %s account ID: got %s, want %s", c.target, accountID, c.want.accountID)
			}
			if starlightURL != c.want.starlightURL {
				t.Errorf("Error finding %s starlight URL: got %s, want %s", c.target, starlightURL, c.want.starlightURL)
			}
		})
	}
}
