package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/sweet/internal/lang/lint"
)

var supportedRules = map[string]func(lint.Severity) lint.Rule{
	"no-star":                  lint.NoStar,
	"duplicate-field":          lint.DuplicateField,
	"no-cte":                   lint.NoCte,
	"cte-columns":              lint.CteColumns,
	"cte-columns-count":        lint.CteColumnsCount,
	"cte-unused":               lint.CteUnused,
	"cte-duplicate":            lint.CteDuplicate,
	"cte-name":                 lint.CteName,
	"std-operator":             lint.StdOperator,
	"missing-where":            lint.MissingWhere,
	"order-with-offset":        lint.OrderWithOffset,
	"enforce-type":             lint.EnforceType,
	"enforce-fetch":            lint.EnforceFetch,
	"enforce-limit":            lint.EnforceFetch,
	"set-columns-count":        lint.SetColumnsCount,
	"set-offset-last":          lint.SetOffsetFetchLast,
	"set-order-last":           lint.SetOrderLast,
	"no-subquery":              lint.NoSubquery,
	"subquery-columns-count":   lint.SubqueryColumnsCount,
	"subquery-names":           lint.SubqueryNames,
	"recommand-use-alias":      lint.RecommandedAlias,
	"missing-alias":            lint.MissingAlias,
	"no-alias":                 lint.NoAlias,
	"self-alias":               lint.SelfAlias,
	"invalid-alias":            lint.InvalidAlias,
	"undefined-alias":          lint.UndefinedAlias,
	"identifier-without-quote": lint.MissingIdentQuoted,
	"identifier-with-quote":    lint.NoIdentQuoted,
	"recommand-use-quote":      lint.RecommandedQuote,
	"no-literal-join":          lint.NoLiteralJoin,
	"unused-join":              lint.JoinUnused,
	"groupby-columns":          lint.GroupbyColumns,
	"no-literal-groupby":       lint.NoLiteralGroupby,
}

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
		rule, severity, ok := strings.Cut(value, ":")

		level := lint.None
		if ok {
			if severity == "" || severity == "error" {
				level = lint.Error
			} else if severity == "warning" {
				level = lint.Warning
			}
		}
		fn, ok := supportedRules[rule]
		if !ok {
			return fmt.Errorf("%s: unknown/unsupported lint rule", rule)
		}
		rules = append(rules, fn(level))
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
	var (
		issues []lint.Issue
		err    error
	)
	if len(rules) == 0 {
		issues, err = lint.LintDefault(r)
	} else {
		issues, err = lint.Lint(r, rules)
	}
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
