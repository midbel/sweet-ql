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
	r.checkColumns(stmt)
	return nil
}

func (r *groupbyColumns) checkColumns(stmt ast.SelectStatement) {
	for _, c := range stmt.Columns {
		if a, ok := c.(ast.Alias); ok {
			c = a.Node
		}
		call, ok := c.(ast.Call)
		if ok && lang.IsAggregateFunc(call.GetIdent()) {
			continue
		}
		var pos token.Position
		switch n := c.(type) {
		case ast.Name:
			ok = slices.ContainsFunc(stmt.Groups, func(g ast.Node) bool {
				x, ok := g.(ast.Name)
				if !ok {
					return false
				}
				return slices.Equal(n.Parts, x.Parts)
			})
			pos = n.Position
		case ast.Value:
			continue
		case ast.Case:
		case ast.Call:
		default:
		}
		if !ok {
			i := Issue{
				Position: pos,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "column must be included in group by clause or used in an aggregate function",
			}
			r.issues = append(r.issues, i)
		}
	}
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
