package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/midbel/sweet/internal/lang/lint"
	"github.com/midbel/sweet/internal/rules"
)

func runLint(args []string) error {
	var (
		set      = flag.NewFlagSet("lint", flag.ExitOnError)
		linter   = lint.NewLinter()
		showList bool
		dialect  string
		config   string
	)
	set.StringVar(&config, "config", "", "linter configuration")
	set.StringVar(&dialect, "dialect", "", "SQL dialect")
	set.BoolVar(&showList, "list", false, "show list of supported rules")
	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {

		}
		return err
	}

	if showList {
		printRules(linter.Rules())
		return nil
	}

	process := func(file string) ([]rules.LintMessage, error) {
		r, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		return linter.Lint(r)
	}
	for _, f := range set.Args() {
		list, err := process(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		for _, m := range list {
			fmt.Fprintf(os.Stdout, "%s (%s): %s", m.Rule, m.Severity, m.Message)
			fmt.Fprintln(os.Stdout)
		}
	}
	return nil
}

func printRules(infos []rules.LintInfo) {
	for _, i := range infos {
		enabled := "\u2717"
		if i.Enabled {
			enabled = "\u2713"
		}
		fmt.Printf("%s %s", enabled, i.Rule)
		fmt.Println()
	}
}
