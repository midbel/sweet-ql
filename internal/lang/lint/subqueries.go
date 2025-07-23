package lint

import (
	"fmt"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
)

type subqueryColumnsCount struct {
	severity Severity
}

func SubqueryColumnsCount(level Severity) Rule {
	return subqueryColumnsCount{
		severity: level,
	}
}

func (_ subqueryColumnsCount) Name() string {
	return "subquery-columns-count"
}

func (r subqueryColumnsCount) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r subqueryColumnsCount) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkColumnsCount)
}

func (r subqueryColumnsCount) checkColumnsCount(q ast.SelectStatement) ([]Issue, error) {
	var (
		list    []Issue
		queries []ast.SelectStatement
	)
	for _, c := range q.Columns {
		queries = slices.Concat(queries, getQueries(c))
	}
	queries = slices.Concat(queries, getQueries(q.Where))
	queries = slices.Concat(queries, getQueries(q.Having))
	for _, q := range queries {
		issues, err := r.checkColumns(q)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func (r subqueryColumnsCount) checkColumns(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	switch len(q.Columns) {
	case 0:
	case 1:
		n, ok := q.Columns[0].(ast.Name)
		if ok && n.All() {
			i := Issue{
				Position: getPosition(q.Columns[0]),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "unknown columns count because of use of '*'",
			}
			list = append(list, i)
		}
	default:
		i := Issue{
			Position: getPosition(q.Columns[0]),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "invalid columns count used by subquery",
		}
		list = append(list, i)
	}
	return list, nil
}

type subqueryNames struct {
	severity Severity
}

func SubqueryNames(level Severity) Rule {
	return subqueryNames{
		severity: level,
	}
}

func (_ subqueryNames) Name() string {
	return "subquery-name"
}

func (r subqueryNames) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r subqueryNames) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkExportedNames)
}

func (r subqueryNames) checkExportedNames(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, t := range q.Tables {
		j, ok := t.(ast.Join)
		if !ok {
			continue
		}
		names, err := r.getExportedNames(j)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			continue
		}
		issues, err := r.checkNames(j.Where, names)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
		if issues, err = r.checkNames(q.Where, names); err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
		for _, c := range q.Columns {
			issues, err := r.checkNames(c, names)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
	}
	return list, nil
}

func (r subqueryNames) checkNames(stmt ast.Node, names [][]string) ([]Issue, error) {
	var list []Issue
	for _, n := range getNames2(stmt) {
		if len(n) == 0 || n[0] != names[0][0] {
			continue
		}
		ok := slices.ContainsFunc(names, func(ns []string) bool {
			return slices.Equal(ns, n)
		})
		if !ok {
			i := Issue{
				Position: getPosition(stmt),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "name not exported by subquery",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

func (r subqueryNames) getExportedNames(j ast.Join) ([][]string, error) {
	a, ok := j.Table.(ast.Alias)
	if !ok {
		return nil, nil
	}

	g, ok := a.Statement.(ast.Group)
	if !ok {
		return nil, nil
	}
	s, ok := g.Statement.(ast.SelectStatement)
	if !ok {
		return nil, fmt.Errorf("%s: unexpected query type", r.Name())
	}
	var names [][]string
	for _, c := range s.Columns {
		var ns []string
		switch n := c.(type) {
		case ast.Alias:
			ns = append(ns, a.Name, n.Name)
		case ast.Name:
			ns = append(ns, a.Name, n.Name())
		default:
		}
		if len(ns) > 0 {
			names = append(names, ns)
		}
	}
	return names, nil
}

type noSubquery struct {
	severity Severity
}

func NoSubquery(level Severity) Rule {
	return noSubquery{
		severity: level,
	}
}

func (_ noSubquery) Name() string {
	return "no-subquery"
}

func (r noSubquery) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noSubquery) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkSubquery)
}

func (r noSubquery) checkSubquery(stmt ast.SelectStatement) ([]Issue, error) {
	queries := getQueries(stmt)
	if len(queries) == 1 {
		return nil, nil
	}
	var list []Issue
	for _, q := range queries[1:] {
		i := Issue{
			Position: q.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "avoid using subquery",
		}
		list = append(list, i)
	}
	return list, nil
}
