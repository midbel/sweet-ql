package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitInsert(stmt ast.InsertStatement) {
	w.Enter()
	defer w.Leave()

	w.WritePrefix()
	w.WriteKeyword("insert")
	w.WriteBlank()
	w.WriteKeyword("into")
	w.WriteBlank()
	stmt.Table.Accept(w)
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
}

func (w *Writer) VisitDelete(stmt ast.DeleteStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("delete")
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
}

func (w *Writer) FormatMerge(stmt ast.MergeStatement) error {
	kw, _ := stmt.Keyword()
	w.WriteKeyword(kw)
	w.WriteBlank()
	w.WriteKeyword("INTO")
	w.WriteBlank()
	if err := w.FormatExpr(stmt.Target, false); err != nil {
		return err
	}
	w.WriteBlank()
	w.WriteKeyword("USING")
	w.WriteBlank()
	if err := w.FormatExpr(stmt.Source, false); err != nil {
		return err
	}
	w.WriteBlank()
	w.WriteKeyword("ON")
	w.WriteBlank()
	if err := w.FormatExpr(stmt.Join, false); err != nil {
		return err
	}
	for _, a := range stmt.Actions {
		m, ok := a.(ast.MatchStatement)
		if !ok {
			return w.CanNotUse("merge", a)
		}
		w.WriteNL()
		if err := w.FormatMatch(m); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) FormatMatch(stmt ast.MatchStatement) error {
	w.WriteKeyword("WHEN")
	w.WriteBlank()
	switch stmt.Node.(type) {
	case ast.DeleteStatement:
		w.WriteKeyword("MATCHED")
	case ast.UpdateStatement:
		w.WriteKeyword("MATCHED")
	case ast.InsertStatement:
		w.WriteKeyword("NOT MATCHED")
	default:
		return w.CanNotUse("merge", stmt.Node)
	}
	if stmt.Condition != nil {
		w.WriteBlank()
		w.WriteKeyword("AND")
		w.WriteBlank()
		if err := w.FormatExpr(stmt.Condition, false); err != nil {
			return err
		}
	}
	w.WriteBlank()
	w.WriteKeyword("THEN")
	w.WriteNL()

	switch stmt := stmt.Node.(type) {
	case ast.DeleteStatement:
		w.WriteKeyword("DELETE")
	case ast.UpdateStatement:
		w.WriteKeyword("UPDATE")
		w.WriteBlank()
		w.WriteKeyword("SET")
		w.WriteBlank()

		compact := w.Compact
		w.Compact = GetCompactMode("")
		defer func() {
			w.Compact = compact
		}()
		if err := w.FormatAssignment(stmt.List); err != nil {
			return err
		}
	case ast.InsertStatement:
		w.WriteKeyword("INSERT")
		w.WriteBlank()
		if len(stmt.Columns) > 0 {
			w.WriteString("(")
			for i := range stmt.Columns {
				if i > 0 {
					w.WriteString(",")
					w.WriteBlank()
				}
			}
			w.WriteString(")")
			w.WriteBlank()
		}
		// values, ok := stmt.Values.(ast.ValuesStatement)
		// if !ok {
		// 	return w.CanNotUse("merge", stmt.Values)
		// }
		compact := w.Compact
		w.Compact = GetCompactMode("")
		defer func() {
			w.Compact = compact
		}()
		// if err := w.FormatValues(values); err != nil {
		// 	return err
		// }
	default:
		return w.CanNotUse("merge", stmt)
	}
	return nil
}

func (w *Writer) FormatUpdate(stmt ast.UpdateStatement) error {
	kw, _ := stmt.Keyword()
	w.WriteKeyword(kw)
	w.WriteBlank()

	w.Enter()
	defer w.Leave()

	switch stmt := stmt.Table.(type) {
	case ast.Name:
		w.FormatName(stmt)
	case ast.Alias:
		if err := w.FormatAlias(stmt); err != nil {
			return err
		}
	default:
		return w.CanNotUse("update", stmt)
	}
	// w.WriteBlank()
	w.WriteNL()
	w.WriteKeyword("SET")
	w.WriteNL()
	// w.WriteBlank()

	w.Enter()
	if err := w.FormatAssignment(stmt.List); err != nil {
		return err
	}
	w.Leave()

	if len(stmt.Tables) > 0 {
		w.WriteBlank()
		// if err := w.FormatFrom(stmt.Tables); err != nil {
		// 	return err
		// }
	}
	if stmt.Where != nil {
		// w.WriteNL()
		// if err := w.FormatWhere(stmt.Where); err != nil {
		// 	return err
		// }
	}
	return nil
}

func (w *Writer) FormatAssignment(list []ast.Node) error {
	var err error
	for i, s := range list {
		if i > 0 {
			w.WriteString(",")
			// if w.Compact.ValuesStacked() {
			// 	w.WriteNL()
			// } else {
			// 	w.WriteBlank()
			// }
		}
		ass, ok := s.(ast.Assignment)
		if !ok {
			return w.CanNotUse("assignment", s)
		}
		switch field := ass.Field.(type) {
		case ast.Name:
			w.WritePrefix()
			w.FormatName(field)
		case ast.List:
			// err = w.formatList(field, w.Compact.ColumnsStacked())
		default:
			return w.CanNotUse("assignment", s)
		}
		if err != nil {
			return err
		}
		if w.Compact.KeepSpacesAround() {
			w.WriteBlank()
		}
		w.WriteString("=")
		if w.Compact.KeepSpacesAround() {
			w.WriteBlank()
		}
		switch value := ass.Value.(type) {
		case ast.List:
			// err = w.formatList(value, w.Compact.ValuesStacked())
		default:
			err = w.FormatExpr(value, false)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func (w *Writer) FormatReturning(stmt ast.Node) error {
	if stmt == nil {
		return nil
	}
	w.WriteKeyword("RETURNING")
	w.WriteBlank()

	list, ok := stmt.(ast.List)
	if !ok {
		return w.FormatExpr(stmt, false)
	}
	for i, v := range list.Values {
		if err := w.FormatExpr(v, false); err != nil {
			return err
		}
		if i < len(list.Values)-1 {
			w.WriteString(",")
			w.WriteBlank()
		}
	}
	return nil
}
