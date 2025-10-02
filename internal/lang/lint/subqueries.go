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

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *subqueryColumnsCount) VisitJoin(join *ast.Join) error {
	var (
		other = subqueryColumnsCount{
			severity: r.severity,
			Visitor:  ast.Noop(),
		}
		sub   = ast.Walk(&other)
		entry = join.Table
	)
	if a, ok := entry.(*ast.Alias); ok {
		entry = a.Node
	}
	if g, ok := entry.(*ast.Group); ok {
		entry = g.Node
	}
	if err := entry.Accept(sub); err != nil {
		return err
	}
	if err := join.Where.Accept(sub); err != nil {
		return err
	}
	r.issues = slices.Concat(r.issues, other.issues)
	return ast.ErrVisit
}

func (r *subqueryColumnsCount) VisitGroup(group *ast.Group) error {
	if stmt, ok := group.Node.(*ast.SelectStatement); ok {
		if len(stmt.Columns) != 1 {
			i := Issue{
				Position: stmt.Columns[0].Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "expected one column returned by subquery",
			}
			r.issues = append(r.issues, i)
		} else {
			n, ok := stmt.Columns[0].(*ast.Name)
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

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *subqueryNames) VisitSelect(stmt *ast.SelectStatement) error {
	for _, t := range stmt.Tables {
		if err := r.visitNode(t, stmt); err != nil {
			return err
		}
	}
	return ast.ErrVisit
}

func (r *subqueryNames) visitNode(node ast.Node, stmt *ast.SelectStatement) error {
	x := &subqueryJoinNames{
		Visitor: ast.Noop(),
	}
	if err := node.Accept(ast.Walk(x)); err != nil {
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
	if err := stmt.Accept(ast.Walk(q)); err != nil {
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
	names  []*ast.Name
	alias  ast.Identifier
}

func (s *queryIdentUsage) VisitName(name *ast.Name) error {
	if len(name.Parts) <= 1 {
		return nil
	}
	if name.Parts[len(name.Parts)-2] != s.alias {
		return nil
	}
	ok := slices.ContainsFunc(s.names, func(n *ast.Name) bool {
		if n.Pos() == name.Pos() {
			return true
		}
		return n.Parts[len(n.Parts)-1] == name.Parts[len(name.Parts)-1]
	})
	if !ok {
		i := Issue{
			Position: name.Pos(),
			Reason:   "the identifier is not declared or returned by subquery",
		}
		s.issues = append(s.issues, i)
	}
	return nil
}

type subqueryJoinNames struct {
	ast.Visitor

	names []*ast.Name
	alias ast.Identifier
}

func (s *subqueryJoinNames) VisitAlias(alias *ast.Alias) error {
	s.alias = alias.Identifier

	entry := alias.Node
	if g, ok := entry.(*ast.Group); ok {
		entry = g.Node
	} else {
		return nil
	}
	if stmt, ok := entry.(*ast.SelectStatement); ok {
		for _, c := range stmt.Columns {
			switch c := c.(type) {
			case *ast.Name:
				s.names = append(s.names, c)
			case *ast.Alias:
				n := &ast.Name{
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
	*rule
}

func NoSubquery(level Severity) Rule {
	a := &noSubquery{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "no-subquery", level)
	return a
}

func (r *noSubquery) VisitGroup(group *ast.Group) error {
	if _, ok := group.Node.(*ast.SelectStatement); ok {
		return r.Report(group, "consider rewriting subqueries with join and/or cte")
	}
	return nil
}
