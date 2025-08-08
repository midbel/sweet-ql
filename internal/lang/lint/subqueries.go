package lint

import (
	"errors"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
)

type subqueryColumnsCount struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func SubqueryColumnsCount(level Severity) Rule {
	return &subqueryColumnsCount{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *subqueryColumnsCount) Name() string {
	return "subquery-columns-count"
}

func (r *subqueryColumnsCount) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *subqueryColumnsCount) VisitJoin(join ast.Join) error {
	var (
		other = subqueryColumnsCount{
			severity: r.severity,
			Visitor:  ast.Noop(),
		}
		sub   = Walk(&other)
		entry = join.Table
	)
	if a, ok := entry.(ast.Alias); ok {
		entry = a.Node
	}
	if g, ok := entry.(ast.Group); ok {
		entry = g.Node
	}
	if err := entry.Accept(sub); err != nil {
		return err
	}
	if err := join.Where.Accept(sub); err != nil {
		return err
	}
	r.issues = slices.Concat(r.issues, other.issues)
	return errVisit
}

func (r *subqueryColumnsCount) VisitGroup(group ast.Group) error {
	if stmt, ok := group.Node.(ast.SelectStatement); ok {
		if len(stmt.Columns) != 1 {
			i := Issue{
				Position: stmt.Columns[0].Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "expected one column returned by subquery",
			}
			r.issues = append(r.issues, i)
		} else {
			n, ok := stmt.Columns[0].(ast.Name)
			if ok && n.All() {
				i := Issue{
					Position: stmt.Columns[0].Pos(),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "unable to determine number of columns returned by subquery",
				}
				r.issues = append(r.issues, i)
			}
		}

	}
	return nil
}

type subqueryNames struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func SubqueryNames(level Severity) Rule {
	return &subqueryNames{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *subqueryNames) Name() string {
	return "subquery-name"
}

func (r *subqueryNames) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *subqueryNames) VisitJoin(join ast.Join) error {
	return nil
}

func (r *subqueryNames) VisitGroup(group ast.Group) error {
	if _, ok := group.Node.(ast.SelectStatement); ok {

	}
	return nil
}

type noSubquery struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func NoSubquery(level Severity) Rule {
	return &noSubquery{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noSubquery) Name() string {
	return "no-subquery"
}

func (r *noSubquery) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noSubquery) VisitGroup(group ast.Group) error {
	if _, ok := group.Node.(ast.SelectStatement); ok {
		i := Issue{
			Position: group.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "consider rewriting subqueries with join and/or cte",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}
