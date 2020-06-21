package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"gitlab.com/EbonJaeger/dolphin"
)

// Version is the version string of the program, set in the Makefile.
var Version string

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

	if opts.Version {
		fmt.Printf("mcdolphin version %s\n", Version)
		os.Exit(0)
	}

	dolphin.NewDolphin(opts)
}
