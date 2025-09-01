package lint

import (
	"errors"
	"fmt"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

type noCte struct {
	ast.Visitor
	issues   []Issue
	severity Severity
}

// Creates a Rule that checks that no WITH statement are used
func NoCte(level Severity) Rule {
	return &noCte{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *noCte) Name() string {
	return "no-cte"
}

func (r *noCte) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]
	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *noCte) VisitWith(with *ast.WithStatement) error {
	i := Issue{
		Position: with.Pos(),
		Severity: r.severity,
		Rule:     r.Name(),
		Reason:   "prefer using subqueries over common table expression",
	}
	r.issues = append(r.issues, i)
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
	severity Severity
	issues   []Issue
}

// Creates a Rule that will check that name given to a cte will not
// shadow the name of a table used in the cte
func CteShadow(level Severity) Rule {
	return &cteShadow{
		Visitor:  ast.Noop(),
		severity: level,
	}
}

func (_ *cteShadow) Name() string {
	return "cte-shadow"
}

func (r *cteShadow) Verify(stmt ast.Node) ([]Issue, error) {
	r.issues = r.issues[:0]

	err := stmt.Accept(ast.Walk(r))
	if errors.Is(err, ast.ErrStop) {
		err = nil
	}
	return r.issues, err
}

func (r *cteShadow) VisitCte(cte *ast.CteStatement) error {
	check := func(stmt *ast.SelectStatement) error {
		for _, n := range stmt.Tables {
			if a, ok := n.(*ast.Alias); ok {
				n = a.Node
			}
			if n, ok := n.(*ast.Name); ok {
				if n.Parts[len(n.Parts)-1].Name == cte.Ident {
					i := Issue{
						Position: cte.Node.Pos(),
						Severity: r.severity,
						Rule:     r.Name(),
						Reason:   "cte name will shadow the table name",
					}
					r.issues = append(r.issues, i)
				}
			}
		}
		return nil
	}
	var (
		visit = visitCteSelect(check)
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
		return nil
	}
	collect := func(stmt *ast.SelectStatement) error {
		for _, c := range stmt.Columns {
			var id ast.Identifier
			switch c := c.(type) {
			case *ast.Name:
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
		visit = visitCteSelect(collect)
		walk  = ast.Walk(visit)
	)
	return cte.Node.Accept(walk)
}

func (r *cteNames) VisitSelect(stmt *ast.SelectStatement) error {
	return nil
}

type cteSelectVisitor struct {
	ast.Visitor
	do func(*ast.SelectStatement) error
}

func visitCteSelect(do func(*ast.SelectStatement) error) ast.Visitor {
	return &cteSelectVisitor{
		Visitor: ast.Noop(),
		do:      do,
	}
}

func (i *cteSelectVisitor) VisitSelect(stmt *ast.SelectStatement) error {
	return i.do(stmt)
}
