package ast

type ProcedureParameter struct {
	Mode    ParameterMode
	Name    string
	Type    Type
	Default Node
}

type CreateProcedureStatement struct {
	Name       Node
	Parameters []ProcedureParameter
	Language   string
	Body       Node
}

func (s CreateProcedureStatement) Accept(visit Visitor) {
	visit.VisitCreateProcedure(s)
}
