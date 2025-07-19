package lint

import (
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
	return verify[ast.SelectStatement](stmt, r.checkStar)
}

func (r noStar) checkStar(q ast.SelectStatement) ([]Issue, error) {
	var (
		list  []Issue
		parts = slx.Make(q.Where, q.Having)
	)
	for _, c := range slices.Concat(q.Columns, q.Tables, parts) {
		if c == nil {
			continue
		}
		issues, err := r.checkStatement(c)
		if err != nil {
			return nil, err
		}
		list = slices.Concat(list, issues)
	}
	return list, nil
}

func (r noStar) checkStatement(stmt ast.Statement) ([]Issue, error) {
	var list []Issue
	for _, n := range collect(stmt) {
		switch n := n.(type) {
		case ast.Name:
			if n.Name() == "*" {
				i := Issue{
					Position: n.Position,
					Severity: r.severity,
					Rule:     r.Name(),
					Reason:   "use explicit field names",
				}
				list = append(list, i)
			}
		case ast.SelectStatement:
			issues, err := r.checkStar(n)
			if err != nil {
				return nil, err
			}
			list = slices.Concat(list, issues)
		}
	}
	return list, nil
}
