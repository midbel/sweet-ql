package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

// rewrite operator to std one like != to <> and = true to is true
type rewriteStdOperator struct {
	ast.Visitor
}

func (r rewriteStdOperator) VisitBinary(binary ast.Binary) error {
	return nil
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

// when position are used in group by, replace by the identifier
type rewriteGroupbyPosField struct {
	ast.Visitor
}

// simplify some boolean expression when they can be evaluate with only the identifier
type rewriteBooleanExpr struct {
	ast.Visitor
}

// set the as keyword
type rewriteSetAs struct {
	ast.Visitor
}

// add alias to all columns in select clause
type rewriteMissingAlias struct {
	ast.Visitor
}

// add columns definition list to cte
type rewriteMissingCteColumns struct {
	ast.Visitor
}

// add columns definition list to create view
type rewriteMissingViewColumns struct {
	ast.Visitor
}

// rewrite literal value in join with placeholders
type rewriteLiteralWithPlaceholderJoin struct {
	ast.Visitor
}

// rewrite literal value in expression with placeholders
type rewriteLiteralWithPlaceholderExpr struct {
	ast.Visitor
}
