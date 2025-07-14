package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/sweet/internal/lang/lint"
)

func runLint(args []string) error {
	var (
		set   = flag.NewFlagSet("lint", flag.ExitOnError)
		rules []lint.Rule
	)
	set.Func("r", "enable rule", func(value string) error {
		_ = rules
		return nil
	})
	if err := set.Parse(args); err != nil {
		return err
	}

	var r io.Reader
	if f, err := os.Open(set.Arg(0)); err == nil {
		defer f.Close()
		r = f
	} else {
		r = strings.NewReader(set.Arg(0))
	}
	issues, err := lint.LintDefault(r)
	if err != nil {
		return err
	}
	for _, i := range issues {
		fmt.Println(i.Position, i.Severity, i.Rule, i.Reason)
		ReportIssue(i)
	}
	if len(issues) > 0 {

	}
	return nil
}
