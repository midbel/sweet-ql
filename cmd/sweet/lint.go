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
		count = set.Int("c", 0, "print n first issue(s)")
		level lint.Severity
		rules []lint.Rule
	)
	set.Func("l", "level", func(value string) error {
		switch value {
		case "all", "":
			level = lint.None | lint.Warning | lint.Error
		case "warning":
			level = lint.Warning
		case "error":
			level = lint.Error
		default:
		}
		return nil
	})
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
	var curr int
	for _, i := range issues {
		if level != 0 && level < i.Severity {
			continue
		}
		curr++
		if *count != 0 && curr >= *count {
			break
		}
		ReportIssue(i)
	}
	if len(issues) > 0 {
		return fmt.Errorf("%d issue(s) found in query", len(issues))
	}
	return nil
}
