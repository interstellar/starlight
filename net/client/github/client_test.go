package github

import (
	"net/url"
	"testing"
)

func TestUpdatePageURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"foo/bar", "foo/bar?page=2"},
		{"foo/bar?page=1", "foo/bar?page=2"},
		{"foo/bar?page=2", "foo/bar?page=3"},
	}

	for _, test := range cases {
		u, err := url.Parse(test.in)
		if err != nil {
			t.Fatal(err)
		}
		err = nextPage(u)
		if err != nil {
			t.Error(err)
			continue
		}
		got := u.String()
		if got != test.want {
			t.Errorf("nextPage(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}
