package lint

import (
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/lang/parser"
	"github.com/midbel/sweet/internal/token"
)

var supportedRules = map[string]func(Severity) Rule{
	"no-star":                  NoStar,
	"only-name":                OnlyName,
	"duplicated-name":          DuplicatedName,
	"ambiguous-name":           AmbiguousName,
	"no-cte":                   NoCte,
	"cte-unused":               CteUnused,
	"cte-name":                 CteNames,
	"cte-exposed-name":         CteExposedNames,
	"std-operator":             StdOperator,
	"self-compare":             SelfCompare,
	"missing-where":            MissingWhere,
	"order-with-offset":        OrderWithOffset,
	"enforce-type":             EnforceType,
	"enforce-fetch":            EnforceFetch,
	"enforce-limit":            EnforceFetch,
	"columns-count":            ColumnsCount,
	"columns-names":            ColumnsNames,
	"set-offset-last":          SetOffsetFetchLast,
	"set-order-last":           SetOrderLast,
	"no-subquery":              NoSubquery,
	"subquery-columns-count":   SubqueryColumnsCount,
	"subquery-names":           SubqueryNames,
	"recommand-use-alias":      RecommandedAlias,
	"missing-alias":            MissingAlias,
	"no-alias":                 NoAlias,
	"self-alias":               SelfAlias,
	"ambiguous-alias":          AmbiguousAlias,
	"invalid-alias":            InvalidAlias,
	"undefined-alias":          UndefinedAlias,
	"unused-alias":             UnusedAlias,
	"identifier-without-quote": MissingIdentQuoted,
	"identifier-with-quote":    NoIdentQuoted,
	"recommand-use-quote":      RecommandedQuoted,
	"no-literal-join":          NoLiteralJoin,
	"unused-join":              JoinUnused,
	"groupby-columns":          GroupbyColumns,
	"no-literal-groupby":       NoLiteralGroupby,
	"groupby-distinct":         GroupbyDistinct,
	"grouby-aggr-func":         GroupbyAggrFunc,
	"having-aggr-func":         HavingAggrFunc,
}

type RuleOption func(Rule) error

func WithSeverity(level string) RuleOption {
	return func(r Rule) error {
		var sev Severity
		switch level {
		case "off", "none", "":
			sev = None
		case "warning", "warn":
			sev = Warning
		case "error":
			sev = Error
		default:
			return fmt.Errorf("%s: unknown severity level", level)
		}
		if s, ok := r.(interface{ setLevel(Severity) }); ok {
			s.setLevel(sev)
		}
		return nil
	}
}

type Severity int8

const (
	None Severity = 1 << iota
	Warning
	Error
)

func (s Severity) String() string {
	switch s {
	case None:
		return "off"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return ""
	}
}

type Issue struct {
	Query string
	token.Position

	Severity
	Rule   string
	Reason string
	Cause  string
}

type Rule interface {
	Verify(ast.Node) ([]Issue, error)
	Name() string
}

func RuleByName(name string, level Severity) (Rule, error) {
	fn, ok := supportedRules[name]
	if !ok {
		return nil, fmt.Errorf("%s: unknown/unsupported lint rule", name)
	}
	if fn == nil {
		return nil, fmt.Errorf("%s: rule not yet implemented", name)
	}
	return fn(level), nil
}

func RuleWith(name string, options ...RuleOption) (Rule, error) {
	r, err := RuleByName(name, Error)
	if err != nil {
		return nil, err
	}
	for _, o := range options {
		if err := o(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

type Linter struct {
	rules []Rule
}

func Lint(r io.Reader, rules []Rule) ([]Issue, error) {
	p, err := parser.NewParser(r)
	if err != nil {
		return nil, err
	}
	var (
		lint = NewLinter(rules)
		all  []Issue
	)
	for {
		stmt, err := p.Parse()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		issues, err := lint.Lint(stmt)
		if err != nil {
			return nil, err
		}
		query := p.Query()
		for i := range issues {
			issues[i].Query = query
		}
		all = slices.Concat(all, issues)
	}
	return all, nil
}

func LintDefault(r io.Reader) ([]Issue, error) {
	p, err := parser.NewParser(r)
	if err != nil {
		return nil, err
	}
	var (
		lint = DefaultLinter()
		all  []Issue
	)
	for {
		stmt, err := p.Parse()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		issues, err := lint.Lint(stmt)
		if err != nil {
			return nil, err
		}
		query := p.Query()
		for i := range issues {
			issues[i].Query = query
		}
		all = slices.Concat(all, issues)
	}
	return all, nil
}

func DefaultLinter() *Linter {
	rules := []Rule{
		NoStar(Error),
		CteUnused(Warning),
		NoSubquery(Warning),
		GroupbyColumns(Error),
	}
	return NewLinter(rules)
}

func NewLinter(rules []Rule) *Linter {
	i := Linter{
		rules: rules,
	}
	return &i
}

func (i *Linter) Lint(stmt ast.Node) ([]Issue, error) {
	var list []Issue
	for _, r := range i.rules {
		issues, err := r.Verify(stmt)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}
