package lint

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/lang/parser"
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

type noStar struct {
	severity Severity
}

func NoStar(level Severity) Rule {
	return noStar{
		severity: level,
	}
}

func (r noStar) Name() string {
	return "no-star"
}

func (r noStar) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noStar) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = issues
	case ast.SelectStatement:
		issues, err := r.checkStar(q)
		if err != nil {
			return nil, err
		}
		list = issues
	default:
	}
	return list, nil
}

func (r noStar) checkStar(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, c := range q.Columns {
		n, ok := c.(ast.Name)
		if !ok {
			continue
		}
		if n.Name() == "*" {
			i := Issue{
				Position: n.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "use explicit column names instead of '*'",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

type noCte struct {
	severity Severity
}

func NoCte(level Severity) Rule {
	return noCte{
		severity: level,
	}
}

func (r noCte) Verify(stmt ast.Statement) ([]Issue, error) {
	if w, ok := stmt.(ast.WithStatement); ok {
		i := Issue{
			Position: w.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use subqueries instead of cte",
		}
		return []Issue{i}, nil
	}
	return nil, nil
}

func (_ noCte) Name() string {
	return "no-cte"
}

type cteDuplicate struct {
	severity Severity
}

func CteDuplicate(level Severity) Rule {
	return cteDuplicate{
		severity: level,
	}
}

func (r cteDuplicate) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var (
		names = make(map[string]struct{})
		list  []Issue
	)
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if _, ok := names[c.Ident]; ok {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "cte use name already defined",
			}
			list = append(list, i)
		}
		names[c.Ident] = struct{}{}
	}
	return list, nil
}

func (_ cteDuplicate) Name() string {
	return "cte-duplicate"
}

type cteUnused struct {
	severity Severity
}

func CteUnused(level Severity) Rule {
	return cteUnused{
		severity: level,
	}
}

func (r cteUnused) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var (
		positions = make(map[string]token.Position)
		names     = make(map[string]int)
	)
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		names[c.Ident] = 0
		positions[c.Ident] = c.Position
	}
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			continue
		}
		used, _ := r.getNames(c.Statement)
		for _, n := range used {
			if _, ok := names[n]; !ok {
				continue
			}
			names[n]++
		}
	}
	used, _ := r.getNames(q.Statement)
	for _, n := range used {
		if _, ok := names[n]; !ok {
			continue
		}
		names[n]++
	}
	var list []Issue
	for n, c := range names {
		if c > 0 {
			continue
		}
		i := Issue{
			Position: positions[n],
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "cte is defined but not used",
		}
		list = append(list, i)
	}
	return list, nil
}

func (_ cteUnused) getNames(stmt ast.Statement) ([]string, error) {
	q, ok := stmt.(ast.SelectStatement)
	if !ok {
		return nil, nil
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
			return ""
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
	return slices.Collect(maps.Keys(names)), nil
}

func (_ cteUnused) Name() string {
	return "cte-unused"
}

type cteColumns struct {
	severity Severity
}

func CteColumns(level Severity) Rule {
	return cteColumns{
		severity: level,
	}
}

func (r cteColumns) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var list []Issue
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if len(c.Columns) == 0 {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "missing explicit columns definition list for cte",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (_ cteColumns) Name() string {
	return "cte-columns"
}

type cteColumnsCount struct {
	severity Severity
}

func CteColumnsCount(level Severity) Rule {
	return cteColumnsCount{
		severity: level,
	}
}

func (r cteColumnsCount) Verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var list []Issue
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		e, ok := c.Statement.(ast.SelectStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if len(c.Columns) > 0 && len(e.Columns) != len(c.Columns) {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "invalid number of columns declared in cte",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (_ cteColumnsCount) Name() string {
	return "cte-columns-count"
}

type noSubquery struct {
	severity Severity
}

func NoSubquery(level Severity) Rule {
	return noSubquery{
		severity: level,
	}
}

func (r noSubquery) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noSubquery) verify(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	switch q := stmt.(type) {
	case ast.WithStatement:
		for _, q := range q.Queries {
			c, ok := q.(ast.CteStatement)
			if !ok {

			}
			issues, err := r.verify(c)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		issues, err := r.verify(q.Statement)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	case ast.CteStatement:
		return r.verify(q.Statement)
	case ast.SelectStatement:
		return r.checkSubquery(q)
	default:
	}
	return list, nil
}

func (r noSubquery) checkSubquery(stmt ast.SelectStatement) ([]Issue, error) {
	var (
		get  func(ast.Statement) ast.Statement
		list []Issue
	)

	get = func(q ast.Statement) ast.Statement {
		switch x := q.(type) {
		case ast.Join:
			return get(x.Table)
		case ast.Alias:
			return get(x.Statement)
		case ast.Group:
			return x.Statement
		default:
			return q
		}
	}
	for i := range stmt.Tables {
		q := get(stmt.Tables[i])
		if e, ok := q.(ast.SelectStatement); ok {
			i := Issue{
				Position: e.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "prefer using cte instead of subquery",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (_ noSubquery) Name() string {
	return "no-subquery"
}

type groupbyColumns struct {
	severity Severity
}

func GroupbyColumns(level Severity) Rule {
	return groupbyColumns{
		severity: level,
	}
}

func (r groupbyColumns) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ groupbyColumns) Name() string {
	return "groupby-columns"
}

type missingAlias struct {
	severity Severity
}

func MissingAlias(level Severity) Rule {
	return missingAlias{
		severity: level,
	}
}

func (r missingAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ missingAlias) Name() string {
	return "missing-alias"
}

type noAlias struct {
	severity Severity
}

func NoAlias(level Severity) Rule {
	return noAlias{
		severity: level,
	}
}

func (r noAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ noAlias) Name() string {
	return "no-alias"
}

type invalidAlias struct {
	severity Severity
}

func InvalidAlias(level Severity) Rule {
	return invalidAlias{
		severity: level,
	}
}

func (r invalidAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ invalidAlias) Name() string {
	return "invalid-alias"
}

type undefinedAlias struct {
	severity Severity
}

func UndefinedAlias(level Severity) Rule {
	return undefinedAlias{
		severity: level,
	}
}

func (r undefinedAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ undefinedAlias) Name() string {
	return "undefined-alias"
}

type noIdentQuoted struct {
	severity Severity
}

func NoIdentQuoted(level Severity) Rule {
	return noIdentQuoted{
		severity: level,
	}
}

func (r noIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ noIdentQuoted) Name() string {
	return "no-ident-quoted"
}

type missingIdentQuoted struct {
	severity Severity
}

func MissingIdentQuoted(level Severity) Rule {
	return missingIdentQuoted{
		severity: level,
	}
}

func (r missingIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
}

func (_ missingIdentQuoted) Name() string {
	return "missing-ident-quoted"
}
