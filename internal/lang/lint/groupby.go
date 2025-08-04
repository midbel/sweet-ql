package lint

import (
	"errors"
	"slices"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
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

func (r *groupbyColumns) VisitSelect(stmt ast.SelectStatement) error {
	if len(stmt.Groups) == 0 {
		return nil
	}
	for _, c := range stmt.Columns {
		if a, ok := c.(ast.Alias); ok {
			c = a
		}
		var (
			ok  bool
			pos token.Position
		)
		switch c := c.(type) {
		case ast.Name:
			ok = r.exists(c, stmt)
			pos = c.Position
		case ast.Call:
			if lang.IsAggregateFunc(c.GetIdent()) {
				ok = true
			}
			pos = c.Position
		default:
			pos = getPosition(c)
		}
		if !ok {
			i := Issue{
				Position: pos,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "column must be used in group by clause if not used in aggregate function",
			}
			r.issues = append(r.issues, i)
		}
	}
	return nil
}

func (r *groupbyColumns) exists(name ast.Name, stmt ast.SelectStatement) bool {
	return slices.ContainsFunc(stmt.Groups, func(n ast.Node) bool {
		if n, ok := n.(ast.Name); ok {
			return slices.Equal(n.Parts, name.Parts)
		}
		return false
	})
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

func (r *noLiteralGroupby) VisitSelect(stmt ast.SelectStatement) error {
	for _, g := range stmt.Groups {
		if v, ok := g.(ast.Value); ok {
			i := Issue{
				Position: v.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "use explicit columns name or expression from select clause in group by",
			}
			r.issues = append(r.issues, i)
		}
	}
	return nil
}

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
