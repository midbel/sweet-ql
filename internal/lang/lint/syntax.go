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

func (r noStar) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noStar) verify(stmt ast.Statement) ([]Issue, error) {
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

func (r duplicateField) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r duplicateField) verify(stmt ast.Statement) ([]Issue, error) {
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

func (r setColumnsCount) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r setColumnsCount) verify(stmt ast.Statement) ([]Issue, error) {
	switch stmt := stmt.(type) {
	case ast.WithStatement:
		var list []Issue
		for _, q := range slices.Concat(stmt.Queries, slx.One(stmt.Statement)) {
			issues, err := r.verify(q)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
		return list, nil
	case ast.CteStatement:
		return r.verify(stmt.Statement)
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

func (r setColumnsCount) checkColumnsCount(left, right ast.Statement) ([]Issue, error) {
	q1, ok := left.(ast.SelectStatement)
	if !ok {
		return nil, fmt.Errorf("%s: unexpected query type", r.Name())
	}
	ok = slices.ContainsFunc(q1.Columns, func(c ast.Statement) bool {
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
	ok = slices.ContainsFunc(q2.Columns, func(c ast.Statement) bool {
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

func (r missingWhere) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r missingWhere) verify(stmt ast.Statement) ([]Issue, error) {
	return nil, nil
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

func (r enforceType) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r enforceType) verify(stmt ast.Statement) ([]Issue, error) {
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

func (r noIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noIdentQuoted) verify(stmt ast.Statement) ([]Issue, error) {
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
		check func(ast.Statement) []Issue
	)
	check = func(q ast.Statement) []Issue {
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
			list = slices.Concat(list, check(q.Statement))
		case ast.Join:
			var (
				list = check(q.Table)
				all  []ast.Statement
			)
			if gs, ok := q.Where.(interface{ GetStatement() []ast.Statement }); ok {
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
		var all []ast.Statement
		if gs, ok := c.(interface{ GetStatement() []ast.Statement }); ok {
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

func (r missingIdentQuoted) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r missingIdentQuoted) verify(stmt ast.Statement) ([]Issue, error) {
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
		check func(ast.Statement) []Issue
	)
	check = func(q ast.Statement) []Issue {
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
			list = slices.Concat(list, check(q.Statement))
		case ast.Join:
			var (
				list = check(q.Table)
				all  []ast.Statement
			)
			if gs, ok := q.Where.(interface{ GetStatement() []ast.Statement }); ok {
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
		var all []ast.Statement
		if gs, ok := c.(interface{ GetStatement() []ast.Statement }); ok {
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
