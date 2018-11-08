package starlighttest

import (
	"flag"
	"os"

	"github.com/interstellar/starlight/starlight/log"
)

var (
	// HorizonURL is the testnet Horizon URL used for testing.
	HorizonURL = flag.String("horizon", "https://horizon-testnet.stellar.org/", "horizon URL")
	verbose    = flag.Bool("verbose", true, "log verbose debugging output")
	out        = flag.String("out", "", "filename to store log output")
)

func init() {
	flag.Parse()
	log.SetVerbose(*verbose)
	if *out != "" {
		file, err := os.Create(*out)
		if err != nil {
			log.Info("error setting log output", err)
			os.Exit(1)
		}
		log.SetOutput(file)
	}
}
