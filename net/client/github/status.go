package github

import (
	"fmt"
	"net/http"
)

type StatusError int

func (e StatusError) Error() string {
	return fmt.Sprintf("GitHub status: %d %s", e, http.StatusText(int(e)))
}
