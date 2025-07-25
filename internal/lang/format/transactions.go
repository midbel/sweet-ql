package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitStartTransaction(stmt ast.StartTransaction) {
	w.Enter()
	// defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("start")
	w.WriteBlank()
	w.WriteKeyword("transaction")
	if stmt.Mode > 0 {
		w.WriteBlank()
		switch stmt.Mode {
		case ast.ModeReadWrite:
			w.WriteKeyword("read")
			w.WriteBlank()
			w.WriteKeyword("write")
		case ast.ModeReadOnly:
			w.WriteKeyword("read")
			w.WriteBlank()
			w.WriteKeyword("only")
		default:
		}
	}
	if stmt.Body != nil {
		w.WriteNL()
		stmt.Body.Accept(w)
	}
	w.Leave()
	if stmt.End != nil {
		stmt.End.Accept(w)
	}
}

func (w *Writer) VisitSetTransaction(stmt ast.SetTransaction) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("set")
	w.WriteBlank()
	w.WriteKeyword("transaction")
	if stmt.Level > 0 {
		w.WriteBlank()
		w.WriteKeyword("isolation")
		w.WriteBlank()
		w.WriteKeyword("level")
	}
	if stmt.Mode > 0 {
		w.WriteBlank()
		switch stmt.Mode {
		case ast.ModeReadWrite:
			w.WriteKeyword("read")
			w.WriteBlank()
			w.WriteKeyword("write")
		case ast.ModeReadOnly:
			w.WriteKeyword("read")
			w.WriteBlank()
			w.WriteKeyword("only")
		default:
		}
	}
}

func (w *Writer) VisitSavepoint(stmt ast.Savepoint) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("savepoint")
	if stmt.Name != nil {
		w.WriteBlank()
		stmt.Name.Accept(w)
	}
}

func (w *Writer) VisitReleaseSavepoint(stmt ast.ReleaseSavepoint) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("release")
	w.WriteBlank()
	w.WriteKeyword("savepoint")
	if stmt.Name != nil {
		w.WriteBlank()
		stmt.Name.Accept(w)
	}
}

func (w *Writer) VisitRollbackSavepoint(stmt ast.RollbackSavepoint) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("rollback")
	w.WriteBlank()
	w.WriteKeyword("savepoint")
	if stmt.Name != nil {
		w.WriteBlank()
		stmt.Name.Accept(w)
	}
}

func (w *Writer) VisitCommit(stmt ast.Commit) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("commit")
}

func (w *Writer) VisitRollback(stmt ast.Rollback) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("rollback")
}
