package ast

import (
	"slices"
	"strings"
)

type Commented struct {
	Before []string
	After  string
	Statement
}

type Group struct {
	Statement
}

type Cast struct {
	Ident Statement
	Type  Type
}

type Type struct {
	Name      string
	Length    int
	Precision int
}

type Not struct {
	Statement
}

func (n Not) GetNames() []string {
	return GetNamesFromStmt([]Statement{n.Statement})
}

type Collate struct {
	Statement
	Collation string
}

type Exists struct {
	Statement
}

var sqlAggregates = []string{
	"max",
	"min",
	"avg",
	"sum",
	"count",
}

var sqlBuiltins = []string{
	"max",
	"min",
	"avg",
	"sum",
	"count",
}

type Call struct {
	Distinct bool
	Ident    Statement
	Args     []Statement
	Filter   Statement
	Over     Statement
}

func (c Call) GetNames() []string {
	return GetNamesFromStmt(c.Args)
}

func (c Call) GetIdent() string {
	n, ok := c.Ident.(Name)
	if !ok {
		return "?"
	}
	return n.Ident()
}

func (c Call) IsAggregate() bool {
	return slices.Contains(sqlAggregates, c.GetIdent())
}

func (c Call) BuiltinSql() bool {
	return slices.Contains(sqlBuiltins, c.GetIdent())
}

type Row struct {
	Values []Statement
}

func (r Row) Keyword() (string, error) {
	return "ROW", nil
}

type Unary struct {
	Right Statement
	Op    string
}

func (u Unary) GetNames() []string {
	return GetNamesFromStmt([]Statement{u.Right})
}

type Binary struct {
	Left  Statement
	Right Statement
	Op    string
}

func (b Binary) GetNames() []string {
	var list []string
	list = append(list, GetNamesFromStmt([]Statement{b.Left})...)
	list = append(list, GetNamesFromStmt([]Statement{b.Right})...)
	return list
}

func (b Binary) IsRelation() bool {
	return b.Op == "AND" || b.Op == "OR"
}

type All struct {
	Statement
}

type Any struct {
	Statement
}

type Is struct {
	Ident Statement
	Value Statement
}

type In struct {
	Ident Statement
	Value Statement
}

func (i In) GetNames() []string {
	var list []string
	list = append(list, GetNamesFromStmt([]Statement{i.Ident})...)
	list = append(list, GetNamesFromStmt([]Statement{i.Value})...)
	return list
}

type Between struct {
	Not   bool
	Ident Statement
	Lower Statement
	Upper Statement
}

type List struct {
	Values []Statement
}

func (i List) Len() int {
	return len(i.Values)
}

type Value struct {
	Literal string
}

type Alias struct {
	Statement
	Alias string
}

type Name struct {
	Parts []string
}

func (n Name) All() bool {
	return false
}

func (n Name) Schema() string {
	switch len(n.Parts) {
	case 2:
		return n.Parts[0]
	case 3:
		return n.Parts[1]
	default:
		return ""
	}
}

func (n Name) Name() string {
	if len(n.Parts) == 0 {
		return "*"
	}
	str := n.Parts[len(n.Parts)-1]
	if str == "" {
		str = "*"
	}
	return str
}

func (n Name) Ident() string {
	z := len(n.Parts)
	if z == 0 {
		return "*"
	}
	if n.Parts[z-1] == "" {
		n.Parts[z-1] = "*"
	}
	return strings.Join(n.Parts, ".")
}