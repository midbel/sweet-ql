package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitBegin(begin *ast.Begin) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("begin")
	w.WriteNL()
	if err := begin.Node.Accept(w); err != nil {
		return err
	}
	w.WritePrefix()
	w.WriteKeyword("end")
	return nil
}

func (w *Writer) VisitIf(stmt *ast.If) error {
	w.Enter()
	defer w.Leave()
	w.visitIf(stmt, "if")
	w.WritePrefix()
	w.WriteKeyword("end")
	w.WriteBlank()
	w.WriteKeyword("if")
	return nil
}

func (w *Writer) visitIf(stmt *ast.If, kw string) {
	w.WritePrefix()
	w.WriteKeyword(kw)
	w.WriteBlank()
	stmt.Cdt.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("then")
	w.WriteNL()
	stmt.Csq.Accept(w)

	if stmt.Alt != nil {
		if s, ok := stmt.Alt.(*ast.If); ok {
			w.visitIf(s, "elseif")
		} else {
			w.WritePrefix()
			w.WriteKeyword("else")
			w.WriteNL()
			stmt.Alt.Accept(w)
		}
	}
}

func (w *Writer) VisitWhile(stmt *ast.While) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("while")
	w.WriteBlank()
	stmt.Cdt.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("do")
	w.WriteNL()

	stmt.Body.Accept(w)
	w.WriteKeyword("end")
	w.WriteBlank()
	w.WriteKeyword("while")
	return nil
}

func (w *Writer) VisitSet(stmt *ast.Set) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("set")
	w.WriteBlank()
	w.WriteString(stmt.Ident)
	w.WriteBlank()
	w.WriteString("=")
	w.WriteBlank()
	stmt.Expr.Accept(w)
	return nil
}

func (w *Writer) VisitDeclare(stmt *ast.Declare) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("declare")
	w.WriteBlank()
	w.WriteString(stmt.Ident)
	w.WriteBlank()
	w.visitType(stmt.Type)
	if stmt.Value != nil {
		w.WriteBlank()
		w.WriteKeyword("default")
		w.WriteBlank()
		stmt.Value.Accept(w)
	}
	return nil
}

func (w *Writer) VisitCall(stmt *ast.CallStatement) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("call")
	w.WriteBlank()
	stmt.Ident.Accept(w)
	w.WriteBlank()
	w.WriteString("(")

	for i, a := range stmt.Args {
		if i > 0 {
			w.WriteComma()
		}
		a.Accept(w)
	}
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitReturn(stmt *ast.Return) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("return")
	w.WriteBlank()
	for i, v := range stmt.Values {
		if i > 0 {
			w.WriteComma()
		}
		if err := v.Accept(w); err != nil {
			return err
		}
	}
	return nil
}
