package lint

import (
	"slices"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
)

type groupbyColumns struct {
	ast.Visitor
	*rule
}

func GroupbyColumns(level Severity) Rule {
	a := &groupbyColumns{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "groupby-columns", level)
	return a
}

func (r *groupbyColumns) VisitSelect(stmt *ast.SelectStatement) error {
	if len(stmt.Groups) == 0 {
		return nil
	}
	for _, c := range stmt.Columns {
		if a, ok := c.(*ast.Alias); ok {
			c = a.Node
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
			err := r.Report(c, "column must be used in group by clause if not used in aggregate function")
			if err != nil {
				return err
			}
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
	*rule
}

func GroupbyDistinct(level Severity) Rule {
	a := &groupbyDistinct{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "groupby-distinct", level)
	return a
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
		err := r.Report(stmt, "use distinct in select clause if no aggregate functions are used")
		if err != nil {
			return err
		}
	}
	return nil
}

type noPositionGroupby struct {
	ast.Visitor
	*rule
}

func NoPositionGroupby(level Severity) Rule {
	a := &noPositionGroupby{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "groupby-no-position", level)
	return a
}

func (r *noPositionGroupby) VisitSelect(stmt *ast.SelectStatement) error {
	for _, g := range stmt.Groups {
		if v, ok := g.(*ast.Value); ok && v.Number() {
			err := r.Report(g, "prefer using field names in group by instead of position")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type noLiteralGroupby struct {
	ast.Visitor
	*rule
}

func NoLiteralGroupby(level Severity) Rule {
	a := &noLiteralGroupby{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "groupby-no-literal", level)
	return a
}

func (r *noLiteralGroupby) VisitSelect(stmt *ast.SelectStatement) error {
	for _, g := range stmt.Groups {
		if v, ok := g.(*ast.Value); ok && !v.Number() {
			err := r.Report(g, "use explicit columns name or expression from select clause in group by")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// check that when aggregate functions are used in select clause, group by list is not empty
type groupbyAggrFunc struct {
	ast.Visitor
	*rule
}

func GroupbyAggrFunc(level Severity) Rule {
	a := &groupbyAggrFunc{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "groupby-aggr-function", level)
	return a
}

func (r *groupbyAggrFunc) VisitSelect(stmt *ast.SelectStatement) error {
	if len(stmt.Columns) > 0 {
		return nil
	}
	for _, c := range stmt.Columns {
		if a, ok := c.(*ast.Alias); ok {
			c = a.Node
		}
		if c, ok := c.(*ast.Call); ok && lang.IsAggregateFunc(c.GetIdent()) {
			if len(stmt.Groups) == 0 {
				err := r.Report(c, "aggregate function used but group by clause is empty")
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type callFuncVisitor struct {
	ast.Visitor
	check func(*ast.Call) error
}

func visitCallFunc(check func(*ast.Call) error) ast.Visitor {
	return &callFuncVisitor{
		Visitor: ast.Noop(),
		check:   check,
	}
}

func (v *callFuncVisitor) VisitCallFunc(call *ast.Call) error {
	return v.check(call)
}

// check that only aggregate function are used in having clause
type havingAggrFunc struct {
	ast.Visitor
	*rule
}

func HavingAggrFunc(level Severity) Rule {
	a := &havingAggrFunc{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "groupby-having-aggr-function", level)
	return a
}

func (r *havingAggrFunc) VisitSelect(stmt *ast.SelectStatement) error {
	if stmt.Having == nil {
		return nil
	}
	if len(stmt.Groups) == 0 && stmt.Having != nil {
		return r.Report(stmt, "use of having clause without group by")
	}
	sub := ast.Walk(visitCallFunc(r.visitCall))
	return stmt.Having.Accept(sub)
}

func (r *havingAggrFunc) visitCall(call *ast.Call) error {
	if !lang.IsAggregateFunc(call.GetIdent()) {
		return r.Report(call, "use only aggregate function in having clause")
	}
	return nil
}
