package lint

import (
	"errors"
	"fmt"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type noCte struct {
	ast.Visitor
	issues   []Issue
	severity Severity
}

func NoCte(level Severity) Rule {
	return &noCte{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noCte) Name() string {
	return "no-cte"
}

func (r *noCte) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noCte) VisitWith(with ast.WithStatement) error {
	i := Issue{
		Position: with.Position,
		Severity: r.severity,
		Rule:     r.Name(),
		Reason:   "prefer using subqueries over common table expression",
	}
	r.issues = append(r.issues, i)
	return errStop
}

type cteDuplicate struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func CteDuplicate(level Severity) Rule {
	return &cteDuplicate{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *cteDuplicate) Name() string {
	return "cte-duplicate"
}

func (r *cteDuplicate) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *cteDuplicate) VisitWith(with ast.WithStatement) error {
	var names = make(map[string]struct{})
	for _, q := range with.Queries {
		q, ok := q.(ast.CteStatement)
		if !ok {
			return fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if _, ok := names[q.Ident]; ok {
			i := Issue{
				Position: q.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "duplicate identifier in with statement",
			}
			r.issues = append(r.issues, i)
		}
		names[q.Ident] = struct{}{}
	}
	return errStop
}

type cteUnused struct {
	ast.Visitor
	severity Severity

	names   map[string]int
	collect bool
}

func CteUnused(level Severity) Rule {
	return &cteUnused{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *cteUnused) Name() string {
	return "cte-unused"
}

func (r *cteUnused) Verify(stmt ast.Node) ([]Issue, error) {
	r.names = make(map[string]int)

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}

	var issues []Issue
	for _, c := range r.names {
		if c > 0 {
			continue
		}
		i := Issue{
			// Position: stmt.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "cte is defined but not used",
		}
		issues = append(issues, i)
	}

	return issues, err
}

func (r *cteUnused) VisitCte(stmt ast.CteStatement) error {
	r.names[stmt.Ident] = 0
	return nil
}

func (r *cteUnused) VisitSelect(stmt ast.SelectStatement) error {
	r.begin()
	defer r.end()

	sub := Walk(r)
	for _, t := range stmt.Tables {
		t.Accept(sub)
	}
	return nil
}

func (r *cteUnused) VisitName(name ast.Name) error {
	r.update(name.Name())
	return nil
}

func (r *cteUnused) begin() {
	r.collect = true
}

func (r *cteUnused) end() {
	r.collect = false
}

func (r *cteUnused) update(name string) {
	if !r.collect {
		return
	}
	if _, ok := r.names[name]; ok {
		r.names[name]++
	}
}

func (r cteUnused) verify(stmt ast.Node) ([]Issue, error) {
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
	var (
		all  = slices.Concat(q.Queries, slx.One(q.Node))
		list []Issue
	)
	for _, q := range all {
		if c, ok := q.(ast.CteStatement); ok {
			q = c.Node
		}
		r.checkTables(q, names)
	}
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

func (r cteUnused) checkTables(q ast.Node, names map[string]int) {
	for _, q := range getQueries(q) {
		for _, n := range getTables(q) {
			if _, ok := names[n]; !ok {
				continue
			}
			names[n]++
		}
	}
}

type cteColumns struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func CteColumns(level Severity) Rule {
	return &cteColumns{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *cteColumns) Name() string {
	return "cte-columns"
}

func (r *cteColumns) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *cteColumns) VisitWith(with ast.WithStatement) error {
	sub := Walk(r)
	for _, q := range with.Queries {
		err := q.Accept(sub)
		if err != nil {
			return err
		}
	}
	return errStop
}

func (r *cteColumns) VisitCte(cte ast.CteStatement) error {
	if len(cte.Columns) == 0 {
		i := Issue{
			Position: cte.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "specify column names explicitly in common table expression",
		}
		r.issues = append(r.issues, i)
	}
	return errStop
}

type cteColumnsCount struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func CteColumnsCount(level Severity) Rule {
	return &cteColumnsCount{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ cteColumnsCount) Name() string {
	return "cte-columns-count"
}

func (r *cteColumnsCount) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *cteColumnsCount) VisitWith(with ast.WithStatement) error {
	sub := Walk(r)
	for _, q := range with.Queries {
		err := q.Accept(sub)
		if err != nil && !errors.Is(err, errStop) {
			return err
		}
	}
	return errStop
}

func (r *cteColumnsCount) VisitCte(cte ast.CteStatement) error {
	var get func(ast.Node) (int, error)
	get = func(q ast.Node) (int, error) {
		switch c := q.(type) {
		case ast.SelectStatement:
			return len(c.Columns), nil
		case ast.UnionStatement:
			return get(c.Left)
		case ast.ExceptStatement:
			return get(c.Left)
		case ast.IntersectStatement:
			return get(c.Left)
		default:
			return 0, fmt.Errorf("%s: unexpected query type", r.Name())
		}
	}
	count, err := get(cte.Node)
	if err != nil {
		return err
	}
	if count != len(cte.Columns) {
		i := Issue{
			Position: cte.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "number of columns returned by select does not match number of columns declared by common table expression",
		}
		r.issues = append(r.issues, i)
	}
	return errStop
}

// check that fields qualified by cte are exposed by it
type cteName struct {
	severity Severity
}

func CteName(level Severity) Rule {
	return cteName{
		severity: level,
	}
}

func (_ cteName) Name() string {
	return "cte-name"
}

func (r cteName) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r cteName) verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}
