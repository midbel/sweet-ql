package format

import (
	"fmt"

	"github.com/midbel/sweet/internal/lang/ast"
)

type Rewriter interface {
	Rewrite(ast.Node) (ast.Node, error)
}

var factory = map[string]func() Rewriter{
	"std-operator":         StdOperator,
	"cte-to-subquery":      nil,
	"subquery-to-cte":      nil,
	"groupby-field":        nil,
	"groupby-aggr":         nil,
	"groupby-pos":          nil,
	"missing-alias-fields": MissingAliasFields,
	"missing-alias-tables": MissingAliasTables,
	"missing-cte-fields":   MissingCteFields,
	"missing-view-fields":  MissingViewFields,
	"placeholder-join":     ReplaceLiteralJoin,
	"placeholder-where":    ReplaceLiteralWhere,
	"no-returning":         NoReturning,
	"limit-to-fetch":       nil,
	"add-primary-key":      nil,
}

func RewriterByName(name string) (Rewriter, error) {
	fn, ok := factory[name]
	if ok {
		if fn == nil {
			return nil, fmt.Errorf("%s: not yet implemented", name)
		}
		return fn(), nil
	}
	return nil, fmt.Errorf("%s: unsupported rewriter", name)
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
	ast.Transformer
}

func (r rewriteSubqueryToCte) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

// transform all cte to subquery where they are used
type rewriteCteToSubquery struct {
	ast.Transformer
}

func (r rewriteCteToSubquery) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

// add fields not use in aggregate function in the select clause into the group by
type rewriteGroupbyFields struct {
	ast.Transformer
}

func (r rewriteGroupbyFields) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r rewriteGroupbyFields) TransformSelect(stmt *ast.SelectStatement) (ast.Node, error) {
	return stmt, nil
}

// when position are used in group by, replace by the identifier
type rewriteGroupbyPosField struct {
	ast.Transformer
}

func (r rewriteGroupbyPosField) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r rewriteGroupbyPosField) TransformSelect(stmt *ast.SelectStatement) (ast.Node, error) {
	return stmt, nil
}

type AliasMode int8

const (
	AliasFields AliasMode = 1 << iota
	AliasTables
	AliasBoth
)

// add alias to all columns in select clause
type rewriteMissingAlias struct {
	ast.Transformer
	mode AliasMode
}

func MissingAliasFields() Rewriter {
	return rewriteMissingAlias{
		Transformer: ast.Keep(),
		mode:        AliasFields,
	}
}

func MissingAliasTables() Rewriter {
	return rewriteMissingAlias{
		Transformer: ast.Keep(),
		mode:        AliasTables,
	}
}

func (r rewriteMissingAlias) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r rewriteMissingAlias) TransformSelect(stmt *ast.SelectStatement) (ast.Node, error) {
	return stmt, nil
}

type MissingMode int8

const (
	CteMissing MissingMode = 1 << iota
	ViewMissing
	AllMissing
)

// add columns definition list to cte
type rewriteMissingColumnsNames struct {
	ast.Transformer
	mode MissingMode
}

func MissingCteFields() Rewriter {
	return rewriteMissingColumnsNames{
		Transformer: ast.Keep(),
		mode:        CteMissing,
	}
}

func MissingViewFields() Rewriter {
	return rewriteMissingColumnsNames{
		Transformer: ast.Keep(),
		mode:        ViewMissing,
	}
}

func (r rewriteMissingColumnsNames) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r rewriteMissingColumnsNames) TransformCte(stmt *ast.CteStatement) (ast.Node, error) {
	return stmt, nil
}

func (r rewriteMissingColumnsNames) TransformCreateView(stmt *ast.CreateViewStatement) (ast.Node, error) {
	return stmt, nil
}

// rewrite literal value in join with placeholders
type rewriteLiteralWithPlaceholder struct {
	ast.Transformer
}

func ReplaceLiteralJoin() Rewriter {
	return rewriteLiteralWithPlaceholder{
		Transformer: ast.Keep(),
	}
}

func ReplaceLiteralWhere() Rewriter {
	return rewriteLiteralWithPlaceholder{
		Transformer: ast.Keep(),
	}
}

func (r rewriteLiteralWithPlaceholder) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

// rewrite use of limit/offset to offset/fetch
type rewriteLimitToFetch struct {
	ast.Transformer
}

func (r rewriteLimitToFetch) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r rewriteLimitToFetch) TransformSelect(stmt *ast.SelectStatement) (ast.Node, error) {
	return stmt, nil
}

type rewriteReturning struct {
	ast.Transformer
}

func NoReturning() Rewriter {
	return rewriteReturning{
		Transformer: ast.Keep(),
	}
}

func (r rewriteReturning) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r rewriteReturning) TransformInsert(stmt *ast.InsertStatement) (ast.Node, error) {
	return stmt, nil
}

func (r rewriteReturning) TransformUpdate(stmt *ast.UpdateStatement) (ast.Node, error) {
	return stmt, nil
}

func (r rewriteReturning) TransformDelete(stmt *ast.DeleteStatement) (ast.Node, error) {
	return stmt, nil
}
