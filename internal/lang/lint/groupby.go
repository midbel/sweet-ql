package lint

import (
	"fmt"
	"slices"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
)

type groupbyColumns struct {
	severity Severity
}

func GroupbyColumns(level Severity) Rule {
	return groupbyColumns{
		severity: level,
	}
}

func (_ groupbyColumns) Name() string {
	return "groupby-columns"
}

func (r groupbyColumns) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r groupbyColumns) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkGroupBy)
}

func (r groupbyColumns) checkGroupBy(stmt ast.SelectStatement) ([]Issue, error) {
	if len(stmt.Groups) == 0 {
		return nil, nil
	}
	var names []string
	for _, g := range stmt.Groups {
		n, ok := g.(ast.Name)
		if !ok {
			return nil, fmt.Errorf("%s: column name expected", r.Name())
		}
		names = append(names, n.Name())
	}

	var (
		list []Issue
		get  func(ast.Node) ast.Node
	)

	get = func(q ast.Node) ast.Node {
		switch q := q.(type) {
		case ast.Name:
			return q
		case ast.Alias:
			return get(q.Statement)
		case ast.Call:
			return q
		default:
		}
		return nil
	}

	for _, c := range stmt.Columns {
		n := get(c)
		switch c := n.(type) {
		case ast.Name:
			if !slices.Contains(names, c.Name()) {
				i := Issue{
					Position: c.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "column does not appear in group by",
				}
				list = append(list, i)
			}
		case ast.Call:
			if !lang.IsAggregateFunc(c.GetIdent()) {
				ns := getNames(c)
				if len(ns) == 1 && slices.Contains(names, ns[0]) {
					break
				}
				i := Issue{
					Position: c.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "column used inside a non-aggregate function",
				}
				list = append(list, i)
			}
		default:
		}
	}
	return list, nil
}

type groupbyAggrFunc struct {
	severity Severity
}

func GroupbyAggrFunc(level Severity) Rule {
	return groupbyAggrFunc{
		severity: level,
	}
}

func (_ groupbyAggrFunc) Name() string {
	return "groupby-aggr-function"
}

func (r groupbyAggrFunc) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r groupbyAggrFunc) verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}
