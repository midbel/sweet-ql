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

type While struct {
	token.Position
	Cdt  Node
	Body Node
}

type If struct {
	token.Position
	Cdt Node
	Csq Node
	Alt Node
}

type Declare struct {
	token.Position
	Ident string
	Type  Type
	Value Node
}

type Case struct {
	token.Position
	Cdt  Node
	Body []Node
	Else Node
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

func (w When) GetStatement() []Node {
	return slx.Make(w.Cdt, w.Body)
}

type Set struct {
	token.Position
	Ident string
	Expr  Node
}

type CallStatement struct {
	token.Position
	Ident Node
	Names []string
	Args  []Node
}

func (_ CallStatement) Keyword() (string, error) {
	return "CALL", nil
}
