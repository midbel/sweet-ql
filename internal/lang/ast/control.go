package ast

import (
	"slices"

	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

type Return struct {
	token.Position
	Statement
}

type While struct {
	token.Position
	Cdt  Statement
	Body Statement
}

type If struct {
	token.Position
	Cdt Statement
	Csq Statement
	Alt Statement
}

type Declare struct {
	token.Position
	Ident string
	Type  Type
	Value Statement
}

type Case struct {
	token.Position
	Cdt  Statement
	Body []Statement
	Else Statement
}

func (c Case) GetStatement() []Statement {
	all := slx.One(c.Cdt)
	return slices.Concat(all, c.Body, slx.One(c.Else))
}

type When struct {
	token.Position
	Cdt  Statement
	Body Statement
}

func (w When) GetStatement() []Statement {
	return slx.Make(w.Cdt, w.Body)
}

type Set struct {
	token.Position
	Ident string
	Expr  Statement
}

type CallStatement struct {
	token.Position
	Ident Statement
	Names []string
	Args  []Statement
}

func (_ CallStatement) Keyword() (string, error) {
	return "CALL", nil
}
