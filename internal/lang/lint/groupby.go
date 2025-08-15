package lint

import (
	"errors"
	"slices"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
)

type groupbyColumns struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func GroupbyColumns(level Severity) Rule {
	return &groupbyColumns{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *groupbyColumns) Name() string {
	return "groupby-columns"
}

func (r *groupbyColumns) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *groupbyColumns) VisitSelect(stmt *ast.SelectStatement) error {
	if len(stmt.Groups) == 0 {
		return nil
	}
	for _, c := range stmt.Columns {
		if a, ok := c.(*ast.Alias); ok {
			c = a
		}
		var ok bool
		switch c := c.(type) {
		case *ast.Name:
			ok = r.exists(c, stmt)
		case *ast.Call:
			if lang.IsAggregateFunc(c.GetIdent()) {
				ok = true
			}
		default:
		}
		if !ok {
			i := Issue{
				Position: c.Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "column must be used in group by clause if not used in aggregate function",
			}
			r.issues = append(r.issues, i)
		}
	}
	return nil
}

func (r *groupbyColumns) exists(name *ast.Name, stmt *ast.SelectStatement) bool {
	return slices.ContainsFunc(stmt.Groups, func(n ast.Node) bool {
		if n, ok := n.(*ast.Name); ok {
			return slices.Equal(n.Parts, name.Parts)
		}
		return false
	})
}

// check that when group by clause is used and no aggregate functions are used in select clause, prefer select distinct
type groupbyDistinct struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func GroupbyDistinct(level Severity) Rule {
	return &groupbyDistinct{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *groupbyDistinct) Name() string {
	return "groupby-distinct"
}

func (r *groupbyDistinct) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *groupbyDistinct) VisitSelect(stmt *ast.SelectStatement) error {
	if len(stmt.Groups) == 0 {
		return nil
	}
	ok := slices.ContainsFunc(stmt.Columns, func(c ast.Node) bool {
		if a, ok := c.(*ast.Alias); ok {
			c = a.Node
		}
		if c, ok := c.(*ast.Call); ok && lang.IsAggregateFunc(c.GetIdent()) {
			return true
		}
		return false
	})
	if !ok && !stmt.Distinct {
		i := Issue{
			Position: stmt.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use distinct in select clause if no aggregate functions are used",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type noLiteralGroupby struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func NoLiteralGroupby(level Severity) Rule {
	return &noLiteralGroupby{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noLiteralGroupby) Name() string {
	return "no-literal-groupby"
}

func (r *noLiteralGroupby) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noLiteralGroupby) VisitSelect(stmt *ast.SelectStatement) error {
	for _, g := range stmt.Groups {
		if v, ok := g.(*ast.Value); ok && !v.Number() {
			i := Issue{
				Position: v.Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "use explicit columns name or expression from select clause in group by",
			}
			r.issues = append(r.issues, i)
		}
	}
	return nil
}

// check that when aggregate functions are used in select clause, group by list is not empty
type groupbyAggrFunc struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func GroupbyAggrFunc(level Severity) Rule {
	return &groupbyAggrFunc{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *groupbyAggrFunc) Name() string {
	return "groupby-aggr-function"
}

func (r *groupbyAggrFunc) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}
