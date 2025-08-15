package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitStartTransaction(stmt *ast.StartTransaction) error {
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
	return nil
}

func (w *Writer) VisitSetTransaction(stmt *ast.SetTransaction) error {
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
	return nil
}

func (w *Writer) VisitSavepoint(stmt *ast.Savepoint) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("savepoint")
	if stmt.Name != nil {
		w.WriteBlank()
		stmt.Name.Accept(w)
	}
	return nil
}

func (w *Writer) VisitReleaseSavepoint(stmt *ast.ReleaseSavepoint) error {
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
	return nil
}

func (w *Writer) VisitRollbackSavepoint(stmt *ast.RollbackSavepoint) error {
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
	return nil
}

func (w *Writer) VisitCommit(stmt *ast.Commit) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("commit")
	return nil
}

func (w *Writer) VisitRollback(stmt *ast.Rollback) error {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("rollback")
	return nil
}
