package ast

import (
	"strings"

	"github.com/midbel/sweet/internal/token"
)

type Group struct {
	Statement
}

type Cast struct {
	token.Position

	Ident Statement
	Type  Type
}

type Type struct {
	token.Position

	Name      string
	Length    int
	Precision int
}

type Not struct {
	token.Position
	Statement
}

type Collate struct {
	token.Position
	Statement
	Collation string
}

type Exists struct {
	token.Position
	Statement
}

type Call struct {
	Position token.Position
	Distinct bool
	Ident    Statement
	Args     []Statement
	Filter   Statement
	Over     Statement
}

func (c Call) GetIdent() string {
	n, ok := c.Ident.(Name)
	if !ok {
		return ""
	}
	return n.Ident()
}

type Row struct {
	token.Position
	Values []Statement
}

func (r Row) Keyword() (string, error) {
	return "ROW", nil
}

type Unary struct {
	token.Position
	Right Statement
	Op    string
}

type Binary struct {
	token.Position
	Left  Statement
	Right Statement
	Op    string
}

func (b Binary) IsRelation() bool {
	return b.Op == "AND" || b.Op == "OR"
}

type All struct {
	token.Position
	Statement
}

type Any struct {
	token.Position
	Statement
}

type Is struct {
	token.Position
	Ident Statement
	Value Statement
}

type In struct {
	token.Position
	Ident Statement
	Value Statement
}

type Between struct {
	token.Position
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

type Placeholder struct {
	token.Position
	Statement
}

type Value struct {
	token.Position
	Literal string
}

func (v Value) Constant() bool {
	return v.Null() || v.True() || v.False()
}

func (v Value) Null() bool {
	return v.Literal == "NULL"
}

func (v Value) True() bool {
	return v.Literal == "TRUE"
}

func (v Value) False() bool {
	return v.Literal == "FALSE"
}

type Alias struct {
	token.Position
	Statement
	Identifier
}

type Identifier struct {
	Quoted bool
	Name   string
}

type Name struct {
	token.Position
	Parts []Identifier
}

func (n Name) All() bool {
	return false
}

func (n Name) Schema() string {
	switch len(n.Parts) {
	case 2:
		return n.Parts[0].Name
	case 3:
		return n.Parts[1].Name
	default:
		return ""
	}
}

func (n Name) Name() string {
	if len(n.Parts) == 0 {
		return "*"
	}
	str := n.Parts[len(n.Parts)-1].Name
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
	if n.Parts[z-1].Name == "" {
		n.Parts[z-1].Name = "*"
	}
	var parts []string
	for _, i := range n.Parts {
		parts = append(parts, i.Name)
	}
	return strings.Join(parts, ".")
}
