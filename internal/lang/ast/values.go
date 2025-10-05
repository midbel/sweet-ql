package ast

import (
	"strconv"
	"strings"

	"github.com/midbel/sweet/internal/token"
)

type Begin struct {
	token.Position
	Node
}

func (b *Begin) Pos() token.Position {
	return b.Position
}

func (b *Begin) Accept(visit Visitor) error {
	return visit.VisitBegin(b)
}

type Body struct {
	token.Position
	Values []Node
}

func (b *Body) Pos() token.Position {
	return b.Position
}

func (b *Body) Accept(visit Visitor) error {
	return visit.VisitBody(b)
}

type List struct {
	token.Position
	Values []Node
}

func (i *List) Pos() token.Position {
	return i.Position
}

func (i *List) Accept(visit Visitor) error {
	return visit.VisitList(i)
}

func (i *List) Transform(tr Transformer) (Node, error) {
	return tr.TransformList(i)
}

func (i *List) Len() int {
	return len(i.Values)
}

type Group struct {
	token.Position
	Node
}

func (g *Group) Pos() token.Position {
	return g.Position
}

func (g *Group) Accept(visit Visitor) error {
	return visit.VisitGroup(g)
}

func (g *Group) Transform(tr Transformer) (Node, error) {
	return tr.TransformGroup(g)
}

type Cast struct {
	token.Position

	Node
	Type Type
}

func (c *Cast) Pos() token.Position {
	return c.Position
}

func (c *Cast) Accept(visit Visitor) error {
	return visit.VisitCast(c)
}

func (c *Cast) Transform(tr Transformer) (Node, error) {
	return tr.TransformCast(c)
}

type Type struct {
	token.Position

	Name      string
	Length    int
	Precision int
}

func (t *Type) Pos() token.Position {
	return t.Position
}

type Not struct {
	token.Position
	Node
}

func (n *Not) Pos() token.Position {
	return n.Position
}

func (n *Not) Accept(visit Visitor) error {
	return visit.VisitNot(n)
}

func (n *Not) Transform(tr Transformer) (Node, error) {
	return tr.TransformNot(n)
}

type Collate struct {
	token.Position
	Ident Node
	Value Node
}

func (c *Collate) Pos() token.Position {
	return c.Position
}

func (_ *Collate) Accept(visit Visitor) error {
	return nil
}

func (c *Collate) Transform(tr Transformer) (Node, error) {
	return nil, nil
}

type Exists struct {
	token.Position
	Node
}

func (e *Exists) Pos() token.Position {
	return e.Position
}

func (e *Exists) Accept(visit Visitor) error {
	return visit.VisitExists(e)
}

func (e *Exists) Transform(tr Transformer) (Node, error) {
	return tr.TransformExists(e)
}

type Call struct {
	token.Position
	Distinct bool
	Ident    Node
	Args     []Node
	Filter   Node
	Over     Node
}

func (c *Call) Pos() token.Position {
	return c.Position
}

func (c *Call) Accept(visit Visitor) error {
	return visit.VisitCallFunc(c)
}

func (c *Call) Transform(tr Transformer) (Node, error) {
	return tr.TransformCallFunc(c)
}

func (c *Call) GetIdent() string {
	n, ok := c.Ident.(*Name)
	if !ok {
		return ""
	}
	return n.Ident()
}

type Row struct {
	token.Position
	Values []Node
}

func (r *Row) Pos() token.Position {
	return r.Position
}

func (r *Row) Accept(visit Visitor) error {
	return nil
}

func (r *Row) Transform(tr Transformer) (Node, error) {
	return nil, nil
}

type Unary struct {
	token.Position
	Right Node
	Op    string
}

func (u *Unary) Pos() token.Position {
	return u.Position
}

func (u *Unary) Accept(visit Visitor) error {
	return visit.VisitUnary(u)
}

func (u *Unary) Transform(tr Transformer) (Node, error) {
	return tr.TransformUnary(u)
}

type Binary struct {
	token.Position
	Left  Node
	Right Node
	Op    string
}

func (b *Binary) Pos() token.Position {
	return b.Position
}

func (b *Binary) Accept(visit Visitor) error {
	return visit.VisitBinary(b)
}

func (b *Binary) Transform(tr Transformer) (Node, error) {
	return tr.TransformBinary(b)
}

func (b *Binary) IsEquality() bool {
	return b.Op == "=" || b.Op == "!=" || b.Op == "<>"
}

func (b *Binary) IsRelation() bool {
	return b.Op == "AND" || b.Op == "OR"
}

type All struct {
	token.Position
	Node
}

func (a *All) Pos() token.Position {
	return a.Position
}

func (a *All) Accept(visit Visitor) error {
	return visit.VisitAll(a)
}

func (a *All) Transform(tr Transformer) (Node, error) {
	return tr.TransformAll(a)
}

type Any struct {
	token.Position
	Node
}

func (a *Any) Pos() token.Position {
	return a.Position
}

func (a *Any) Accept(visit Visitor) error {
	return visit.VisitAny(a)
}

func (a *Any) Transform(tr Transformer) (Node, error) {
	return tr.TransformAny(a)
}

type Is struct {
	token.Position
	Ident Node
	Value Node
}

func (i *Is) Pos() token.Position {
	return i.Position
}

func (i *Is) Accept(visit Visitor) error {
	return visit.VisitIs(i)
}

func (i *Is) Transform(tr Transformer) (Node, error) {
	return tr.TransformIs(i)
}

type In struct {
	token.Position
	Ident Node
	Value Node
}

func (i *In) Pos() token.Position {
	return i.Position
}

func (i *In) Accept(visit Visitor) error {
	return visit.VisitIn(i)
}

func (i *In) Transform(tr Transformer) (Node, error) {
	return tr.TransformIn(i)
}

type Between struct {
	token.Position
	Ident Node
	Lower Node
	Upper Node
}

func (b *Between) Pos() token.Position {
	return b.Position
}

func (b *Between) Accept(visit Visitor) error {
	return visit.VisitBetween(b)
}

func (b *Between) Transform(tr Transformer) (Node, error) {
	return tr.TransformBetween(b)
}

type Placeholder struct {
	token.Position
	Node
}

func (p *Placeholder) Pos() token.Position {
	return p.Position
}

func (p *Placeholder) Accept(visit Visitor) error {
	return visit.VisitPlaceholder(p)
}

func (p *Placeholder) Transform(tr Transformer) (Node, error) {
	return p, nil
}

type Value struct {
	token.Position
	Literal string
}

func (v *Value) Pos() token.Position {
	return v.Position
}

func (v *Value) Accept(visit Visitor) error {
	return visit.VisitValue(v)
}

func (v *Value) Transform(tr Transformer) (Node, error) {
	return tr.TransformValue(v)
}

func (v *Value) Type() StaticType {
	if v.Bool() {
		return TypeBool
	}
	if v.Null() {
		return TypeNull
	}
	if v.Number() {
		return TypeNumber
	}
	return TypeText
}

func (v *Value) Number() bool {
	_, err := strconv.ParseFloat(v.Literal, 64)
	return err == nil
}

func (v *Value) Constant() bool {
	return v.Null() || v.True() || v.False()
}

func (v *Value) Default() bool {
	return v.Literal == "DEFAULT"
}

func (v *Value) Bool() bool {
	return v.True() || v.False()
}

func (v *Value) Null() bool {
	return v.Literal == "NULL"
}

func (v *Value) True() bool {
	return v.Literal == "TRUE"
}

func (v *Value) False() bool {
	return v.Literal == "FALSE"
}

type Alias struct {
	token.Position
	Node
	Identifier
	Columns []Node
}

func (a *Alias) Pos() token.Position {
	return a.Position
}

func (a *Alias) Accept(visit Visitor) error {
	return visit.VisitAlias(a)
}

func (a *Alias) Transform(tr Transformer) (Node, error) {
	return tr.TransformAlias(a)
}

type Identifier struct {
	Quoted bool
	Name   string
}

func (i *Identifier) Star() bool {
	return !i.Quoted && i.Name == ""
}

type Name struct {
	token.Position
	Parts []Identifier
}

func (n *Name) Pos() token.Position {
	return n.Position
}

func (n *Name) Accept(visit Visitor) error {
	return visit.VisitName(n)
}

func (n *Name) Transform(tr Transformer) (Node, error) {
	return tr.TransformName(n)
}

func (n *Name) All() bool {
	c := len(n.Parts)
	return c == 0 || n.Parts[c-1].Star()
}

func (n *Name) Schema() string {
	switch len(n.Parts) {
	case 2:
		return n.Parts[0].Name
	case 3:
		return n.Parts[1].Name
	default:
		return ""
	}
}

func (n *Name) Name() string {
	if len(n.Parts) == 0 {
		return "*"
	}
	str := n.Parts[len(n.Parts)-1].Name
	if str == "" {
		str = "*"
	}
	return str
}

func (n *Name) Ident() string {
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
