package starlighttest

import (
	"flag"
	"os"
	"testing"
)

var (
	HorizonURL = flag.String("horizon", "https://horizon-testnet.stellar.org/", "horizon URL")
)

// TODO(vniu): add logging flags
func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}
