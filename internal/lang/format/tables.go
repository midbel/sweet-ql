package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitCreateTable(stmt ast.CreateTableStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("create")
	w.WriteBlank()
	w.WriteKeyword("table")
	w.WriteBlank()
	stmt.Name.Accept(w)
	w.WriteBlank()
	w.WriteString("(")
	w.WriteNL()
	for i, c := range stmt.Columns {
		w.Enter()
		w.WritePrefix()
		c.Accept(w)
		if i < len(stmt.Columns)-1 {
			w.WriteComma()
			w.WriteNL()
		}
		w.Leave()
	}
	if len(stmt.Constraints) > 0 {
		w.WriteComma()
		w.WriteNL()
	}
	for i, c := range stmt.Constraints {
		w.Enter()
		w.WritePrefix()
		c.Accept(w)
		if i < len(stmt.Constraints)-1 {
			w.WriteComma()
			w.WriteNL()
		}
		w.Leave()
	}
	w.WriteNL()
	w.WriteString(")")
}

func (w *Writer) VisitDropTable(stmt ast.DropTableStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("drop")
	w.WriteBlank()
	w.WriteKeyword("table")
	w.WriteBlank()
	for i, s := range stmt.Names {
		if i > 0 {
			w.WriteComma()
		}
		s.Accept(w)
	}
	if stmt.Cascade == ast.Cascade {
		w.WriteBlank()
		w.WriteKeyword("cascade")
	} else if stmt.Cascade == ast.Restrict {
		w.WriteBlank()
		w.WriteKeyword("restrict")
	}
}

func (w *Writer) VisitAlterTable(stmt ast.AlterTableStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("alter")
	w.WriteBlank()
	w.WriteKeyword("table")
	w.WriteBlank()
	stmt.Name.Accept(w)
	w.WriteBlank()
	stmt.Action.Accept(w)
}

func (w *Writer) VisitAddColumn(action AddColumnAction) {
	w.WriteKeyword("add")
	w.WriteBlank()
	w.WriteKeyword("column")
	w.WriteBlank()
	action.Def.Accept(w)
}

func (w *Writer) VisitAlterColumn(action AlterColumnAction) {

}

func (w *Writer) VisitDropColumn(action DropColumnAction) {
	w.WriteKeyword("drop")
	w.WriteBlank()
	w.WriteKeyword("column")
	w.WriteBlank()
	action.Name.Accept(w)
	if stmt.Cascade == ast.Cascade {
		w.WriteBlank()
		w.WriteKeyword("cascade")
	} else if stmt.Cascade == ast.Restrict {
		w.WriteBlank()
		w.WriteKeyword("restrict")
	}
}

func (w *Writer) VisitAddConstraint(action AddConstraintAction) {
	w.WriteKeyword("add")
	w.WriteBlank()
	w.WriteKeyword("constraint")
	w.WriteBlank()
	action.Constraint.Accept(w)
}

func (w *Writer) VisitDropConstraint(action DropConstraintAction) {
	w.WriteKeyword("drop")
	w.WriteBlank()
	w.WriteKeyword("constraint")
	w.WriteBlank()
	action.Name.Accept(w)
	if stmt.Cascade == ast.Cascade {
		w.WriteBlank()
		w.WriteKeyword("cascade")
	} else if stmt.Cascade == ast.Restrict {
		w.WriteBlank()
		w.WriteKeyword("restrict")
	}
}

func (w *Writer) VisitCreateView(stmt ast.CreateViewStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("create")
	w.WriteBlank()
	w.WriteKeyword("view")
	w.WriteBlank()
	stmt.Name.Accept(w)
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
	w.WriteBlank()
	w.WriteKeyword("as")
	w.WriteNL()
	stmt.Select.Accept(w)
}

func (w *Writer) VisitDropView(stmt ast.DropViewStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("drop")
	w.WriteBlank()
	w.WriteKeyword("view")
	w.WriteBlank()
	for i, s := range stmt.Names {
		if i > 0 {
			w.WriteComma()
		}
		s.Accept(w)
	}
	if stmt.Cascade == ast.Cascade {
		w.WriteBlank()
		w.WriteKeyword("cascade")
	} else if stmt.Cascade == ast.Restrict {
		w.WriteBlank()
		w.WriteKeyword("restrict")
	}
}

func (w *Writer) VisitColumnDef(def ast.ColumnDef) {
	def.Name.Accept(w)
	w.WriteBlank()
	w.visitType(def.Type)
	for _, c := range def.Constraints {
		w.WriteBlank()
		c.Accept(w)
	}
}

func (w *Writer) VisitConstraint(cst ast.Constraint) {
	if cst.Name != "" {
		w.WriteKeyword("constraint")
		w.WriteBlank()
		w.WriteIdent(cst.Name)
		w.WriteBlank()
	}
	cst.Node.Accept(w)
}

func (w *Writer) VisitPrimaryKey(cst ast.PrimaryKeyConstraint) {
	w.WriteKeyword("primary")
	w.WriteBlank()
	w.WriteKeyword("key")
	if len(cst.Columns) == 0 {
		return
	}
	w.WriteBlank()
	w.WriteString("(")
	for i, c := range cst.Columns {
		if i > 0 {
			w.WriteComma()
		}
		c.Accept(w)
	}
	w.WriteString(")")
}

func (w *Writer) VisitForeignKey(cst ast.ForeignKeyConstraint) {
	if len(cst.Locals) > 0 {
		w.WriteKeyword("foreign")
		w.WriteBlank()
		w.WriteKeyword("key")
		w.WriteBlank()
		w.WriteString("(")
		for i, c := range cst.Locals {
			if i > 0 {
				w.WriteComma()
			}
			c.Accept(w)
		}
		w.WriteString(")")
		w.WriteBlank()
	}
	if len(cst.Remotes) > 0 {
		w.WriteKeyword("references")
		w.WriteBlank()
		cst.Table.Accept(w)
		w.WriteString("(")
		for i, c := range cst.Remotes {
			if i > 0 {
				w.WriteComma()
			}
			c.Accept(w)
		}
		w.WriteString(")")
	}
}

func (w *Writer) VisitNotNull(cst ast.NotNullConstraint) {
	w.WriteKeyword("not")
	w.WriteBlank()
	w.WriteKeyword("null")
}

func (w *Writer) VisitUnique(cst ast.UniqueConstraint) {
	w.WriteKeyword("unique")
	if len(cst.Columns) == 0 {
		return
	}
	w.WriteString("(")
	for i, c := range cst.Columns {
		if i > 0 {
			w.WriteComma()
		}
		c.Accept(w)
	}
	w.WriteString(")")
}

func (w *Writer) VisitDefault(cst ast.DefaultConstraint) {
	compact := w.Compact
	w.Compact = compactAll
	defer func() {
		w.Compact = compact
	}()

	w.WriteKeyword("default")
	w.WriteBlank()
	cst.Expr.Accept(w)
}

func (w *Writer) VisitCheck(cst ast.CheckConstraint) {
	compact := w.Compact
	w.Compact = compactAll
	defer func() {
		w.Compact = compact
	}()
	w.WriteKeyword("check")
	w.WriteBlank()
	cst.Expr.Accept(w)
}

func (w *Writer) VisitGenerated(cst ast.GeneratedConstraint) {
}
