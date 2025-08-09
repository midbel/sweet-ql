package lint

import (
	"errors"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
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

func (r *subqueryNames) VisitSelect(stmt ast.SelectStatement) error {
	for _, t := range stmt.Tables {
		if err := r.visitNode(t, stmt); err != nil {
			return err
		}
	}
	return errVisit
}

func (r *subqueryNames) visitNode(node ast.Node, stmt ast.SelectStatement) error {
	x := &subqueryJoinNames{
		Visitor: ast.Noop(),
	}
	if err := node.Accept(Walk(x)); err != nil {
		return err
	}
	if len(x.names) == 0 {
		return nil
	}
	q := &queryIdentUsage{
		Visitor: ast.Noop(),
		names:   x.names,
		alias:   x.alias,
	}
	if err := stmt.Accept(Walk(q)); err != nil {
		return err
	}
	for i := range q.issues {
		q.issues[i].Severity = r.severity
		q.issues[i].Rule = r.Name()
	}
	r.issues = slices.Concat(r.issues, q.issues)
	return nil
}

type queryIdentUsage struct {
	ast.Visitor
	issues []Issue
	names  []ast.Name
	alias  ast.Identifier
}

func (s *queryIdentUsage) VisitSelect(stmt ast.SelectStatement) error {
	for _, c := range stmt.Columns {
		if err := c.Accept(s); err != nil {
			return err
		}
	}
	if stmt.Where != nil {
		if err := stmt.Where.Accept(s); err != nil {
			return err
		}
	}
	for _, g := range stmt.Groups {
		if err := g.Accept(s); err != nil {
			return err
		}
	}
	if stmt.Having != nil {
		if err := stmt.Having.Accept(s); err != nil {
			return err
		}
	}
	return nil
}

func (s *queryIdentUsage) VisitName(name ast.Name) error {
	if len(name.Parts) <= 1 {
		return nil
	}
	if name.Parts[len(name.Parts)-2] != s.alias {
		return nil
	}
	ok := slices.ContainsFunc(s.names, func(n ast.Name) bool {
		return n.Parts[len(n.Parts)-1] == name.Parts[len(name.Parts)-1]
	})
	if !ok {
		i := Issue{
			Position: name.Pos(),
			Reason:   "identifier is not defined in subquery",
		}
		s.issues = append(s.issues, i)
	}
	return nil
}

type subqueryJoinNames struct {
	ast.Visitor

	names []ast.Name
	alias ast.Identifier
}

func (s *subqueryJoinNames) VisitAlias(alias ast.Alias) error {
	s.alias = alias.Identifier

	entry := alias.Node
	if g, ok := entry.(ast.Group); ok {
		entry = g.Node
	} else {
		return nil
	}
	if stmt, ok := entry.(ast.SelectStatement); ok {
		for _, c := range stmt.Columns {
			switch c := c.(type) {
			case ast.Name:
				s.names = append(s.names, c)
			case ast.Alias:
				n := ast.Name{
					Position: c.Pos(),
					Parts:    slx.One(c.Identifier),
				}
				s.names = append(s.names, n)
			default:
			}
		}
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
