package lint

import (
	"fmt"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
)

type noStar struct {
	severity Severity
}

func NoStar(level Severity) Rule {
	return noStar{
		severity: level,
	}
}

func (_ noStar) Name() string {
	return "no-star"
}

func (r noStar) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noStar) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkStar)
}

func (r noStar) checkStar(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, q := range getQueries(q) {
		issues, err := r.checkColumns(q)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func (r noStar) checkColumns(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, c := range q.Columns {
		if n, ok := c.(ast.Name); ok && n.All() {
			i := Issue{
				Position: n.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "use explicit field names",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

type duplicateField struct {
	severity Severity
}

func DuplicateField(level Severity) Rule {
	return duplicateField{
		severity: level,
	}
}

func (_ duplicateField) Name() string {
	return "duplicate-field"
}

func (r duplicateField) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r duplicateField) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkDuplicateFields)
}

func (r duplicateField) checkDuplicateFields(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, q := range getQueries(q) {
		issues, err := r.checkColumns(q)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func (r duplicateField) checkColumns(q ast.SelectStatement) ([]Issue, error) {
	var (
		list  []Issue
		names [][]ast.Identifier
	)
	for _, c := range q.Columns {
		var id []ast.Identifier
		switch c := c.(type) {
		case ast.Name:
			if c.All() && len(q.Columns) > 1 {
				i := Issue{
					Position: getPosition(c),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "implicit duplicated field",
				}
				list = append(list, i)
				continue
			}
			id = c.Parts
		case ast.Alias:
			id = slx.One(c.Identifier)
		default:
			continue
		}
		ok := slices.ContainsFunc(names, func(n []ast.Identifier) bool {
			return slices.Equal(id, n)
		})
		if ok {
			i := Issue{
				Position: getPosition(c),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "duplicated field",
			}
			list = append(list, i)
			continue
		}
		names = append(names, id)
	}
	return list, nil
}

type setColumnsCount struct {
	severity Severity
}

func SetColumnsCount(level Severity) Rule {
	return setColumnsCount{
		severity: level,
	}
}

func (_ setColumnsCount) Name() string {
	return "set-columns-count"
}

func (r setColumnsCount) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r setColumnsCount) verify(stmt ast.Node) ([]Issue, error) {
	switch stmt := stmt.(type) {
	case ast.WithStatement:
		var list []Issue
		for _, q := range slices.Concat(stmt.Queries, slx.One(stmt.Node)) {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		return list, nil
	case ast.CteStatement:
		return r.verify(stmt.Node)
	case ast.SelectStatement:
		return nil, nil
	case ast.UnionStatement:
		return r.checkUnionColumnsCount(stmt)
	case ast.ExceptStatement:
		return r.checkExceptColumnsCount(stmt)
	case ast.IntersectStatement:
		return r.checkIntersectColumnsCount(stmt)
	default:
		return nil, nil
	}
}

func (r setColumnsCount) checkUnionColumnsCount(q ast.UnionStatement) ([]Issue, error) {
	return r.checkColumnsCount(q.Left, q.Right)
}

func (r setColumnsCount) checkExceptColumnsCount(q ast.ExceptStatement) ([]Issue, error) {
	return r.checkColumnsCount(q.Left, q.Right)
}

func (r setColumnsCount) checkIntersectColumnsCount(q ast.IntersectStatement) ([]Issue, error) {
	return r.checkColumnsCount(q.Left, q.Right)
}

func (r setColumnsCount) checkColumnsCount(left, right ast.Node) ([]Issue, error) {
	q1, ok := left.(ast.SelectStatement)
	if !ok {
		return nil, fmt.Errorf("%s: unexpected query type", r.Name())
	}
	ok = slices.ContainsFunc(q1.Columns, func(c ast.Node) bool {
		n, ok := c.(ast.Name)
		return ok && n.All()
	})
	if ok {
		i := Issue{
			Position: q1.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "unknown columns count because of use of '*'",
		}
		return slx.One(i), nil
	}

	q2, ok := right.(ast.SelectStatement)
	if !ok {
		return nil, fmt.Errorf("%s: unexpected query type", r.Name())
	}
	ok = slices.ContainsFunc(q2.Columns, func(c ast.Node) bool {
		n, ok := c.(ast.Name)
		return ok && n.All()
	})
	if ok {
		i := Issue{
			Position: q2.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "unknown columns count because of use of '*'",
		}
		return slx.One(i), nil
	}
	var list []Issue
	if len(q1.Columns) != len(q2.Columns) {
		i := Issue{
			Position: q1.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "columns count mismatched",
		}
		list = append(list, i)
	}
	for _, s := range slx.Make(q1, q2) {
		issues, err := r.verify(s)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

type missingWhere struct {
	severity Severity
}

func MissingWhere(level Severity) Rule {
	return missingWhere{
		severity: level,
	}
}

func (_ missingWhere) Name() string {
	return "missing-where"
}

func (r missingWhere) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r missingWhere) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkMissingWhere)
}

func (r missingWhere) checkMissingWhere(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, q := range getQueries(q) {
		if q.Where == nil {
			i := Issue{
				Position: q.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "missing where clause from query",
			}
			list = append(list, i)
		}
	}
	return list, nil
}

type enforceType struct {
	severity Severity
}

func EnforceType(level Severity) Rule {
	return missingWhere{
		severity: level,
	}
}

func (_ enforceType) Name() string {
	return "enforce-type"
}

func (r enforceType) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r enforceType) verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

type noIdentQuoted struct {
	severity Severity
	options  RuleOptions
}

func NoIdentQuoted(level Severity) Rule {
	return noIdentQuoted{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func (_ noIdentQuoted) Name() string {
	return "no-ident-quoted"
}

func (r noIdentQuoted) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noIdentQuoted) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkQuotedIdentifiers)
}

func (r noIdentQuoted) checkQuotedIdentifiers(stmt ast.SelectStatement) ([]Issue, error) {
	var (
		queries = getQueries(stmt)
		list    []Issue
	)
	for _, q := range queries {
		issues, err := r.checkQuotes(q)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func (r noIdentQuoted) checkQuotes(q ast.SelectStatement) ([]Issue, error) {
	var (
		list  []Issue
		check func(ast.Node) []Issue
	)
	check = func(q ast.Node) []Issue {
		switch q := q.(type) {
		case ast.Name:
			for _, n := range q.Parts {
				if n.Quoted {
					i := Issue{
						Position: q.Position,
						Severity: r.severity,
						Rule:     r.Name(),
						Reason:   "identifier used with double quote",
					}
					return slx.One(i)
				}
			}
		case ast.Alias:
			var list []Issue
			if q.Quoted {
				i := Issue{
					Position: q.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "alias used with double quote",
				}
				list = append(list, i)
			}
			list = slices.Concat(list, check(q.Node))
		case ast.Join:
			var (
				list = check(q.Table)
				all  []ast.Node
			)
			if gs, ok := q.Where.(interface{ GetStatement() []ast.Node }); ok {
				all = gs.GetStatement()
			} else {
				all = slx.One(q.Where)
			}
			for _, s := range all {
				list = slices.Concat(list, check(s))
			}
			return list
		default:
		}
		return nil
	}
	for _, c := range slices.Concat(q.Columns, q.Tables, q.Groups) {
		list = slices.Concat(list, check(c))
	}
	for _, c := range slx.Make(q.Where, q.Having) {
		var all []ast.Node
		if gs, ok := c.(interface{ GetStatement() []ast.Node }); ok {
			all = gs.GetStatement()
		} else {
			all = slx.One(c)
		}
		for _, s := range all {
			list = slices.Concat(list, check(s))
		}
	}
	return list, nil
}

type missingIdentQuoted struct {
	severity Severity
	options  RuleOptions
}

func MissingIdentQuoted(level Severity) Rule {
	return missingIdentQuoted{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func (_ missingIdentQuoted) Name() string {
	return "missing-ident-quoted"
}

func (r missingIdentQuoted) Verify(stmt ast.Node) ([]Issue, error) {
	return r.verify(stmt)
}

func (r missingIdentQuoted) verify(stmt ast.Node) ([]Issue, error) {
	return verify(stmt, r.checkMissingQuotes)
}

func (r missingIdentQuoted) checkMissingQuotes(stmt ast.SelectStatement) ([]Issue, error) {
	var (
		queries = getQueries(stmt)
		list    []Issue
	)
	for _, q := range queries {
		issues, err := r.checkQuotes(q)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func (r missingIdentQuoted) checkQuotes(q ast.SelectStatement) ([]Issue, error) {
	var (
		list  []Issue
		check func(ast.Node) []Issue
	)
	check = func(q ast.Node) []Issue {
		switch q := q.(type) {
		case ast.Name:
			for _, n := range q.Parts {
				if !n.Quoted {
					i := Issue{
						Position: q.Position,
						Severity: r.severity,
						Rule:     r.Name(),
						Reason:   "identifier used without double quote",
					}
					return slx.One(i)
				}
			}
		case ast.Alias:
			var list []Issue
			if !q.Quoted {
				i := Issue{
					Position: q.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "alias used without double quote",
				}
				list = append(list, i)
			}
			list = slices.Concat(list, check(q.Node))
		case ast.Join:
			var (
				list = check(q.Table)
				all  []ast.Node
			)
			if gs, ok := q.Where.(interface{ GetStatement() []ast.Node }); ok {
				all = gs.GetStatement()
			} else {
				all = slx.One(q.Where)
			}
			for _, s := range all {
				list = slices.Concat(list, check(s))
			}
			return list
		default:
		}
		return nil
	}
	for _, c := range slices.Concat(q.Columns, q.Tables, q.Groups) {
		list = slices.Concat(list, check(c))
	}
	for _, c := range slx.Make(q.Where, q.Having) {
		var all []ast.Node
		if gs, ok := c.(interface{ GetStatement() []ast.Node }); ok {
			all = gs.GetStatement()
		} else {
			all = slx.One(c)
		}
		for _, s := range all {
			list = slices.Concat(list, check(s))
		}
	}
	return list, nil
}

// avoid using literal value in join predicate
type noLiteralJoin struct {
	severity Severity
}

func NoLiteralInJoin(level Severity) Rule {
	return noLiteralJoin{
		severity: level,
	}
}

func (_ noLiteralJoin) Name() string {
	return "no-literal-join"
}

func (r noLiteralJoin) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// prefer using offset fetch syntax over limit offset
type offsetFetch struct {
	severity Severity
}

func OffsetFetch(level Severity) Rule {
	return offsetFetch{
		severity: level,
	}
}

func (_ offsetFetch) Name() string {
	return "offset-and-fetch"
}

func (r offsetFetch) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// when using order by clause, specify offset fetch clause
type orderOffsetFetch struct {
	severity Severity
}

func OrderWithOffset(level Severity) Rule {
	return orderOffsetFetch{
		severity: level,
	}
}

func (_ orderOffsetFetch) Name() string {
	return "order-with-offset"
}

func (r orderOffsetFetch) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// check that only the second select in union/except/intersect has the order by clause
type setOrderLast struct {
	severity Severity
}

func SetOrderLast(level Severity) Rule {
	return setOrderLast{
		severity: level,
	}
}

func (_ setOrderLast) Name() string {
	return "set-order-last"
}

func (r setOrderLast) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

// check that only the second select in union/except/intersect has the offset/fetch clause
type setOffsetFetchLast struct {
	severity Severity
}

func SetOffsetFetchLast(level Severity) Rule {
	return setOrderLast{
		severity: level,
	}
}

func (_ setOffsetFetchLast) Name() string {
	return "set-offset-fetch-last"
}

func (r setOffsetFetchLast) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}

type unconditionalMatch struct {
	severity Severity
}

// check that only one unconditional match in a merge statement is present
func UnconditionalMatch(level Severity) Rule {
	return unconditionalMatch{
		severity: level,
	}
}

func (_ unconditionalMatch) Name() string {
	return "merge-unconditional-match"
}

func (r unconditionalMatch) Verify(stmt ast.Node) ([]Issue, error) {
	return nil, nil
}
