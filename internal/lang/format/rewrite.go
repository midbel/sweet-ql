package format

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
)

type Rewriter interface {
	Rewrite(ast.Node) (ast.Node, error)
}

type NameFunc func(int, ast.Node) string

func NameWithPrefix(prefix string) NameFunc {
	return func(ix int, _ ast.Node) string {
		return fmt.Sprintf("%s%03d", prefix, ix)
	}
}

func SelfName(ix int, node ast.Node) string {
	switch n := node.(type) {
	case *ast.Name:
		if n.All() {
			return fmt.Sprintf("s%03d", ix)
		}
		return n.Parts[len(n.Parts)-1].Name
	case *ast.Call:
		return fmt.Sprintf("%s%03d", n.GetIdent(), ix)
	default:
	}
	return ""
}

var factory = map[string]func() Rewriter{
	"std-operator":         StdOperator,
	"cte-to-subquery":      nil,
	"subquery-to-cte":      nil,
	"groupby-field":        GroupbyFields,
	"groupby-pos":          GroupbyPosToFields,
	"groupby-aggr":         nil,
	"missing-alias-fields": MissingAliasFields,
	"missing-alias-tables": MissingAliasTables,
	"missing-cte-fields":   MissingCteFields,
	"missing-view-fields":  MissingViewFields,
	"placeholder-join":     ReplaceLiteralJoin,
	"placeholder-where":    ReplaceLiteralWhere,
	"no-returning":         NoReturning,
	"limit-to-fetch":       nil,
	"add-primary-key":      nil,
	"ansi":                 RewriteAnsi,
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

type rewriteAnsi struct {
	inner []Rewriter
}

func RewriteAnsi() Rewriter {
	all := []Rewriter{
		StdOperator(),
		GroupbyFields(),
		NoReturning(),
	}
	return rewriteAnsi{
		inner: all,
	}
}

func (r rewriteAnsi) Rewrite(stmt ast.Node) (ast.Node, error) {
	var err error
	for i := range r.inner {
		stmt, err = r.inner[i].Rewrite(stmt)
		if err != nil {
			break
		}
	}
	return stmt, err
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

func GroupbyFields() Rewriter {
	return rewriteGroupbyFields{
		Transformer: ast.Keep(),
	}
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
	if len(stmt.Groups) == 0 {
		return stmt, nil
	}
	for i, c := range stmt.Columns {
		err := r.addMissingField(c, i+1, stmt)
		if err != nil {
			return stmt, err
		}
	}
	return stmt, nil
}

func (r rewriteGroupbyFields) addMissingField(c ast.Node, pos int, stmt *ast.SelectStatement) error {
	if a, ok := c.(*ast.Alias); ok {
		c = a.Node
	}
	var id []ast.Identifier
	switch v := c.(type) {
	case *ast.Name:
		id = v.Parts
	case *ast.Call:
		if lang.IsAggregateFunc(v.GetIdent()) {
			return nil
		}
		if len(v.Args) == 0 {
			return fmt.Errorf("function without argument")
		}
		x, ok := v.Args[0].(*ast.Name)
		if !ok {
			return fmt.Errorf("first argument expected to be an identifier")
		}
		id, c = x.Parts, x
	default:
		return nil
	}
	ok := slices.ContainsFunc(stmt.Groups, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Name:
			return slices.Equal(id, n.Parts)
		case *ast.Value:
			x, err := strconv.Atoi(n.Literal)
			if err != nil {
				return false
			}
			return x == pos
		default:
			return false
		}
	})
	if !ok {
		stmt.Groups = append(stmt.Groups, c)
	}
	return nil
}

// when position are used in group by, replace by the identifier
type rewriteGroupbyPosField struct {
	ast.Transformer
}

func GroupbyPosToFields() Rewriter {
	return rewriteGroupbyPosField{
		Transformer: ast.Keep(),
	}
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
	for i, g := range stmt.Groups {
		n, ok := g.(*ast.Value)
		if !ok {
			continue
		}
		x, err := strconv.Atoi(n.Literal)
		if err != nil {
			continue
		}
		x--
		if x < 0 || x >= len(stmt.Columns) {
			return nil, fmt.Errorf("no column at given position")
		}
		c := stmt.Columns[x]
		if a, ok := c.(*ast.Alias); ok {
			c = a.Node
		}
		switch c := c.(type) {
		case *ast.Name:
			stmt.Groups[i] = c
		case *ast.Call:
		default:
		}
	}
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
	name NameFunc
}

func MissingAliasFields() Rewriter {
	return rewriteMissingAlias{
		Transformer: ast.Keep(),
		mode:        AliasFields,
		name:        NameWithPrefix("sel"),
	}
}

func MissingAliasTables() Rewriter {
	return rewriteMissingAlias{
		Transformer: ast.Keep(),
		mode:        AliasTables,
		name:        NameWithPrefix("tbl"),
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
	if r.mode == AliasFields || r.mode == AliasBoth {
		r.addMissingAliasToFields(stmt)
	}
	if r.mode == AliasTables || r.mode == AliasBoth {
		r.addMissingAliasToTables(stmt)
	}
	return stmt, nil
}

func (r rewriteMissingAlias) TransformCte(stmt *ast.CteStatement) (ast.Node, error) {
	node, err := r.Rewrite(stmt.Node)
	if err == nil {
		stmt.Node = node
	}
	return stmt, err
}

func (r rewriteMissingAlias) addMissingAliasToFields(stmt *ast.SelectStatement) {
	for i, c := range stmt.Columns {
		if _, ok := c.(*ast.Alias); ok {
			continue
		}
		if n, ok := c.(*ast.Name); ok && n.All() {
			continue
		}
		id := ast.Identifier{
			Name: r.name(i+1, c),
		}
		as := ast.Alias{
			Identifier: id,
			Node:       c,
			Position:   c.Pos(),
		}
		stmt.Columns[i] = &as
	}
}

func (r rewriteMissingAlias) addMissingAliasToTables(stmt *ast.SelectStatement) {
	for i, t := range stmt.Tables {
		_, _ = i, t
	}
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
	name NameFunc
}

func MissingCteFields() Rewriter {
	return rewriteMissingColumnsNames{
		Transformer: ast.Keep(),
		mode:        CteMissing,
		name:        NameWithPrefix("cte"),
	}
}

func MissingViewFields() Rewriter {
	return rewriteMissingColumnsNames{
		Transformer: ast.Keep(),
		mode:        ViewMissing,
		name:        NameWithPrefix("view"),
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
	if r.mode != CteMissing && r.mode != AllMissing {
		return stmt, nil
	}
	names, err := r.getColumns(stmt.Node)
	if err != nil {
		return nil, err
	}
	stmt.Columns = names
	return stmt, nil
}

func (r rewriteMissingColumnsNames) TransformCreateView(stmt *ast.CreateViewStatement) (ast.Node, error) {
	if r.mode != ViewMissing && r.mode != AllMissing {
		return stmt, nil
	}
	names, err := r.getColumns(stmt.Select)
	if err != nil {
		return nil, err
	}
	stmt.Columns = names
	return stmt, nil
}

func (r rewriteMissingColumnsNames) getColumns(stmt ast.Node) ([]ast.Node, error) {
	q, ok := stmt.(*ast.SelectStatement)
	if !ok {
		return nil, fmt.Errorf("unexpected query type")
	}
	var (
		names []ast.Node
		seen  = make(map[ast.Identifier]struct{})
	)
	for i, c := range q.Columns {
		var n ast.Name
		switch x := c.(type) {
		case *ast.Value:
			id := ast.Identifier{
				Name: r.name(i+1, c),
			}
			n.Parts = append(n.Parts, id)
		case *ast.Name:
			if x.All() {
				return nil, nil
			}
			n.Parts = append(n.Parts, x.Parts[len(x.Parts)-1])
		case *ast.Alias:
			n.Parts = append(n.Parts, x.Identifier)
		case *ast.Call:
			id := ast.Identifier{
				Name: r.name(i+1, c),
			}
			n.Parts = append(n.Parts, id)
		default:
			id := ast.Identifier{
				Name: r.name(i+1, c),
			}
			n.Parts = append(n.Parts, id)
		}
		id := n.Parts[len(n.Parts)-1]
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("duplicate identifier")
		}
		seen[id] = struct{}{}
		names = append(names, &n)
	}
	return names, nil
}

type LiteralMode int8

const (
	LiteralWhere LiteralMode = 1 << iota
	LiteralJoin
	LiteralAll
)

type PlaceholderType int8

const (
	TypeClassic PlaceholderType = 1 << iota
	TypeName
	TypePosition
)

// rewrite literal value in join with placeholders
type rewriteLiteralWithPlaceholder struct {
	ast.Transformer
	mode LiteralMode
	kind PlaceholderType
}

func ReplaceLiteralJoin() Rewriter {
	return rewriteLiteralWithPlaceholder{
		Transformer: ast.Keep(),
		mode:        LiteralJoin,
		kind:        TypeClassic,
	}
}

func ReplaceLiteralWhere() Rewriter {
	return rewriteLiteralWithPlaceholder{
		Transformer: ast.Keep(),
		mode:        LiteralWhere,
		kind:        TypeClassic,
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

func (r rewriteLiteralWithPlaceholder) TransformSelect(stmt *ast.SelectStatement) (ast.Node, error) {
	if r.mode == LiteralWhere || r.mode == LiteralAll {
		where, err := r.replaceWithPlaceholder(stmt.Where)
		if err == nil {
			stmt.Where = where
		}
		return stmt, err
	}
	return stmt, nil
}

func (r rewriteLiteralWithPlaceholder) TransformJoin(join *ast.Join) (ast.Node, error) {
	if r.mode == LiteralJoin || r.mode == LiteralAll {
		where, err := r.replaceWithPlaceholder(join.Where)
		if err == nil {
			join.Where = where
		}
		return join, err
	}
	return join, nil
}

func (r rewriteLiteralWithPlaceholder) replaceWithPlaceholder(node ast.Node) (ast.Node, error) {
	replace := replaceWithPlaceholder{
		Transformer: r.Transformer,
	}
	return replace.Rewrite(node)
}

type replaceWithPlaceholder struct {
	ast.Transformer
	kind PlaceholderType
}

func (r replaceWithPlaceholder) Rewrite(stmt ast.Node) (ast.Node, error) {
	t, ok := stmt.(ast.TransformableNode)
	if !ok {
		return stmt, nil
	}
	walker := ast.Transform(r)
	return t.Transform(walker)
}

func (r replaceWithPlaceholder) TransformBinary(binary *ast.Binary) (ast.Node, error) {
	if _, ok := binary.Left.(*ast.Value); ok {
		binary.Left = r.create(binary.Left)
	}
	if _, ok := binary.Right.(*ast.Value); ok {
		binary.Right = r.create(binary.Right)
	}
	return binary, nil
}

func (r replaceWithPlaceholder) TransformBetween(between *ast.Between) (ast.Node, error) {
	if _, ok := between.Lower.(*ast.Value); ok {
		between.Lower = r.create(between.Lower)
	}
	if _, ok := between.Upper.(*ast.Value); ok {
		between.Upper = r.create(between.Upper)
	}
	return between, nil
}

func (r replaceWithPlaceholder) TransformIs(is *ast.Is) (ast.Node, error) {
	if _, ok := is.Value.(*ast.Value); ok {
		is.Value = r.create(is.Value)
	}
	return is, nil
}

func (r replaceWithPlaceholder) TransformIn(in *ast.In) (ast.Node, error) {
	return in, nil
}

func (r replaceWithPlaceholder) TransformNot(not *ast.Not) (ast.Node, error) {
	n, err := r.Rewrite(not.Node)
	if err != nil {
		return nil, err
	}
	not.Node = n
	return not, nil
}

func (r replaceWithPlaceholder) create(n ast.Node) ast.Node {
	return &ast.Placeholder{
		Position: n.Pos(),
	}
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
	stmt.Limit = r.rewriteLimit(stmt.Limit)
	return stmt, nil
}

func (r rewriteLimitToFetch) rewriteLimit(node ast.Node) ast.Node {
	if node == nil {
		return nil
	}
	limit, ok := node.(*ast.Limit)
	if !ok {
		return node
	}
	offset := ast.Offset{
		Position: limit.Position,
		Count:    limit.Count,
		Offset:   limit.Offset,
	}
	return &offset
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
	stmt.Returning = nil
	return stmt, nil
}

func (r rewriteReturning) TransformUpdate(stmt *ast.UpdateStatement) (ast.Node, error) {
	stmt.Returning = nil
	return stmt, nil
}

func (r rewriteReturning) TransformDelete(stmt *ast.DeleteStatement) (ast.Node, error) {
	stmt.Returning = nil
	return stmt, nil
}
