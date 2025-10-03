package main

import (
	"flag"
	"fmt"
	"io"
)

type usageError struct {
	Help string
	Err  error
}

func UsageError(help string, err error) error {
	return usageError{
		Help: help,
		Err:  err,
	}
}

func (u usageError) Error() string {
	return fmt.Sprintf("%s\n\n%s", u.Err, u.Help)
}

func createFlag(name, help string) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(io.Discard)
	return set
}
