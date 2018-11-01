package starlighttest

import (
	"flag"
	"os"
	"testing"
)

var (
	HorizonURL = flag.String("horizon", "https://horizon-testnet.stellar.org/", "horizon URL")
)

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}
