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

const (
	identifierNoStar      = "identifier.star"
	identifierNoDuplicate = "identifier.duplicate"
	identifierOnlyName    = "identifier.name"
	aliasNoAlias          = "aliasing.alias"
	aliasSelf             = "aliasing.self"
	aliasMissing          = "aliasing.missing"
	aliasInvalid          = "aliasing.invalid"
	aliasUndefined        = "aliasing.undefined"
	aliasUnused           = "aliasing.unused"
	aliasRecommanded      = "aliasing.recommanded"
	cteSelect             = "cte.select"
	cteNoCte              = "cte.nocte"
	cteUnused             = "cte.unused"
	cteShadow             = "cte.shadow"
)

var supportedRules = map[string]func(Severity) Rule{
	identifierNoStar:           NoStar,
	identifierOnlyName:         OnlyName,
	identifierNoDuplicate:      DuplicatedName,
	"unqualified-name":         UnqualifiedName,
	cteSelect:                  CteOnlySelect,
	cteNoCte:                   NoCte,
	cteUnused:                  CteUnused,
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
	aliasRecommanded:           RecommandedAlias,
	aliasMissing:               MissingAlias,
	aliasNoAlias:               NoAlias,
	aliasSelf:                  SelfAlias,
	aliasInvalid:               InvalidAlias,
	aliasUndefined:             UndefinedAlias,
	aliasUnused:                UnusedAlias,
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
	"no-returning":             NoReturning,
}

type RuleOption func(Rule) error

func WithCount(count int) RuleOption {
	return func(r Rule) error {
		if count < 0 {
			return fmt.Errorf("negative limit not allowed")
		}
		if s, ok := r.(interface{ setLimit(int) }); ok {
			s.setLimit(count)
		}
		return nil
	}
}

func WithClause(clause string) RuleOption {
	return func(r Rule) error {
		return nil
	}
}

func WithMinLength(n int) RuleOption {
	return func(r Rule) error {
		return nil
	}
}

func WithMaxLength(n int) RuleOption {
	return func(r Rule) error {
		return nil
	}
}

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
		if s, ok := r.(interface{ setSeverity(Severity) }); ok {
			s.setSeverity(sev)
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
	lint := NewLinter(rules)
	return lint.Lint(r)
}

func LintDefault(r io.Reader) ([]Issue, error) {
	lint := DefaultLinter()
	return lint.Lint(r)
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
	if len(rules) == 0 {
		return DefaultLinter()
	}
	i := Linter{
		rules: rules,
	}
	return &i
}

func (i *Linter) Lint(r io.Reader) ([]Issue, error) {
	p, err := parser.NewParser(r)
	if err != nil {
		return nil, err
	}
	var list []Issue
	for {
		stmt, err := p.Parse()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		issues, err := i.lint(stmt)
		if err != nil {
			return nil, err
		}
		query := p.Query()
		for i := range issues {
			issues[i].Query = query
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func (i *Linter) lint(stmt ast.Node) ([]Issue, error) {
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

var ErrIssuesLimit = errors.New("too many issues detected")

type rule struct {
	ast.Visitor
	name     string
	severity Severity
	issues   []Issue
	count    int
}

func stdRule(visit ast.Visitor, name string, level Severity) *rule {
	return &rule{
		Visitor:  visit,
		severity: level,
		name:     name,
	}
}

func (r *rule) Name() string {
	return r.name
}

func (r *rule) Verify(stmt ast.Node) ([]Issue, error) {
	defer r.reset()

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return slices.Clone(r.issues), err
}

func (r *rule) Report(stmt ast.Node, reason string) error {
	iss := Issue{
		Position: stmt.Pos(),
		Severity: r.severity,
		Rule:     r.Name(),
		Reason:   reason,
	}
	return r.Add(iss)
}

func (r *rule) Add(iss Issue) error {
	if r.count > 0 && len(r.issues) >= r.count {
		return ErrIssuesLimit
	}
	r.issues = append(r.issues, iss)
	return nil
}

func (r *rule) setSeverity(level Severity) {
	r.severity = level
}

func (r *rule) setLimit(count int) {
	r.count = count
}

func (r *rule) reset() {
	r.issues = r.issues[:0]
}
