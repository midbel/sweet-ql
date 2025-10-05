package lint

import (
	"fmt"

	"github.com/midbel/sweet/internal/lang/ast"
)

type noReturning struct {
	ast.Visitor
	*rule
}

func NoReturning(level Severity) Rule {
	a := &noReturning{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, upsertReturn, level)
	return a
}

func (r *noReturning) VisitDelete(stmt *ast.DeleteStatement) error {
	return r.check(stmt.Returning)
}

func (r *noReturning) VisitUpdate(stmt *ast.UpdateStatement) error {
	return r.check(stmt.Returning)
}

func (r *noReturning) VisitInsert(stmt *ast.InsertStatement) error {
	return r.check(stmt.Returning)
}

func (r *noReturning) check(stmt ast.Node) error {
	if stmt != nil {
		return r.Report(stmt, "using returning is not ansi compliant")
	}
	return nil
}

// check that no "default" is provided in insert statement
type noDefaultValue struct {
	ast.Visitor
	*rule
}

func NoDefaultValue(level Severity) Rule {
	a := &noDefaultValue{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, upsertDefault, level)
	return a
}

func (r *noDefaultValue) VisitInsert(stmt *ast.InsertStatement) error {
	q, ok := stmt.Values.(*ast.ValuesStatement)
	if !ok {
		return nil
	}
	for _, v := range q.List {
		if err := r.visitList(v); err != nil {
			return err
		}
	}
	return nil
}

func (r *noDefaultValue) visitList(list ast.Node) error {
	i, ok := list.(*ast.List)
	if !ok {
		return nil
	}
	for _, i := range i.Values {
		n, ok := i.(*ast.Value)
		if ok && n.Default() {
			if err := r.Report(i, "avoid using default in values"); err != nil {
				return err
			}
		}
	}
	return nil
}

// check that only one unconditional match in a merge statement is present
type unconditionalMatch struct {
	ast.Visitor
	*rule
}

func UnconditionalMatch(level Severity) Rule {
	a := &unconditionalMatch{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, upsertMergeUnconditional, level)
	return a
}

func (r *unconditionalMatch) VisitMerge(stmt *ast.MergeStatement) error {
	var (
		matched       int
		lastMatched   ast.Node
		noMatched     int
		lastNoMatched ast.Node
	)
	for _, a := range stmt.Actions {
		m, ok := a.(*ast.MatchStatement)
		if !ok {
			return fmt.Errorf("%s: unexpected query type", r.Name())
		}
		switch m.Node.(type) {
		case *ast.InsertStatement:
			if m.Condition == nil {
				noMatched++
				lastNoMatched = m
			}
		case *ast.UpdateStatement, *ast.DeleteStatement:
			if m.Condition == nil {
				matched++
				lastMatched = m
			}
		default:
			return fmt.Errorf("%s: unexpected query type", r.Name())
		}
	}
	if matched > 1 {
		r.Report(lastMatched, "only one unconditional \"match\" allowed in merge statement")
	}
	if noMatched > 1 {
		r.Report(lastNoMatched, "only one unconditional \"not match\" allowed in merge statement")
	}
	return nil
}
