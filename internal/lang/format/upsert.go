package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitInsert(stmt ast.InsertStatement) {
	w.Enter()
	defer w.Leave()

	w.WritePrefix()
	w.WriteKeyword("insert")
	if stmt.Table != nil {
		w.WriteBlank()
		w.WriteKeyword("into")
		w.WriteBlank()
		stmt.Table.Accept(w)
	}
	w.WriteBlank()
	if len(stmt.Columns) > 0 {
		w.WriteString("(")
		for i, c := range stmt.Columns {
			if i > 0 {
				w.WriteComma()
			}
			c.Accept(w)
		}
		w.WriteString(")")
	}
	w.WriteNL()
	stmt.Values.Accept(w)
}

func (w *Writer) VisitUpdate(stmt ast.UpdateStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("update")
	w.WriteBlank()
	if stmt.Table != nil {
		stmt.Table.Accept(w)
		w.WriteNL()
		w.Enter()
		defer w.Leave()
		w.WritePrefix()
	}
	w.WriteKeyword("set")
	w.WriteBlank()
	for i, n := range stmt.List {
		if i > 0 {
			w.WriteComma()
			w.WriteNL()
			w.WritePrefix()
		}
		n.Accept(w)
	}
	w.visitWhere(stmt.Where)
}

func (w *Writer) VisitAssignment(stmt ast.Assignment) {
	stmt.Field.Accept(w)
	w.WriteBlank()
	w.WriteString("=")
	w.WriteBlank()
	stmt.Value.Accept(w)
}

func (w *Writer) VisitDelete(stmt ast.DeleteStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("delete")
	if stmt.Table == nil {
		return
	}
	w.WriteBlank()
	w.WriteKeyword("from")
	w.WriteBlank()
	stmt.Table.Accept(w)
	w.visitWhere(stmt.Where)
}

func (w *Writer) VisitTruncate(stmt ast.TruncateStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("truncate")
	w.WriteBlank()
	w.WriteKeyword("table")
	w.WriteBlank()
	for i, n := range stmt.Tables {
		if i > 0 {
			w.WriteComma()
		}
		n.Accept(w)
	}
	if stmt.Identity == ast.RestartIdentity {
		w.WriteBlank()
		w.WriteKeyword("restart")
		w.WriteBlank()
		w.WriteKeyword("identity")
	} else if stmt.Identity == ast.ContinueIdentity {
		w.WriteBlank()
		w.WriteKeyword("continue")
		w.WriteBlank()
		w.WriteKeyword("identity")
	}
	if stmt.Cascade == ast.Cascade {
		w.WriteBlank()
		w.WriteKeyword("cascade")
	} else if stmt.Cascade == ast.Restrict {
		w.WriteBlank()
		w.WriteKeyword("restrict")
	}
}

func (w *Writer) VisitMerge(stmt ast.MergeStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("merge")
	w.WriteBlank()
	w.WriteKeyword("into")
	w.WriteBlank()
	stmt.Target.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("using")
	w.WriteBlank()
	stmt.Source.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("on")
	w.WriteBlank()
	stmt.Join.Accept(w)
	w.WriteNL()
	for i, a := range stmt.Actions {
		if i > 0 {
			w.WriteNL()
		}
		w.WritePrefix()
		a.Accept(w)
	}
}

func (w *Writer) VisitMatch(stmt ast.MatchStatement) {
	w.WriteKeyword("when")
	w.WriteBlank()
	switch stmt.Node.(type) {
	case ast.DeleteStatement, ast.UpdateStatement:
		w.WriteKeyword("matched")
	case ast.InsertStatement:
		w.WriteKeyword("not")
		w.WriteBlank()
		w.WriteKeyword("matched")
	default:
		return
	}
	if stmt.Condition != nil {
		w.WriteBlank()
		w.WriteKeyword("and")
		w.WriteBlank()
		stmt.Condition.Accept(w)
	}
	w.WriteBlank()
	w.WriteKeyword("then")
	w.WriteNL()
	stmt.Node.Accept(w)
}
