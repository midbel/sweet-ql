package ast

import (
	"slices"

	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type Return struct {
	token.Position
	Values []Node
}

func (r Return) Pos() token.Position {
	return r.Position
}

func (r Return) Accept(visit Visitor) error {
	return visit.VisitReturn(r)
}

type While struct {
	token.Position
	Cdt  Node
	Body Node
}

func (w While) Pos() token.Position {
	return w.Position
}

func (w While) Accept(visit Visitor) error {
	return visit.VisitWhile(w)
}

type If struct {
	token.Position
	Cdt Node
	Csq Node
	Alt Node
}

func (i If) Pos() token.Position {
	return i.Position
}

func (i If) Accept(visit Visitor) error {
	return visit.VisitIf(i)
}

type Declare struct {
	token.Position
	Ident string
	Type  Type
	Value Node
}

func (d Declare) Pos() token.Position {
	return d.Position
}

func (d Declare) Accept(visit Visitor) error {
	return visit.VisitDeclare(d)
}

type Case struct {
	token.Position
	Cdt  Node
	Body []Node
	Else Node
}

func (c Case) Pos() token.Position {
	return c.Position
}

func (c Case) Accept(visit Visitor) error {
	return visit.VisitCase(c)
}

func (c Case) GetStatement() []Node {
	all := slx.One(c.Cdt)
	return slices.Concat(all, c.Body, slx.One(c.Else))
}

type When struct {
	token.Position
	Cdt  Node
	Body Node
}

func (w When) Pos() token.Position {
	return w.Position
}

func (w When) Accept(visit Visitor) error {
	return visit.VisitWhen(w)
}

func (w When) GetStatement() []Node {
	return slx.Make(w.Cdt, w.Body)
}

type Set struct {
	token.Position
	Ident string
	Expr  Node
}

func (s Set) Pos() token.Position {
	return s.Position
}

func (s Set) Accept(visit Visitor) error {
	return visit.VisitSet(s)
}
