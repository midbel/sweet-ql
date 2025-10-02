package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func createFlag(name, help string) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(io.Discard)
	if help != "" {
		set.Usage = func() {
			fmt.Fprintln(os.Stderr, help)
			os.Exit(2)
		}
	}
	return set
}
