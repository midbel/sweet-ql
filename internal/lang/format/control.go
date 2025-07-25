package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitIf(stmt ast.If) {
	w.Enter()
	defer w.Leave()
	w.visitIf(stmt, "if")
	w.WriteKeyword("end")
	w.WriteBlank()
	w.WriteKeyword("if")
}

func (w *Writer) visitIf(stmt ast.If, kw string) {
	w.WritePrefix()
	w.WriteKeyword(kw)
	w.WriteBlank()
	stmt.Cdt.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("then")
	w.WriteNL()
	stmt.Csq.Accept(w)

	if stmt.Alt != nil {
		if s, ok := stmt.Alt.(ast.If); ok {
			w.visitIf(s, "elseif")
		} else {
			w.WritePrefix()
			w.WriteKeyword("else")
			w.WriteNL()
			stmt.Alt.Accept(w)
		}
	}
}

func (w *Writer) VisitWhile(stmt ast.While) {
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
}

func (w *Writer) VisitSet(stmt ast.Set) {
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
}

func (w *Writer) VisitDeclare(stmt ast.Declare) {
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
}

func (w *Writer) FormatCall(stmt ast.CallStatement) error {
	kw, _ := stmt.Keyword()
	w.WriteKeyword(kw)
	w.WriteString("(")
	defer w.WriteString(")")

	w.WriteNL()
	for i, a := range stmt.Args {
		if i > 0 {
			w.WriteString(",")
			w.WriteNL()
		}
		if err := w.FormatExpr(a, false); err != nil {
			return err
		}
	}
	w.WriteNL()
	return nil
}
