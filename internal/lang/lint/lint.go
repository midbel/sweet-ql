package lint

import (
	"slices"

	"github.com/midbel/sweet/internal/ast"
	"github.com/midbel/sweet/internal/token"
)

type Level int8

const (
	None Level = 1 << iota
	Warning
	Error
)

type LintError struct {
	token.Position
	Severity Level
	Rule     string
	Reason   string
	Query    string
}

type Rule interface {
	Verify(ast.Statement) []error
	Name() string
}

type Linter struct {
	rules []Rule
}

func Default() *Linter {
	rules := []Rule{}
	return Lint(rules)
}

func Lint(rules []Rule) *Linter {
	i := Linter{
		rules: rules,
	}
	return &i
}

func (i *Linter) Lint(stmt ast.Statement) []error {
	var list []Error
	for _, r := range i.Rules {
		errs := r.Verify(stmt)
		list = slices.Concat(list, err)
	}
	return list
}

type noStar struct {
	severity Level
}

func NoStar(level Severity) Rule {
	return noStar{
		severity: level,
	}
}

func (r noStar) Name() string {
	return "no-star"
}

func (r noStar) Verify(stmt ast.Statement) []error {
	switch q := stmt.(type) {
	case ast.SelectStatement:
	case ast.WithStatement:
	default:
	}
	return nil
}

type noCte struct {
	severity Level
}

func NoCte(level Severity) Rule {
	return noCte{
		severity: level,
	}
}

func (r noCte) Verify(stmt ast.Statement) []error {
	switch q := stmt.(type) {
	case ast.SelectStatement:
	case ast.WithStatement:
	default:
	}
	return nil
}

func (_ noCte) Name() string {
	return "no-cte"
}

type cteColumns struct {
	severity Severity
}

func CteColumns(level Severity) Rule {
	return cteColumns{
		severity: level,
	}
}

func (r cteColumns) Verify(stmt ast.Statement) []error {
	return nil
}

func (_ cteColumns) Name() string {
	return "cte-columns"
}

type noSubquery struct {
	severity Level
}

func NoSubquery(level Severity) Rule {
	return noSubquery{
		severity: level,
	}
}

func (r noSubquery) Verify(stmt ast.Statement) []error {
	switch q := stmt.(type) {
	case ast.SelectStatement:
	case ast.WithStatement:
	default:
	}
	return nil
}

func (_ noSubquery) Name() string {
	return "no-subquery"
}

type groupbyColumns struct {
	severity Severity
}

func GroupbyColumns(level Severity) Rule {
	return groupbyColumns{
		severiry: level,
	}
}

func (r groupbyColumns) Verify(stmt ast.Statement) []error {
	switch q := stmt.(type) {
	case ast.SelectStatement:
	case ast.WithStatement:
	default:
	}
	return nil
}

func (_ groupbyColumns) Name() string {
	return "groupby-columns"
}

type missingAlias struct {
	severity Severity
}

func MissingAlias(level Severity) Rule {
	return missingAlias{
		severity: level,
	}
}

func (r missingAlias) Verify(stmt ast.Statement) []error {
	return nil
}

func (_ missingAlias) Name() string {
	return "missing-alias"
}

type noAlias struct {
	severity Severity
}

func NoAlias(level Severity) Rule {
	return noAlias{
		severity: level,
	}
}

func (r noAlias) Verify(stmt ast.Statement) []error {
	return nil
}

func (_ noAlias) Name() string {
	return "no-alias"
}

type invalidAlias struct {
	severity Severity
}

func InvalidAlias(level Severity) Rule {
	return invalidAlias{
		severity: level,
	}
}

func (r invalidAlias) Verify(stmt ast.Statement) []error {
	return nil
}

func (_ invalidAlias) Name() string {
	return "invalid-alias"
}

type undefinedAlias struct {
	severity Severity
}

func UndefinedAlias(level Severity) Rule {
	return undefinedAlias{
		severity: level,
	}
}

func (r undefinedAlias) Verify(stmt ast.Statement) []error {
	return nil
}

func (_ undefinedAlias) Name() string {
	return "undefined-alias"
}
