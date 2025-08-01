package lint

import (
	"errors"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type recommandedAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func RecommandedAlias(level Severity) Rule {
	return &recommandedAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (r *recommandedAlias) Name() string {
	return "recommanded-alias"
}

func (r *recommandedAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *recommandedAlias) VisitSelect(stmt ast.SelectStatement) error {
	for _, q := range stmt.Columns {
		var pos token.Position
		switch q := q.(type) {
		case ast.Call:
			pos = q.Position
		case ast.Group:
			pos = q.Position
		case ast.Binary:
			pos = q.Position
		default:
			continue
		}
		i := Issue{
			Position: pos,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type missingAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func MissingAlias(level Severity) Rule {
	return &missingAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ missingAlias) Name() string {
	return "missing-alias"
}

func (r *missingAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *missingAlias) VisitSelect(stmt ast.SelectStatement) error {
	for _, c := range slices.Concat(stmt.Columns, stmt.Tables) {
		if _, ok := c.(ast.Alias); !ok {
			i := Issue{
				Position: getPosition(c),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "prefer using alias to improve your query",
			}
			r.issues = append(r.issues, i)
		}
	}
	return nil
}

type noAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func NoAlias(level Severity) Rule {
	return &noAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noAlias) Name() string {
	return "no-alias"
}

func (r *noAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noAlias) VisitSelect(stmt ast.SelectStatement) error {
	for _, c := range slices.Concat(stmt.Columns, stmt.Tables) {
		if a, ok := c.(ast.Alias); ok {
			i := Issue{
				Position: a.Position,
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "aliases are not recommended unless needed",
			}
			r.issues = append(r.issues, i)
		}
	}
	return nil
}

type invalidAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue

	aliases [][]ast.Identifier
	depth   int
}

func InvalidAlias(level Severity) Rule {
	return &invalidAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *invalidAlias) Name() string {
	return "invalid-alias"
}

func (r *invalidAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	r.depth++
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *invalidAlias) VisitSelect(stmt ast.SelectStatement) error {
	var aliases []ast.Identifier
	for _, c := range stmt.Columns {
		a, ok := c.(ast.Alias)
		if ok {
			aliases = append(aliases, a.Identifier)
		}
	}
	r.push(aliases)
	defer r.pop()
	return r.visit(stmt)
}

func (r *invalidAlias) VisitName(name ast.Name) error {
	if r.exists(name) {
		i := Issue{
			Position: name.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "alias is not expected in group by/where/having clause of query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *invalidAlias) visit(stmt ast.SelectStatement) error {
	var (
		where  = slx.One(stmt.Where)
		having = slx.One(stmt.Having)
		parts  = slices.Concat(stmt.Columns, stmt.Groups, where, having)
		sub    = Walk(r)
	)
	for _, q := range parts {
		if q == nil {
			continue
		}
		if err := q.Accept(sub); err != nil {
			return err
		}
	}
	if r.depth > 1 {
		return nil
	}
	return errStop
}

func (r *invalidAlias) push(list []ast.Identifier) {
	r.aliases = append(r.aliases, list)
	r.depth++
}

func (r *invalidAlias) pop() {
	n := len(r.aliases)
	if n > 0 {
		r.aliases = r.aliases[:n-1]
		r.depth--
	}
}

func (r *invalidAlias) exists(name ast.Name) bool {
	n := len(r.aliases)
	if n == 0 || len(name.Parts) != 1 {
		return false
	}
	for i := n - 1; i >= 0; i-- {
		ok := slices.ContainsFunc(r.aliases[i], func(i ast.Identifier) bool {
			return i.Name == name.Parts[0].Name
		})
		if ok {
			return ok
		}
	}
	return false
}

type undefinedAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
	aliases  [][]ast.Identifier

	depth int
}

func UndefinedAlias(level Severity) Rule {
	return &undefinedAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *undefinedAlias) Name() string {
	return "undefined-alias"
}

func (r *undefinedAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	r.depth++
	err := stmt.Accept(Walk(r))
	if errors.Is(err, errStop) {
		err = nil
	}
	return r.issues, err
}

func (r *undefinedAlias) VisitSelect(stmt ast.SelectStatement) error {
	var aliases []ast.Identifier
	for _, t := range stmt.Tables {
		if a, ok := t.(ast.Alias); ok {
			aliases = append(aliases, a.Identifier)
		}
	}
	r.push(aliases)
	defer r.pop()
	return r.visit(stmt)
}

func (r *undefinedAlias) VisitName(name ast.Name) error {
	if !r.exists(name) {
		i := Issue{
			Position: name.Position,
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "alias is not defined in from clause of query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *undefinedAlias) visit(stmt ast.SelectStatement) error {
	var (
		where  = slx.One(stmt.Where)
		having = slx.One(stmt.Having)
		parts  = slices.Concat(stmt.Columns, stmt.Groups, where, having)
		sub    = Walk(r)
	)
	for _, q := range parts {
		if q == nil {
			continue
		}
		if err := q.Accept(sub); err != nil {
			return err
		}
	}
	if r.depth > 1 {
		return nil
	}
	return errStop
}

func (r *undefinedAlias) push(list []ast.Identifier) {
	r.depth++
	r.aliases = append(r.aliases, list)
}

func (r *undefinedAlias) pop() {
	n := len(r.aliases)
	if n > 0 {
		r.depth--
		r.aliases = r.aliases[:n-1]
	}
}

func (r *undefinedAlias) exists(name ast.Name) bool {
	n := len(r.aliases)
	if n == 0 || len(name.Parts) <= 1 {
		return true
	}
	for i := n - 1; i >= 0; i-- {
		ok := slices.ContainsFunc(r.aliases[i], func(i ast.Identifier) bool {
			return i.Name == name.Parts[0].Name
		})
		if ok {
			return ok
		}
	}
	return false
}
