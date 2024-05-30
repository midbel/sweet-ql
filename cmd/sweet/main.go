package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/midbel/sweet/internal/lang"
)

func main() {
	flag.Parse()

	var err error
	switch n, args := flag.Arg(0), flag.Args(); n {
	case "format", "fmt":
		err = runFormat(args[1:])
	case "lint", "check", "verify":
		err = runLint(args[1:])
	default:
		err = fmt.Errorf("unknown command %s", n)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func runFormat(args []string) error {
	var (
		set     = flag.NewFlagSet("format", flag.ExitOnError)
		writer  = lang.NewWriter(os.Stdout)
		dialect string
		config  string
	)
	set.StringVar(&config, "config", "", "formatter configuration")
	set.StringVar(&dialect, "dialect", "lang", "SQL dialect")
	set.BoolVar(&writer.Compact, "compact", writer.Compact, "produces compact SQL queries")
	set.BoolVar(&writer.UseAs, "use-as", writer.UseAs, "always use as to define alias")
	set.BoolVar(&writer.UseQuote, "use-quote", writer.UseQuote, "quote all identifier")
	set.IntVar(&writer.UseIndent, "use-indent", writer.UseIndent, "number of space to use to indent SQL")
	set.BoolVar(&writer.UseSpace, "use-space", writer.UseSpace, "use tabs instead of space to indent SQL")
	set.BoolVar(&writer.UseColor, "use-color", writer.UseColor, "colorify SQL keywords, identifiers")
	set.BoolVar(&writer.UseSubQuery, "use-subquery", writer.UseSubQuery, "replace cte by subqueries")
	set.BoolVar(&writer.UseCte, "use-cte", writer.UseCte, "replace subqueries by cte")
	set.BoolVar(&writer.UseCrlf, "use-crlf", writer.UseCrlf, "use crlf for newline")
	set.BoolVar(&writer.PrependComma, "prepend-comma", writer.PrependComma, "write comma before expressions")
	set.BoolVar(&writer.KeepComment, "keep-comment", writer.KeepComment, "keep comments")
	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {

		}
		return err
	}
	process := func(file string) error {
		r, err := os.Open(file)
		if err != nil {
			return err
		}
		defer r.Close()
		return writer.Format(r)
	}
	for _, f := range set.Args() {
		if err := process(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	return nil
}

func runLint(args []string) error {
	var (
		set     = flag.NewFlagSet("lint", flag.ExitOnError)
		linter  = lang.NewLinter()
		dialect string
		config  string
		init    bool
	)
	set.StringVar(&config, "config", "", "linter configuration")
	set.StringVar(&dialect, "dialect", "", "SQL dialect")
	set.BoolVar(&init, "init", false, "create linter configuration file")
	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {

		}
		return err
	}
	if init {
		return runInit()
	}
	process := func(file string) ([]lang.LintMessage, error) {
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

func runInit() error {
	return nil
}