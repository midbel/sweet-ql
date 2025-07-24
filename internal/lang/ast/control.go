package ast

import (
	"slices"

	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type Return struct {
	token.Position
	Node
}

func (_ Return) Accept(visit Visitor) {}

type While struct {
	token.Position
	Cdt  Node
	Body Node
}

func (_ While) Accept(visit Visitor) {}

type If struct {
	token.Position
	Cdt Node
	Csq Node
	Alt Node
}

func (_ If) Accept(visit Visitor) {}

type Declare struct {
	token.Position
	Ident string
	Type  Type
	Value Node
}

func (_ Declare) Accept(visit Visitor) {}

type Case struct {
	token.Position
	Cdt  Node
	Body []Node
	Else Node
}

func (c Case) Accept(visit Visitor) {
	visit.VisitCase(c)
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

func (w When) Accept(visit Visitor) {
	visit.VisitWhen(w)
}

func (w When) GetStatement() []Node {
	return slx.Make(w.Cdt, w.Body)
}

type Set struct {
	token.Position
	Ident string
	Expr  Node
}

func (_ Set) Accept(visit Visitor) {}
