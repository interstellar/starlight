package starlighttest

import "flag"

var (
	// HorizonURL is the testnet Horizon URL used for testing.
	HorizonURL = flag.String("horizon", "https://horizon-testnet.stellar.org/", "horizon URL")
	debug      = flag.Bool("debug", false, "log verbose debugging output")
)

func init() {
	flag.Parse()
}

func SetDebug(d bool) {
	debug = &d
}
