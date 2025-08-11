package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) Rewrite(stmt ast.Node) (ast.Node, error) {
	if w.Rules.KeepAsIs() {
		return stmt, nil
	}
	return w.rewrite(stmt)
}

func (w *Writer) rewrite(stmt ast.Node) (ast.Node, error) {
	return stmt, nil
}

// rewrite != to <> and x = null|true|false to x is null|true|false
func (w *Writer) rewriteStdOperator(stmt ast.Node) (ast.Node, error) {
	return nil
}

// replace any subqueries in a sql query in a with statement
func rewriteSubqueryAsCte(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

// replace all cte from a with statement and put it as subqueries
func rewriteCteAsSubquery(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

// add in fields list missing names present in the group by clause
func rewriteGroupBy(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

// simplify boolean expression x is true becomes x
func rewriteBooleanExpr(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

// set as keyword to all fields in select and in tables
func rewriteAlias(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

// for all fields and tables without aliases, create one
func rewriteMissingAlias(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

// foreach cte in a with statement, add a columns list definition if not set
func rewriteMissingCteColumns(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

// add a columns list definition if not set in a create view statement
func rewriteMissingViewColumns(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}

func replaceLiteralWithPlaceholder(stmt ast.Node) (ast.Node, error) {
	return nil, nil
}
