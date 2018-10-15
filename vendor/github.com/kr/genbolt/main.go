package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var flagO = flag.String("o", "", "output `file` (default stdout)")

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: genbolt [-o output.go] [input.go]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	b, err := gen(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *flagO == "" {
		os.Stdout.Write(b)
	} else {
		err = ioutil.WriteFile(os.Args[2], b, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
