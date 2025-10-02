package lint

import (
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type selfAlias struct {
	ast.Visitor
	*rule
}

// Creates a rule that verifies whether the alias assigned to a field
// is identical to the field's original name.
func SelfAlias(level Severity) Rule {
	a := &selfAlias{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, aliasSelf, level)
	return a
}

func (r *selfAlias) VisitAlias(alias *ast.Alias) error {
	var err error
	switch n := alias.Node.(type) {
	case *ast.Name:
		if n.Name() == alias.Name {
			err = r.Report(n, "do not use an alias identical to the name being aliased")
		}
	case *ast.Call:
		if n.GetIdent() == alias.Name {
			err = r.Report(n, "use a different identifier of the function as alias to avoid ambiguouity")
		}
	default:
	}
	return err
}

type recommandedAlias struct {
	ast.Visitor
	*rule
}

func RecommandedAlias(level Severity) Rule {
	a := &recommandedAlias{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, aliasRecommanded, level)
	return a
}

func (r *recommandedAlias) VisitSelect(stmt *ast.SelectStatement) error {
	var err error
	for _, q := range stmt.Columns {
		switch q.(type) {
		case *ast.Call, *ast.Group, *ast.Binary, *ast.Unary:
			err = r.Report(q, "alias is recommanded for function call, subquery, binary and/or unary expression")
		default:
		}
		if err != nil {
			break
		}
	}
	return err
}

type missingAlias struct {
	ast.Visitor
	*rule
}

// Creates a Rule that ensures all fields and tables are assigned an alias.
func MissingAlias(level Severity) Rule {
	a := &missingAlias{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, aliasMissing, level)
	return a
}

func (r *missingAlias) VisitSelect(stmt *ast.SelectStatement) error {
	for _, c := range slices.Concat(stmt.Columns, stmt.Tables) {
		if _, ok := c.(*ast.Alias); !ok {
			err := r.Report(c, "prefer using alias to improve readability of your query")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type noAlias struct {
	ast.Visitor
	*rule
}

// Creates a Rule that ensures no fields or tables are assigned an alias.
func NoAlias(level Severity) Rule {
	a := &noAlias{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, aliasNoAlias, level)
	return a
}

func (r *noAlias) VisitSelect(stmt *ast.SelectStatement) error {
	for _, c := range slices.Concat(stmt.Columns, stmt.Tables) {
		if a, ok := c.(*ast.Alias); ok {
			err := r.Report(a, "alias are not recommended unless needed")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type invalidAlias struct {
	ast.Visitor
	*rule

	aliases [][]ast.Identifier
}

// Creates a Rule that checks if alias defined in the SELECT clause are
// not used in where/having/group by clauses
func InvalidAlias(level Severity) Rule {
	a := &invalidAlias{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, aliasInvalid, level)
	return a
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
	var err error
	if r.exists(name) {
		err = r.Report(name, "alias is not expected in group by/where/having clause of query")
	}
	return err
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
	*rule

	aliases [][]ast.Identifier
}

// Creates a rule that checks whether the qualified fields in the
// SELECT clause use the table aliases defined in the query.
func UndefinedAlias(level Severity) Rule {
	a := &undefinedAlias{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, aliasUndefined, level)
	return a
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
	var err error
	if !r.exists(name) {
		err = r.Report(name, "alias is not defined in from clause of query")
	}
	return err
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
	*rule
}

func UnusedAlias(level Severity) Rule {
	a := &unusedAlias{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, aliasUnused, level)
	return a
}

func (r *unusedAlias) VisitSelect(stmt *ast.SelectStatement) error {
	var (
		alias     = make(map[ast.Identifier]int)
		positions = make(map[ast.Identifier]token.Position)
		join      []ast.Node
	)
	for _, n := range stmt.Tables {
		if j, ok := n.(*ast.Join); ok {
			n = j.Table
			join = append(join, j.Where)
		}
		if a, ok := n.(*ast.Alias); ok {
			alias[a.Identifier] = 0
			positions[a.Identifier] = a.Pos()
		}
	}
	if len(alias) == 0 {
		return nil
	}
	var (
		check = func(name *ast.Name) error {
			if len(name.Parts) == 1 {
				return nil
			}
			id := name.Parts[0]
			alias[id]++
			return nil
		}
		where  = slx.One(stmt.Where)
		having = slx.One(stmt.Having)
		parts  = slices.Concat(stmt.Columns, stmt.Groups, where, having, join)
		visit  = ast.VisitName(check)
		sub    = ast.Walk(visit)
	)
	for _, q := range parts {
		if q == nil {
			continue
		}
		if err := q.Accept(sub); err != nil {
			return err
		}
	}
	for _, count := range alias {
		if count == 0 {
			err := r.Report(stmt, "alias declared and not used")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
