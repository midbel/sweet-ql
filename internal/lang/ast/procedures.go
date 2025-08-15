package ast

import (
	"github.com/midbel/sweet/internal/token"
)

type ProcedureParameter struct {
	token.Position
	Mode    ParameterMode
	Name    string
	Type    Type
	Default Node
}

func (p *ProcedureParameter) Pos() token.Position {
	return p.Position
}

type CreateProcedureStatement struct {
	Position   token.Position
	Name       Node
	Parameters []ProcedureParameter
	Language   string
	Body       Node
}

func (s *CreateProcedureStatement) Pos() token.Position {
	return s.Position
}

func (s *CreateProcedureStatement) Accept(visit Visitor) error {
	return visit.VisitCreateProcedure(s)
}
