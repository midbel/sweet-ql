package ast

import (
	"strconv"
	"strings"

	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type Group struct {
	Node
}

func (g Group) Accept(visit Visitor) {
	visit.VisitGroup(g)
}

func (g Group) GetStatement() []Node {
	return slx.One(g.Node)
}

type Cast struct {
	token.Position

	Node
	Type Type
}

func (c Cast) Accept(visit Visitor) {
	visit.VisitCast(c)
}

type Type struct {
	token.Position

	Name      string
	Length    int
	Precision int
}

type Not struct {
	token.Position
	Node
}

func (n Not) Accept(visit Visitor) {
	visit.VisitNot(n)
}

func (n Not) GetStatement() []Node {
	return slx.One(n.Node)
}

type Collate struct {
	token.Position
	Ident Node
	Value Node
}

func (_ Collate) Accept(visit Visitor) {}

type Exists struct {
	token.Position
	Node
}

func (e Exists) Accept(visit Visitor) {
	visit.VisitExists(e)
}

func (e Exists) GetStatement() []Node {
	return slx.One(e.Node)
}

type Call struct {
	Position token.Position
	Distinct bool
	Ident    Node
	Args     []Node
	Filter   Node
	Over     Node
}

func (c Call) Accept(visit Visitor) {
	visit.VisitCall(c)
}

func (c Call) GetStatement() []Node {
	return c.Args
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
	Values []Node
}

func (_ Row) Accept(visit Visitor) {}

func (r Row) GetStatement() []Node {
	return r.Values
}

func (r Row) Keyword() (string, error) {
	return "ROW", nil
}

type Unary struct {
	token.Position
	Right Node
	Op    string
}

func (u Unary) Accept(visit Visitor) {
	visit.VisitUnary(u)
}

func (u Unary) GetStatement() []Node {
	return slx.One(u.Right)
}

type Binary struct {
	token.Position
	Left  Node
	Right Node
	Op    string
}

func (b Binary) Accept(visit Visitor) {
	visit.VisitBinary(b)
}

func (b Binary) GetStatement() []Node {
	return slx.Make(b.Left, b.Right)
}

func (b Binary) IsRelation() bool {
	return b.Op == "AND" || b.Op == "OR"
}

type All struct {
	token.Position
	Node
}

func (a All) Accept(visit Visitor) {
	visit.VisitAll(a)
}

func (a All) GetStatement() []Node {
	return slx.One(a.Node)
}

type Any struct {
	token.Position
	Node
}

func (a Any) Accept(visit Visitor) {
	visit.VisitAny(a)
}

func (a Any) GetStatement() []Node {
	return slx.One(a.Node)
}

type Is struct {
	token.Position
	Ident Node
	Value Node
}

func (i Is) Accept(visit Visitor) {
	visit.VisitIs(i)
}

func (i Is) GetStatement() []Node {
	return slx.One(i.Value)
}

type In struct {
	token.Position
	Ident Node
	Value Node
}

func (i In) Accept(visit Visitor) {
	visit.VisitIn(i)
}

func (i In) GetStatement() []Node {
	return slx.One(i.Value)
}

type Between struct {
	token.Position
	Ident Node
	Lower Node
	Upper Node
}

func (b Between) Accept(visit Visitor) {
	visit.VisitBetween(b)
}

func (b Between) GetStatement() []Node {
	return slx.Make(b.Lower, b.Upper)
}

type List struct {
	Values []Node
}

func (i List) Accept(visit Visitor) {
	visit.VisitList(i)
}

func (i List) Len() int {
	return len(i.Values)
}

type Placeholder struct {
	token.Position
	Node
}

func (_ Placeholder) Accept(visit Visitor) {}

type Value struct {
	token.Position
	Literal string
}

func (v Value) Accept(visit Visitor) {
	visit.VisitValue(v)
}

func (v Value) Number() bool {
	_, err := strconv.ParseFloat(v.Literal, 64)
	return err == nil
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
	Node
	Identifier
	Columns []Node
}

func (a Alias) Accept(visit Visitor) {
	visit.VisitAlias(a)
}

type Identifier struct {
	Quoted bool
	Name   string
}

func (i Identifier) Star() bool {
	return !i.Quoted && i.Name == ""
}

type Name struct {
	token.Position
	Parts []Identifier
}

func (n Name) Accept(visit Visitor) {
	visit.VisitName(n)
}

func (n Name) All() bool {
	c := len(n.Parts)
	return c == 0 || n.Parts[c-1].Name == ""
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
