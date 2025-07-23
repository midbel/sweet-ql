package ast

type ParameterMode int

const (
	ModeIn ParameterMode = 1 << (iota + 1)
	ModeOut
	ModeInOut
)

type ProcedureParameter struct {
	Mode    ParameterMode
	Name    string
	Type    Type
	Default Node
}

func (_ ProcedureParameter) Accept(visit Visitor) {}

type CreateProcedureStatement struct {
	Replace    bool
	Name       Node
	Parameters []Node
	Language   string
	Body       Node
}

func (_ CreateProcedureStatement) Accept(visit Visitor) {}

func (s CreateProcedureStatement) Keyword() (string, error) {
	if s.Replace {
		return "CREATE OR REPLACE PROCEDURE", nil
	}
	return "CREATE PROCEDURE", nil
}
