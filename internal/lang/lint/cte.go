package lint

import (
	"errors"
	"fmt"
	"slices"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type cteSelectOnly struct {
	ast.Visitor
	*rule
}

// Creates a rule that verifies that queries in CTE and WITH statement are only SELECT
// queries.
func CteOnlySelect(level Severity) Rule {
	a := &cteSelectOnly{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "cte-select-only", level)
	return a
}

func (r *cteSelectOnly) VisitWith(stmt *ast.WithStatement) error {
	return r.check(stmt.Node)
}

func (r *cteSelectOnly) VisitCte(stmt *ast.CteStatement) error {
	return r.check(stmt.Node)
}

func (r *cteSelectOnly) check(stmt ast.Node) error {
	if _, ok := stmt.(*ast.SelectStatement); !ok {
		return r.Report(stmt, "queries inside with statements should be select query")
	}
	return nil
}

type noCte struct {
	ast.Visitor
	*rule
}

// Creates a Rule that checks that no WITH statement are used
func NoCte(level Severity) Rule {
	a := &noCte{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "no-cte", level)
	return a
}

func (r *noCte) VisitWith(with *ast.WithStatement) error {
	err := r.Report(with, "prefer using subqueries over common table expression")
	if err != nil {
		return err
	}
	return ast.ErrStop
}

type cteUnused struct {
	ast.Visitor
	severity Severity

	names     map[string]int
	positions map[string]token.Position
	collect   bool
}

// Creates a Rule that check that all cte declared are used in the main
// SELECT of the query
func CteUnused(level Severity) Rule {
	return &cteUnused{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *cteUnused) Name() string {
	return "cte-unused"
}

func (r *cteUnused) Verify(stmt ast.Node) ([]Issue, error) {
	r.names = make(map[string]int)
	r.positions = make(map[string]token.Position)

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}

	var issues []Issue
	for n, c := range r.names {
		if c > 0 {
			continue
		}
		i := Issue{
			Position: r.positions[n],
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "cte is defined but not used",
		}
		issues = append(issues, i)
	}

	return issues, err
}

func (r *cteUnused) VisitCte(stmt *ast.CteStatement) error {
	r.names[stmt.Ident] = 0
	r.positions[stmt.Ident] = stmt.Pos()
	return nil
}

func (r *cteUnused) VisitSelect(stmt *ast.SelectStatement) error {
	r.begin()
	defer r.end()

	sub := ast.Walk(r)
	for _, t := range stmt.Tables {
		t.Accept(sub)
	}
	return nil
}

func (r *cteUnused) VisitName(name *ast.Name) error {
	r.update(name.Name())
	return nil
}

func (r *cteUnused) begin() {
	r.collect = true
}

func (r *cteUnused) end() {
	r.collect = false
}

func (r *cteUnused) update(name string) {
	if !r.collect {
		return
	}
	if _, ok := r.names[name]; ok {
		r.names[name]++
	}
}

type cteShadow struct {
	ast.Visitor
	*rule
}

// Creates a Rule that will check that name given to a cte will not
// shadow the name of a table used in the cte
func CteShadow(level Severity) Rule {
	a := &cteShadow{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "cte-shadow", level)
	return a
}

func (_ *cteShadow) Name() string {
	return "cte-shadow"
}

func (r *cteShadow) VisitCte(cte *ast.CteStatement) error {
	check := func(stmt *ast.SelectStatement) error {
		for _, n := range stmt.Tables {
			if a, ok := n.(*ast.Alias); ok {
				n = a.Node
			}
			if n, ok := n.(*ast.Name); ok {
				if n.Parts[len(n.Parts)-1].Name == cte.Ident {
					err := r.Report(n, "cte name will shadow the table name")
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	var (
		visit = ast.VisitSelect(check)
		walk  = ast.Walk(visit)
	)
	return cte.Node.Accept(walk)
}

type cteNames struct {
	ast.Visitor
	severity Severity
	issues   []Issue

	names map[string][]ast.Identifier
}

// Creates a Rule that checks if fields used in a SELECT clause are
// exposed by a cte
func CteNames(level Severity) Rule {
	return &cteNames{
		Visitor:  ast.Noop(),
		severity: level,
		names:    make(map[string][]ast.Identifier),
	}
}

func (_ *cteNames) Name() string {
	return "cte-name"
}

func (r *cteNames) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *cteNames) VisitCte(cte *ast.CteStatement) error {
	if len(cte.Columns) > 0 {
		var all []ast.Identifier
		for _, n := range cte.Columns {
			n, ok := n.(*ast.Name)
			if !ok {
				return fmt.Errorf("name expected")
			}
			id := n.Parts[len(n.Parts)-1]
			all = append(all, id)
		}
		r.names[cte.Ident] = all
		return ast.ErrVisit
	}
	collect := func(stmt *ast.SelectStatement) error {
		for _, c := range stmt.Columns {
			var id ast.Identifier
			switch c := c.(type) {
			case *ast.Name:
				if c.All() {
					continue
				}
				id = c.Parts[len(c.Parts)-1]
			case *ast.Alias:
				id = c.Identifier
			default:
				continue
			}
			r.names[cte.Ident] = append(r.names[cte.Ident], id)
		}
		return nil
	}
	var (
		visit = ast.VisitSelect(collect)
		walk  = ast.Walk(visit)
	)
	if err := cte.Node.Accept(walk); err != nil {
		return err
	}
	return ast.ErrVisit
}

func (r *cteNames) VisitSelect(stmt *ast.SelectStatement) error {
	var (
		tables = r.getFieldsFromTables(stmt.Tables)
		check  = func(n *ast.Name) error {
			if len(n.Parts) <= 1 {
				r.checkFromAll(n)
			} else {
				r.checkFromNames(n, tables)
			}
			return nil
		}
	)
	for _, c := range stmt.Columns {
		if a, ok := c.(*ast.Alias); ok {
			c = a.Node
		}
		n, ok := c.(*ast.Name)
		if !ok {
			continue
		}
		if n.All() {
			continue
		}
		if err := check(n); err != nil {
			return err
		}
	}
	var (
		where  = slx.One(stmt.Where)
		having = slx.One(stmt.Having)
		parts  = slices.Concat(stmt.Groups, where, having)
		visit  = ast.VisitName(check)
		sub    = ast.Walk(visit)
	)
	for _, t := range stmt.Tables {
		j, ok := t.(*ast.Join)
		if ok && j.Where != nil {
			if err := j.Where.Accept(sub); err != nil {
				return err
			}
		}
	}
	for _, q := range parts {
		if q == nil {
			continue
		}
		if err := q.Accept(sub); err != nil {
			return err
		}
	}
	return nil
}

func (r *cteNames) checkFromNames(n *ast.Name, tables map[string][]ast.Identifier) {
	if n.All() {
		return
	}
	columns, ok := tables[n.Schema()]
	if !ok || len(columns) == 0 {
		return
	}
	ok = slices.ContainsFunc(columns, func(c ast.Identifier) bool {
		return !c.Star() && c.Name == n.Name()
	})
	if !ok {
		i := Issue{
			Position: n.Pos(),
			Severity: r.severity,
			Rule:     r.Name(),
			Reason:   "field name not exposed by any cte",
		}
		r.issues = append(r.issues, i)
	}
}

func (r *cteNames) checkFromAll(n *ast.Name) {
	if len(r.names) == 0 {
		return
	}
	for _, columns := range r.names {
		if len(columns) == 0 && len(r.names) == 1 {
			return
		}
		ok := slices.ContainsFunc(columns, func(c ast.Identifier) bool {
			return !c.Star() && c.Name == n.Name()
		})
		if ok {
			return
		}
	}
	i := Issue{
		Position: n.Pos(),
		Severity: r.severity,
		Rule:     r.Name(),
		Reason:   "field name not exposed by any cte",
	}
	r.issues = append(r.issues, i)
}

func (r *cteNames) getFieldsFromTables(nodes []ast.Node) map[string][]ast.Identifier {
	tables := make(map[string][]ast.Identifier)
	for _, t := range nodes {
		if j, ok := t.(*ast.Join); ok {
			t = j.Table
		}
		var alias *ast.Alias
		if a, ok := t.(*ast.Alias); ok {
			t = a.Node
			alias = a
		}
		n, ok := t.(*ast.Name)
		if !ok {
			continue
		}
		id := n.Parts[len(n.Parts)-1]
		if cs, ok := r.names[id.Name]; ok {
			if alias != nil {
				id = alias.Identifier
			}
			tables[id.Name] = cs
		}
	}
	return tables
}

type cteExposedNames struct {
	ast.Visitor
	*rule
}

func CteExposedNames(level Severity) Rule {
	a := &cteExposedNames{
		Visitor: ast.Noop(),
	}
	a.rule = stdRule(a, "cte-exposed-name", level)
	return a
}

func (r *cteExposedNames) VisitCte(stmt *ast.CteStatement) error {
	if len(stmt.Columns) == 0 {
		return nil
	}
	q, ok := stmt.Node.(*ast.SelectStatement)
	if !ok {
		return fmt.Errorf("%s: unexpected query type", r.Name())
	}
	var matched int
	for _, c := range q.Columns {
		var id ast.Identifier
		switch c := c.(type) {
		case *ast.Name:
			if c.All() {
				return nil
			}
			id = c.Parts[len(c.Parts)-1]
		case *ast.Alias:
			id = c.Identifier
		case *ast.Call:
			id = ast.Identifier{
				Name: c.GetIdent(),
			}
		default:
			continue
		}
		ok := slices.ContainsFunc(stmt.Columns, func(c ast.Node) bool {
			n, ok := c.(*ast.Name)
			if !ok {
				return ok
			}
			return n.Parts[0] == id
		})
		if ok {
			matched++
		}
	}
	if matched == len(stmt.Columns) {
		return r.Report(stmt, "declared names of cte identical to names in select clause")
	}
	return nil
}
