package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

type Rewriter interface {
	Rewrite(ast.Node) (ast.Node, error)
}

// rewrite operator to std one like != to <> and = true to is true
type rewriteStdOperator struct {
	ast.Transformer
}

func StdOperator() Rewriter {
	return rewriteStdOperator{
		Transformer: ast.Keep(),
	}
}

func (r rewriteStdOperator) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r rewriteStdOperator) TransformBinary(binary *ast.Binary) (ast.Node, error) {
	if binary.IsRelation() {
		return binary, nil
	}
	if binary.Op == "!=" {
		binary.Op = "<>"
	}
	if v, ok := binary.Right.(*ast.Value); ok && v.Constant() {
		is := &ast.Is{
			Position: binary.Pos(),
			Ident:    binary.Left,
			Value:    binary.Right,
		}
		switch binary.Op {
		case "=":
		case "<>":
			not := &ast.Not{
				Position: binary.Pos(),
				Node:     is,
			}
			return not, nil
		default:
		}
		return is, nil
	}
	return binary, nil
}

// transform all subqueries to an equivalent cte
type rewriteSubqueryToCte struct {
	ast.Visitor
}

// transform all cte to subquery where they are used
type rewriteCteToSubquery struct {
	ast.Visitor
}

// add fields not use in aggregate function in the select clause into the group by
type rewriteGroupbyFields struct {
	ast.Visitor
}

func (r rewriteGroupbyFields) VisitSelect(stmt *ast.SelectStatement) error {
	return nil
}

// when position are used in group by, replace by the identifier
type rewriteGroupbyPosField struct {
	ast.Visitor
}

func (r rewriteGroupbyPosField) VisitSelect(stmt *ast.SelectStatement) error {
	return nil
}

// add alias to all columns in select clause
type rewriteMissingAlias struct {
	ast.Visitor
}

// add columns definition list to cte
type rewriteMissingColumnsNames struct {
	ast.Visitor
}

func (r rewriteMissingColumnsNames) VisitCte(stmt *ast.CteStatement) error {
	return nil
}

func (r rewriteMissingColumnsNames) VisitCreateView(stmt *ast.CreateViewStatement) error {
	return nil
}

// rewrite literal value in join with placeholders
type rewriteLiteralWithPlaceholderJoin struct {
	ast.Visitor
}

// rewrite literal value in expression with placeholders
type rewriteLiteralWithPlaceholderExpr struct {
	ast.Visitor
}

// rewrite use of limit/offset to offset/fetch
type rewriteLimitToFetch struct {
	ast.Visitor
}

func (r rewriteLimitToFetch) VisitSelect(stmt *ast.SelectStatement) error {
	return nil
}
