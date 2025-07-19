package lint

import (
	"fmt"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type noCte struct {
	severity Severity
}

func NoCte(level Severity) Rule {
	return noCte{
		severity: level,
	}
}

func (_ noCte) Name() string {
	return "no-cte"
}

func (r noCte) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noCte) verify(stmt ast.Statement) ([]Issue, error) {
	if w, ok := stmt.(ast.WithStatement); ok {
		i := Issue{
			Position: w.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use subquery instead of cte",
		}
		return slx.One(i), nil
	}
	return nil, nil
}

type cteDuplicate struct {
	severity Severity
}

func CteDuplicate(level Severity) Rule {
	return cteDuplicate{
		severity: level,
	}
}

func (_ cteDuplicate) Name() string {
	return "cte-duplicate"
}

func (r cteDuplicate) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r cteDuplicate) verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var (
		names = make(map[string]struct{})
		list  []Issue
	)
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if _, ok := names[c.Ident]; ok {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "duplicate cte name",
			}
			list = append(list, i)
		}
		names[c.Ident] = struct{}{}
	}
	return list, nil
}

type cteUnused struct {
	severity Severity
}

func CteUnused(level Severity) Rule {
	return cteUnused{
		severity: level,
	}
}

func (_ cteUnused) Name() string {
	return "cte-unused"
}

func (r cteUnused) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r cteUnused) verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var (
		positions = make(map[string]token.Position)
		names     = make(map[string]int)
	)
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		names[c.Ident] = 0
		positions[c.Ident] = c.Position
	}
	var (
		all  = slices.Concat(q.Queries, slx.One(q.Statement))
		list []Issue
	)
	for _, q := range all {
		c, ok := q.(ast.CteStatement)
		if !ok {
			continue
		}
		used := getTables(c.Statement)
		for _, n := range used {
			if _, ok := names[n]; !ok {
				continue
			}
			names[n]++
		}
	}
	for n, c := range names {
		if c > 0 {
			continue
		}
		i := Issue{
			Position: positions[n],
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "cte is defined but not used",
		}
		list = append(list, i)
	}
	return list, nil
}

type cteColumns struct {
	severity Severity
}

func CteColumns(level Severity) Rule {
	return cteColumns{
		severity: level,
	}
}

func (_ cteColumns) Name() string {
	return "cte-columns"
}

func (r cteColumns) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r cteColumns) verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var list []Issue
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if len(c.Columns) == 0 {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "columns definition missing for cte",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

type cteColumnsCount struct {
	severity Severity
}

func CteColumnsCount(level Severity) Rule {
	return cteColumnsCount{
		severity: level,
	}
}

func (_ cteColumnsCount) Name() string {
	return "cte-columns-count"
}

func (r cteColumnsCount) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r cteColumnsCount) verify(stmt ast.Statement) ([]Issue, error) {
	q, ok := stmt.(ast.WithStatement)
	if !ok {
		return nil, nil
	}
	var list []Issue
	for _, q := range q.Queries {
		c, ok := q.(ast.CteStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		e, ok := c.Statement.(ast.SelectStatement)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected query type", r.Name())
		}
		if len(c.Columns) > 0 && len(e.Columns) != len(c.Columns) {
			i := Issue{
				Position: c.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "invalid number of columns declared in cte",
			}
			list = append(list, i)
		}
	}
	return list, nil
}
