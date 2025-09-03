package lint

import (
	"errors"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
)

type selfAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

// Creates a rule that verifies whether the alias assigned to a field
// is identical to the field's original name.
func SelfAlias(level Severity) Rule {
	return &selfAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *selfAlias) Name() string {
	return "self-alias"
}

func (r *selfAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *selfAlias) VisitAlias(alias *ast.Alias) error {
	n, ok := alias.Node.(*ast.Name)
	if !ok {
		return nil
	}
	x := len(n.Parts) - 1
	if n.Parts[x] == alias.Identifier {
		i := Issue{
			Position: alias.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "do not use an alias identical to the name being aliased",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

type ambiguousAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

// Creates a Rule that checks if alias given to a function call is not
// the same as the function identifier
func AmbiguousAlias(level Severity) Rule {
	return &ambiguousAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *ambiguousAlias) Name() string {
	return "ambiguous-alias"
}

func (r *ambiguousAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *ambiguousAlias) VisitAlias(alias *ast.Alias) error {
	c, ok := alias.Node.(*ast.Call)
	if !ok {
		return nil
	}
	if c.GetIdent() == alias.Identifier.Name {
		i := Issue{
			Position: c.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "use a different identifier of the function as alias to avoid ambiguouity",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

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
	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *recommandedAlias) VisitSelect(stmt *ast.SelectStatement) error {
	for _, q := range stmt.Columns {
		switch q.(type) {
		case *ast.Call, *ast.Group, *ast.Binary, *ast.Unary:
			i := Issue{
				Position: q.Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "alias is recommanded for function call, subquery, binary and/or unary expression",
			}
			r.issues = append(r.issues, i)
		default:
		}
	}
	return nil
}

type missingAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

// Creates a Rule that ensures all fields and tables are assigned an alias.
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
	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *missingAlias) VisitSelect(stmt *ast.SelectStatement) error {
	for _, c := range slices.Concat(stmt.Columns, stmt.Tables) {
		if _, ok := c.(*ast.Alias); !ok {
			i := Issue{
				Position: c.Pos(),
				Severity: r.severity,
				Rule:     r.Name(),
				Reason:   "prefer using alias to improve readability of your query",
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

// Creates a Rule that ensures no fields or tables are assigned an alias.
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
	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noAlias) VisitSelect(stmt *ast.SelectStatement) error {
	for _, c := range slices.Concat(stmt.Columns, stmt.Tables) {
		if a, ok := c.(*ast.Alias); ok {
			i := Issue{
				Position: a.Pos(),
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
}

// Creates a Rule that checks if alias defined in the SELECT clause are
// not used in where/having/group by clauses
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
	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *invalidAlias) VisitSelect(stmt *ast.SelectStatement) error {
	var aliases []ast.Identifier
	for _, c := range stmt.Columns {
		a, ok := c.(*ast.Alias)
		if ok {
			aliases = append(aliases, a.Identifier)
		}
	}
	r.push(aliases)
	defer r.pop()
	return r.visit(stmt)
}

func (r *invalidAlias) VisitName(name *ast.Name) error {
	if r.exists(name) {
		i := Issue{
			Position: name.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "alias is not expected in group by/where/having clause of query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *invalidAlias) visit(stmt *ast.SelectStatement) error {
	var (
		where  = slx.One(stmt.Where)
		having = slx.One(stmt.Having)
		parts  = slices.Concat(stmt.Columns, stmt.Tables, stmt.Groups, where, having)
		sub    = ast.Walk(r)
	)
	for _, q := range parts {
		if q == nil {
			continue
		}
		if err := q.Accept(sub); err != nil {
			return err
		}
	}
	return ast.ErrVisit
}

func (r *invalidAlias) push(list []ast.Identifier) {
	r.aliases = append(r.aliases, list)
}

func (r *invalidAlias) pop() {
	n := len(r.aliases)
	if n > 0 {
		r.aliases = r.aliases[:n-1]
	}
}

func (r *invalidAlias) exists(name *ast.Name) bool {
	n := len(r.aliases)
	if n == 0 || len(name.Parts) == 1 {
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

	aliases [][]ast.Identifier
}

// Creates a rule that checks whether the qualified fields in the
// SELECT clause use the table aliases defined in the query.
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
	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *undefinedAlias) VisitSelect(stmt *ast.SelectStatement) error {
	var aliases []ast.Identifier
	for _, t := range stmt.Tables {
		if a, ok := t.(*ast.Alias); ok {
			aliases = append(aliases, a.Identifier)
		}
	}
	r.push(aliases)
	defer r.pop()
	return r.visit(stmt)
}

func (r *undefinedAlias) VisitName(name *ast.Name) error {
	if !r.exists(name) {
		i := Issue{
			Position: name.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "alias is not defined in from clause of query",
		}
		r.issues = append(r.issues, i)
	}
	return nil
}

func (r *undefinedAlias) visit(stmt *ast.SelectStatement) error {
	var (
		where  = slx.One(stmt.Where)
		having = slx.One(stmt.Having)
		parts  = slices.Concat(stmt.Columns, stmt.Groups, where, having)
		sub    = ast.Walk(r)
	)
	for _, q := range parts {
		if q == nil {
			continue
		}
		if err := q.Accept(sub); err != nil {
			return err
		}
	}
	return ast.ErrVisit
}

func (r *undefinedAlias) push(list []ast.Identifier) {
	r.aliases = append(r.aliases, list)
}

func (r *undefinedAlias) pop() {
	n := len(r.aliases)
	if n > 0 {
		r.aliases = r.aliases[:n-1]
	}
}

func (r *undefinedAlias) exists(name *ast.Name) bool {
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

type unusedAlias struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func UnusedAlias(level Severity) Rule {
	return &unusedAlias{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *unusedAlias) Name() string {
	return "unused-alias"
}

func (r *unusedAlias) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}
