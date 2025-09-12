package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/midbel/sweet/internal/lang/lint"
)

func createLinterFromArgs(args []string) (*lint.Linter, []string, error) {
	linter, files, err := createLinterFromConfig(args)
	if err != nil {
		if !errors.Is(err, errConfig) {
			linter, files, err = createLinterFromOptions(args)
		}
	}
	return linter, files, err
}

func createLinterFromConfig(args []string) (*lint.Linter, []string, error) {
	return nil, nil, errConfig
}

func createLinterFromOptions(args []string) (*lint.Linter, []string, error) {
	var (
		set = flag.NewFlagSet("lint", flag.ContinueOnError)
		// count = set.Int("c", 0, "print n first issue(s)")
		// fix   = set.Bool("fix", false, "fix all errors/warning when possible")
		// level lint.Severity
		rules []lint.Rule
	)
	set.Func("r", "enable rule", func(value string) error {
		rule, severity, ok := strings.Cut(value, ":")

		level := lint.None
		if ok {
			if severity == "" || severity == "error" {
				level = lint.Error
			} else if severity == "warning" {
				level = lint.Warning
			}
		}
		check, err := lint.RuleByName(rule, level)
		if err == nil {
			rules = append(rules, check)
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		return nil, nil, err
	}
	return lint.NewLinter(rules), set.Args(), nil
}

func runLint(args []string) error {
	linter, files, err := createLinterFromArgs(args)
	if err != nil {
		return err
	}
	var found int
	for _, f := range files {
		issues, err := lintFile(linter, f)
		if err != nil {
			return err
		}
		for _, i := range issues {
			ReportIssue(i)
		}
		found += len(issues)
	}
	if found > 0 {
		return fmt.Errorf("%d issue(s) found in query", found)
	}
	return nil
}

func lintFile(linter *lint.Linter, file string) ([]lint.Issue, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return linter.Lint(r)
}
