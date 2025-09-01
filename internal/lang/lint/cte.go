package lint

import (
	"errors"

	"github.com/midbel/sweet/internal/lang/ast"
	"github.com/midbel/sweet/internal/token"
)

type noCte struct {
	ast.Visitor
	issues   []Issue
	severity Severity
}

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

type cteSelectVisitor struct {
	ast.Visitor
	check func(*ast.SelectStatement) error
}

func visitCteSelect(check func(*ast.SelectStatement) error) ast.Visitor {
	return &cteSelectVisitor{
		Visitor: ast.Noop(),
		check:   check,
	}
}

func (i *cteSelectVisitor) VisitSelect(stmt *ast.SelectStatement) error {
	return i.check(stmt)
}

type cteShadow struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

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

// check that fields qualified by cte are exposed by it
type cteNames struct {
	ast.Visitor
	severity Severity
	issues   []Issue
}

func CteNames(level Severity) Rule {
	return &cteNames{
		Visitor:  ast.Noop(),
		severity: level,
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
