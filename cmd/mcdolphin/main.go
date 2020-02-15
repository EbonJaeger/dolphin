package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"gitlab.com/EbonJaeger/dolphin"
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
