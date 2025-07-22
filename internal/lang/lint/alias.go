package lint

import (
	"fmt"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
)

type setAlias struct {
	severity Severity
}

func SetAlias(level Severity) Rule {
	return setAlias{
		severity: level,
	}
}

func (r setAlias) Name() string {
	return "set-alias"
}

func (r setAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r setAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkAliasForCalculatedFields)
}

func (r setAlias) checkAliasForCalculatedFields(q ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	for _, c := range q.Columns {
		switch c.(type) {
		case ast.Call, ast.Binary:
			i := Issue{
				Position: getPosition(c),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "use alias for calculated field",
			}
			list = append(list, i)
		default:
		}
	}
	return list, nil
}

type missingAlias struct {
	severity Severity
	options  RuleOptions
}

func MissingAlias(level Severity) Rule {
	return missingAlias{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func MissingAliasOnFields(level Severity) Rule {
	return missingAlias{
		severity: level,
		options:  CheckFields,
	}
}

func MissingAliasOnTables(level Severity) Rule {
	return missingAlias{
		severity: level,
		options:  CheckTables,
	}
}

func (_ missingAlias) Name() string {
	return "missing-alias"
}

func (r missingAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r missingAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkMissingAlias)
}

func (r missingAlias) checkMissingAlias(stmt ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	if r.options.withCheckFields() {
		for _, c := range stmt.Columns {
			if _, ok := c.(ast.Alias); !ok {
				i := Issue{
					Position: getPosition(c),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "field used without alias",
				}
				list = append(list, i)
			}
		}
	}
	if r.options.withCheckTables() {
		for _, t := range stmt.Tables {
			if _, ok := t.(ast.Alias); !ok {
				i := Issue{
					Position: getPosition(t),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "table used without alias",
				}
				list = append(list, i)
			}
		}
	}
	return list, nil
}

type noAlias struct {
	severity Severity
	options  RuleOptions
}

func NoAlias(level Severity) Rule {
	return noAlias{
		severity: level,
		options:  CheckFields | CheckTables,
	}
}

func NoAliasOnFields(level Severity) Rule {
	return noAlias{
		severity: level,
		options:  CheckFields,
	}
}

func NoAliasOnTables(level Severity) Rule {
	return noAlias{
		severity: level,
		options:  CheckTables,
	}
}

func (_ noAlias) Name() string {
	return "no-alias"
}

func (r noAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r noAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkNoAlias)
}

func (r noAlias) checkNoAlias(stmt ast.SelectStatement) ([]Issue, error) {
	var list []Issue
	if r.options.withCheckFields() {
		for _, c := range stmt.Columns {
			if a, ok := c.(ast.Alias); ok {
				i := Issue{
					Position: a.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "field used with an alias",
				}
				list = append(list, i)
			}
		}
	}
	if r.options.withCheckTables() {
		for _, t := range stmt.Tables {
			if a, ok := t.(ast.Alias); ok {
				i := Issue{
					Position: a.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "table used with an alias",
				}
				list = append(list, i)
			}
		}
	}
	return list, nil
}

type invalidAlias struct {
	severity Severity
}

func InvalidAlias(level Severity) Rule {
	return invalidAlias{
		severity: level,
	}
}

func (_ invalidAlias) Name() string {
	return "invalid-alias"
}

func (r invalidAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r invalidAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkInvalidAlias)
}

func (r invalidAlias) checkInvalidAlias(q ast.SelectStatement) ([]Issue, error) {
	if q.Where == nil {
		return nil, nil
	}
	var aliases []string
	for _, c := range q.Columns {
		a, ok := c.(ast.Alias)
		if ok {
			aliases = append(aliases, a.Name)
		}
	}
	if len(aliases) == 0 {
		return nil, nil
	}
	b, ok := q.Where.(ast.Binary)
	if !ok {
		return nil, fmt.Errorf("%s: unexpected query type", r.Name())
	}
	var list []Issue
	for _, n := range getNames(b) {
		if ok := slices.Contains(aliases, n); ok {
			i := Issue{
				Position: b.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "alias used in \"where\" clause of query",
			}
			list = append(list, i)
		}
	}
	for _, g := range q.Groups {
		for _, n := range getNames(g) {
			if ok := slices.Contains(aliases, n); ok {
				i := Issue{
					Position: getPosition(g),
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "alias used in \"group by\" clause of query",
				}
				list = append(list, i)
			}
		}
	}
	return list, nil
}

type undefinedAlias struct {
	severity Severity
}

func UndefinedAlias(level Severity) Rule {
	return undefinedAlias{
		severity: level,
	}
}

func (_ undefinedAlias) Name() string {
	return "undefined-alias"
}

func (r undefinedAlias) Verify(stmt ast.Statement) ([]Issue, error) {
	return r.verify(stmt)
}

func (r undefinedAlias) verify(stmt ast.Statement) ([]Issue, error) {
	return verify(stmt, r.checkUndefinedAlias)
}

func (r undefinedAlias) checkUndefinedAlias(stmt ast.SelectStatement) ([]Issue, error) {
	var (
		aliases []string
		list    []Issue
	)
	for _, t := range stmt.Tables {
		if a, ok := t.(ast.Alias); ok {
			aliases = append(aliases, a.Name)
		}
	}

	for _, c := range stmt.Columns {
		ns := getNames(c)
		if len(ns) <= 1 {
			continue
		}
		ok := slices.Contains(aliases, ns[len(ns)-2])
		if !ok || len(aliases) == 0 {
			i := Issue{
				Position: getPosition(c),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "undefined name",
			}
			list = append(list, i)
		}
	}
	return list, nil
}
