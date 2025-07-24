package format

import (
	"fmt"
	"strings"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitUnion(stmt ast.UnionStatement) {
	stmt.Left.Accept(w)
	w.WriteNL()
	w.WriteKeyword("union")
	if stmt.All {
		w.WriteBlank()
		w.WriteKeyword("all")
	}
	if stmt.Distinct {
		w.WriteBlank()
		w.WriteKeyword("distinct")
	}
	w.WriteNL()
	stmt.Right.Accept(w)
}

func (w *Writer) VisitIntersect(stmt ast.IntersectStatement) {
	stmt.Left.Accept(w)
	w.WriteNL()
	w.WriteKeyword("intersect")
	if stmt.All {
		w.WriteBlank()
		w.WriteKeyword("all")
	}
	if stmt.Distinct {
		w.WriteBlank()
		w.WriteKeyword("distinct")
	}
	w.WriteNL()
	stmt.Right.Accept(w)
}

func (w *Writer) VisitExcept(stmt ast.ExceptStatement) {
	stmt.Left.Accept(w)
	w.WriteNL()
	w.WriteKeyword("except")
	if stmt.All {
		w.WriteBlank()
		w.WriteKeyword("all")
	}
	if stmt.Distinct {
		w.WriteBlank()
		w.WriteKeyword("distinct")
	}
	w.WriteNL()
	stmt.Right.Accept(w)
}

func (w *Writer) VisitWith(stmt ast.WithStatement) {
	w.Enter()

	w.WritePrefix()
	w.WriteKeyword("with")
	w.WriteBlank()
	for i, q := range stmt.Queries {
		if i > 0 {
			w.WriteComma()
		}
		q.Accept(w)
	}
	w.WriteNL()
	w.Leave()
	stmt.Node.Accept(w)
}

func (w *Writer) VisitCte(stmt ast.CteStatement) {
	if w.Upperize.Identifier() {
		stmt.Ident = strings.ToUpper(stmt.Ident)
	}
	w.WriteString(stmt.Ident)
	if len(stmt.Columns) > 0 {
		w.WriteBlank()
		w.WriteString("(")
		for i, c := range stmt.Columns {
			if i > 0 {
				w.WriteComma()
			}
			if w.Upperize.All() || w.Upperize.Identifier() {
				c = strings.ToUpper(c)
			}
			w.WriteString(c)
		}
		w.WriteString(")")
	}
	w.WriteBlank()
	w.WriteKeyword("as")
	w.WriteBlank()
	w.WriteString("(")
	w.WriteNL()
	stmt.Node.Accept(w)
	w.WriteNL()
	w.WriteString(")")
}

func (w *Writer) VisitSelect(stmt ast.SelectStatement) {
	w.Enter()
	defer w.Leave()

	w.WritePrefix()
	w.WriteKeyword("select")
	w.WriteNL()

	w.visitSelectColumns(stmt)
	w.visitSelectFrom(stmt)
	w.visitWhere(stmt.Where)
	w.visitSelectGroupBy(stmt)
	w.visitSelectHaving(stmt)
	w.visitSelectOrderBy(stmt)
	w.visitSelectLimit(stmt)
}

func (w *Writer) VisitJoin(join ast.Join) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	if w.Compact.Keyword() {
		join.Type = lang.CompactKeyword(join.Type)
	} else {
		join.Type = lang.ExpandKeyword(join.Type)
	}
	w.WriteKeyword(join.Type)
	w.WriteBlank()
	join.Table.Accept(w)
	w.WriteBlank()
	switch join.Where.(type) {
	case ast.Binary:
		w.WriteKeyword("on")
	case ast.List:
		w.WriteKeyword("using")
	default:
		return
	}
	w.WriteBlank()
	join.Where.Accept(w)
}

func (w *Writer) VisitOrder(order ast.Order) {
	order.Node.Accept(w)
	if order.Dir > 0 {
		w.WriteBlank()
	}
	switch order.Dir {
	case ast.AscOrder:
		w.WriteKeyword("asc")
	case ast.DescOrder:
		w.WriteKeyword("desc")
	default:
	}
}

func (w *Writer) VisitLimit(limit ast.Limit) {
	parts := []struct {
		Keyword string
		ast.Node
	}{
		{
			Keyword: "limit",
			Node:    limit.Count,
		},
		{
			Keyword: "offset",
			Node:    limit.Offset,
		},
	}
	for i, p := range parts {
		if i > 0 {
			w.WriteBlank()
		}
		w.WriteKeyword(p.Keyword)
		w.WriteBlank()
		p.Node.Accept(w)
	}
}

func (w *Writer) VisitOffset(offset ast.Offset) {
	if offset.Offset != nil {
		w.WriteKeyword("offset")
		w.WriteBlank()
		offset.Offset.Accept(w)
		w.WriteBlank()
		w.WriteKeyword("rows")
	}
	if offset.Count != nil {
		if offset.Offset != nil {
			w.WriteBlank()
		}
		w.WriteKeyword("fetch")
		w.WriteBlank()
		if offset.Next {
			w.WriteKeyword("next")
		} else {
			w.WriteKeyword("first")
		}
		w.WriteBlank()
		offset.Count.Accept(w)
		w.WriteBlank()
		w.WriteKeyword("rows")
		w.WriteBlank()
		w.WriteKeyword("only")
	}
}

func (w *Writer) visitSelectFrom(stmt ast.SelectStatement) {
	w.WriteNL()
	w.WritePrefix()
	w.WriteKeyword("from")
	w.WriteBlank()
	for i, t := range stmt.Tables {
		if i > 0 {
			w.WriteNL()
			w.WritePrefix()
		}
		t.Accept(w)
	}
}

func (w *Writer) visitSelectColumns(stmt ast.SelectStatement) {
	w.Enter()
	defer w.Leave()
	if stmt.Distinct {
		w.WritePrefix()
		w.WriteKeyword("distinct")
		w.WriteNL()
	}
	for i, c := range stmt.Columns {
		if i > 0 {
			w.WriteNL()
		}
		w.WritePrefix()
		if i > 0 && w.PrependComma {
			w.WriteComma()
		}
		c.Accept(w)
		if i < len(stmt.Columns)-1 && !w.PrependComma {
			w.WriteComma()
		}
	}
}

func (w *Writer) visitSelectGroupBy(stmt ast.SelectStatement) {
	if len(stmt.Groups) == 0 {
		return
	}
	w.WriteNL()
	w.WritePrefix()
	w.WriteKeyword("group by")
	w.WriteBlank()
	for i, g := range stmt.Groups {
		if i > 0 {
			w.WriteComma()
		}
		g.Accept(w)
	}
}

func (w *Writer) visitSelectHaving(stmt ast.SelectStatement) {
	if stmt.Having == nil {
		return
	}
	w.WriteNL()
	w.WritePrefix()
	w.WriteKeyword("having")
	w.WriteBlank()
	stmt.Having.Accept(w)
}

func (w *Writer) visitSelectOrderBy(stmt ast.SelectStatement) {
	if len(stmt.Orders) == 0 {
		return
	}
	w.WriteNL()
	w.WritePrefix()
	w.WriteKeyword("order by")
	w.WriteBlank()
	for i, g := range stmt.Orders {
		if i > 0 {
			w.WriteComma()
		}
		g.Accept(w)
	}
}

func (w *Writer) visitSelectLimit(stmt ast.SelectStatement) {
	if stmt.Limit == nil {
		return
	}
	w.WriteNL()
	w.WritePrefix()
	stmt.Limit.Accept(w)
}

func (w *Writer) visitWhere(where ast.Node) {
	if where == nil {
		return
	}
	w.WriteNL()
	w.WritePrefix()
	w.WriteKeyword("where")
	w.WriteBlank()
	where.Accept(w)
}

func (w *Writer) FormatValues(stmt ast.ValuesStatement) error {
	kw, _ := stmt.Keyword()
	w.WriteKeyword(kw)
	if len(stmt.List) > 1 && w.Compact.ValuesStacked() {
		w.WriteNL()
	} else {
		w.WriteBlank()
	}
	for i := range stmt.List {
		if i > 0 {
			w.WriteString(",")
			if w.Compact.ValuesStacked() {
				w.WriteNL()
			} else {
				w.WriteBlank()
			}
		}
		if len(stmt.List) > 1 && w.Compact.ValuesStacked() {
			w.WritePrefix()
		}
		if err := w.FormatExpr(stmt.List[i], false); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) FormatSelect(stmt ast.SelectStatement) error {
	return nil
}

func (w *Writer) FormatWhere(stmt ast.Node) error {
	return nil
}

func (w *Writer) formatJoin(join ast.Join) error {
	return nil
}

func (w *Writer) FormatGroupBy(groups []ast.Node) error {
	return nil
}

func (w *Writer) FormatWindows(windows []ast.Node) error {
	w.WriteKeyword("WINDOW")

	if len(windows) > 1 {
		w.WriteNL()
	} else {
		w.WriteBlank()
	}

	for i, c := range windows {
		def, ok := c.(ast.WindowDefinition)
		if !ok {
			return fmt.Errorf("window: unexpected statement type %T", c)
		}
		if i > 0 {
			w.WriteString(",")
			w.WriteNL()
		}
		if err := w.FormatExpr(def.Ident, false); err != nil {
			return err
		}
		w.WriteBlank()
		w.WriteKeyword("AS")
		w.WriteBlank()
		w.WriteString("(")
		win, ok := def.Window.(ast.Window)
		if !ok {
			return fmt.Errorf("window: unexpected statement type %T", def.Window)
		}
		if win.Ident != nil {
			if err := w.FormatExpr(win.Ident, false); err != nil {
				return err
			}
			w.WriteBlank()
		}
		if win.Ident == nil && len(win.Partitions) > 0 {
			w.WriteKeyword("PARTITION BY")
			w.WriteBlank()
			for i, p := range win.Partitions {
				if err := w.FormatExpr(p, false); err != nil {
					return err
				}
				if i < len(win.Partitions)-1 {
					w.WriteString(",")
					w.WriteBlank()
				}
			}
		}
		if len(win.Orders) > 0 {
			w.WriteBlank()
			w.WriteKeyword("ORDER BY")
			w.WriteBlank()
			for i, s := range win.Orders {
				if i > 0 {
					w.WriteString(",")
					w.WriteBlank()
				}
				order, ok := s.(ast.Order)
				if !ok {
					return w.CanNotUse("order by", s)
				}
				if err := w.formatOrder(order); err != nil {
					return err
				}
			}
		}
		w.WriteString(")")
	}
	return nil
}

func (w *Writer) formatOrder(order ast.Order) error {
	return nil
}
