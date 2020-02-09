package main

import (
	"os"

	"github.com/EbonJaeger/dolphin"
	"github.com/jessevdk/go-flags"
)

func main() {
	var opts dolphin.Flags
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	dolphin.NewDolphin(opts)
}
