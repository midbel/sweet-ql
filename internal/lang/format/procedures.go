package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitCreateProcedure(stmt ast.CreateProcedureStatement) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("create")
	w.WriteBlank()
	w.WriteKeyword("procedure")
	w.WriteBlank()
	stmt.Name.Accept(w)
	w.WriteString("(")
	for i, p := range stmt.Parameters {
		if i > 0 {
			w.WriteComma()
		}
		w.visitParamter(p)
	}
	w.WriteString(")")
	w.WriteNL()
	if stmt.Language != "" {
		w.WriteKeyword("language")
		w.WriteBlank()
		w.WriteIdent(stmt.Language)
		w.WriteNL()
	}
	w.WriteKeyword("begin")
	w.WriteNL()
	stmt.Body.Accept(w)
	w.WriteKeyword("end")
	return nil
}

func (w *Writer) visitParamter(param ast.ProcedureParameter) {
	switch param.Mode {
	case ast.ModeIn:
		w.WriteKeyword("in")
	case ast.ModeOut:
		w.WriteKeyword("out")
	case ast.ModeInOut:
		w.WriteKeyword("inout")
	}
	if param.Mode != 0 {
		w.WriteBlank()
	}
	w.WriteIdent(param.Name)
	w.WriteBlank()
	w.visitType(param.Type)
	if param.Default != nil {
		w.WriteBlank()
		w.WriteKeyword("default")
		w.WriteBlank()
		param.Default.Accept(w)
	}
}
