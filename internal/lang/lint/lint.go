package lint

import (
	"errors"
	"io"
	"maps"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/lang/parser"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

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

type RuleOptions uint64

const (
	CheckFields RuleOptions = 1 << iota
	CheckTables
)

func (r RuleOptions) withCheckFields() bool {
	return r&CheckFields == CheckFields
}

func (r RuleOptions) withCheckTables() bool {
	return r&CheckFields == CheckTables
}

type Rule interface {
	Verify(ast.Statement) ([]Issue, error)
	Name() string
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
		CteColumns(Error),
		CteColumnsCount(Error),
		CteUnused(Warning),
		CteDuplicate(Error),
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

func (i *Linter) Lint(stmt ast.Statement) ([]Issue, error) {
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

type checkSelectFunc func(ast.SelectStatement) ([]Issue, error)

func verify(stmt ast.Statement, check checkSelectFunc) ([]Issue, error) {
	switch q := stmt.(type) {
	case ast.WithStatement:
		var (
			list []Issue
			all  = slices.Clone(q.Queries)
		)
		all = append(all, q.Statement)
		for _, q := range all {
			issues, err := verify(q, check)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		return list, nil
	case ast.CteStatement:
		return verify(q.Statement, check)
	case ast.Group:
		return verify(q.Statement, check)
	case ast.UnionStatement:
		return verifyList(slx.Make(q.Left, q.Right), check)
	case ast.IntersectStatement:
		return verifyList(slx.Make(q.Left, q.Right), check)
	case ast.ExceptStatement:
		return verifyList(slx.Make(q.Left, q.Right), check)
	case ast.SelectStatement:
		return check(q)
	default:
		return nil, nil
	}
}

func verifyList(stmts []ast.Statement, check checkSelectFunc) ([]Issue, error) {
	var list []Issue
	for _, s := range stmts {
		issues, err := verify(s, check)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func getNames2(q ast.Statement) [][]string {
	switch q := q.(type) {
	case ast.Name:
		var parts []string
		for i := range q.Parts {
			parts = append(parts, q.Parts[i].Name)
		}
		return slx.One(parts)
	case ast.Alias:
		return getNames2(q.Statement)
	case ast.Call:
		var list [][]string
		for i := range q.Args {
			list = slices.Concat(list, getNames2(q.Args[i]))
		}
		return list
	case ast.Binary:
		list := slices.Concat(getNames2(q.Left), getNames2(q.Right))
		return list
	default:
		return nil
	}
}

func getNames(q ast.Statement) []string {
	switch q := q.(type) {
	case ast.Name:
		var parts []string
		for i := range q.Parts {
			parts = append(parts, q.Parts[i].Name)
		}
		return parts
	case ast.Alias:
		return getNames(q.Statement)
	case ast.Call:
		var list []string
		for i := range q.Args {
			list = slices.Concat(list, getNames(q.Args[i]))
		}
		return list
	case ast.Binary:
		list := slices.Concat(getNames(q.Left), getNames(q.Right))
		return list
	default:
		return nil
	}
}

func getTables(stmt ast.Statement) []string {
	q, ok := stmt.(ast.SelectStatement)
	if !ok {
		return nil
	}
	var (
		get   func(ast.Statement) string
		names = make(map[string]struct{})
	)
	get = func(stmt ast.Statement) string {
		switch q := stmt.(type) {
		case ast.Join:
			return get(q.Table)
		case ast.Name:
			return q.Name()
		case ast.Alias:
			return get(q.Statement)
		case ast.Group:
			return get(q.Statement)
		default:
			return ""
		}
	}
	for _, t := range q.Tables {
		n := get(t)
		if n == "" {
			continue
		}
		names[n] = struct{}{}
	}
	return slices.Collect(maps.Keys(names))
}

func getPosition(stmt ast.Statement) token.Position {
	var pos token.Position
	switch q := stmt.(type) {
	case ast.Name:
		return q.Position
	case ast.Call:
		return q.Position
	case ast.Value:
		return q.Position
	case ast.Group:
		return getPosition(q.Statement)
	case ast.Binary:
		return q.Position
	default:
		return pos
	}
}

func getQueries(stmt ast.Statement) []ast.SelectStatement {
	if a, ok := stmt.(ast.Alias); ok {
		return getQueries(a.Statement)
	}
	if gs, ok := stmt.(interface{ GetStatement() []ast.Statement }); ok {
		var res []ast.SelectStatement
		for _, s := range gs.GetStatement() {
			res = slices.Concat(res, getQueries(s))
		}
		return res
	}
	q, ok := stmt.(ast.SelectStatement)
	if !ok {
		return nil
	}
	list := slx.One(q)
	for _, c := range q.Columns {
		list = slices.Concat(list, getQueries(c))
	}
	for _, t := range q.Tables {
		if j, ok := t.(ast.Join); ok {
			t = j.Table
		}
		list = slices.Concat(list, getQueries(t))
	}
	list = slices.Concat(list, getQueries(q.Where))
	list = slices.Concat(list, getQueries(q.Having))
	return list
}
