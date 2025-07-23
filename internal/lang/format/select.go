package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/midbel/sweet/internal/lang"
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitSelect(stmt ast.SelectStatement) {
	w.Enter()
	defer w.Leave()

	kw, _ := stmt.Keyword()
	w.WritePrefix()
	w.WriteKeyword(kw)
	w.WriteNL()

	w.visitSelectColumns(stmt)
	w.visitSelectFrom(stmt)
	w.visitWhere(stmt.Where)
	if len(stmt.Groups) > 0 {

	}
	if stmt.Having != nil {

	}
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
			w.WriteString(",")
		}
		c.Accept(w)
		if i < len(stmt.Columns)-1 && !w.PrependComma {
			w.WriteString(",")
		}
	}
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

func (w *Writer) FormatUnion(stmt ast.UnionStatement) error {
	if err := w.FormatStatement(stmt.Left); err != nil {
		return err
	}
	w.WriteNL()
	w.Enter()
	w.WritePrefix()
	w.WriteKeyword("UNION")
	if stmt.All {
		w.WriteBlank()
		w.WriteKeyword("ALL")
	}
	if stmt.Distinct {
		w.WriteBlank()
		w.WriteKeyword("DISTINCT")
	}
	w.WriteNL()
	w.Leave()
	return w.FormatStatement(stmt.Right)
}

func (w *Writer) FormatExcept(stmt ast.ExceptStatement) error {
	if err := w.FormatStatement(stmt.Left); err != nil {
		return err
	}
	w.WriteNL()
	w.Enter()
	w.WritePrefix()
	w.WriteKeyword("EXCEPT")
	if stmt.All {
		w.WriteBlank()
		w.WriteKeyword("ALL")
	}
	if stmt.Distinct {
		w.WriteBlank()
		w.WriteKeyword("DISTINCT")
	}
	w.WriteNL()
	w.Leave()
	return w.FormatStatement(stmt.Right)
}

func (w *Writer) FormatIntersect(stmt ast.IntersectStatement) error {
	if err := w.FormatStatement(stmt.Left); err != nil {
		return err
	}
	w.WriteNL()
	w.Enter()
	w.WritePrefix()
	w.WriteKeyword("INTERSECT")
	if stmt.All {
		w.WriteBlank()
		w.WriteKeyword("ALL")
	}
	if stmt.Distinct {
		w.WriteBlank()
		w.WriteKeyword("DISTINCT")
	}
	w.WriteNL()
	w.Leave()
	return w.FormatStatement(stmt.Right)
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
	if len(stmt.Groups) > 0 {
		w.WriteNL()
		w.WritePrefix()
		if err := w.FormatGroupBy(stmt.Groups); err != nil {
			return err
		}
	}
	if stmt.Having != nil {
		w.WriteNL()
		if err := w.FormatHaving(stmt.Having); err != nil {
			return err
		}
	}
	if len(stmt.Windows) > 0 {
		w.WriteNL()
		w.WritePrefix()
		if err := w.FormatWindows(stmt.Windows); err != nil {
			return err
		}
	}
	if len(stmt.Orders) > 0 {
		w.WriteNL()
		w.WritePrefix()
		if err := w.FormatOrderBy(stmt.Orders); err != nil {
			return err
		}
	}
	if stmt.Limit != nil {
		w.WriteNL()
		w.WritePrefix()
		if err := w.FormatLimit(stmt.Limit); err != nil {
			return nil
		}
	}
	return nil
}

func (w *Writer) FormatWhere(stmt ast.Node) error {
	return nil
}

func (w *Writer) formatJoin(join ast.Join) error {
	return nil
}

func (w *Writer) FormatGroupBy(groups []ast.Node) error {
	if len(groups) == 0 {
		return nil
	}
	w.WriteKeyword("GROUP BY")

	w.Enter()
	defer w.Leave()

	for i := range groups {
		w.WriteNL()
		w.writeCommentBefore(groups[i])
		w.WritePrefix()
		if err := w.FormatExpr(groups[i], false); err != nil {
			return err
		}
		if i < len(groups)-1 {
			w.WriteString(",")
		}
		w.writeCommentAfter(groups[i])
	}
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

func (w *Writer) FormatHaving(having ast.Node) error {
	if having == nil {
		return nil
	}
	w.WriteKeyword("HAVING")
	w.WriteBlank()
	return w.FormatExpr(having, true)
}

func (w *Writer) FormatOrderBy(orders []ast.Node) error {
	if len(orders) == 0 {
		return nil
	}
	w.Enter()
	defer w.Leave()

	w.WriteKeyword("ORDER BY")
	for i := range orders {
		w.WriteNL()
		w.writeCommentBefore(orders[i])

		w.WritePrefix()
		if err := w.FormatExpr(orders[i], false); err != nil {
			return err
		}
		if i < len(orders)-1 {
			w.WriteString(",")
		}
		w.writeCommentAfter(orders[i])
	}
	return nil
}

func (w *Writer) formatOrder(order ast.Order) error {
	n, ok := order.Node.(ast.Name)
	if !ok {
		return w.CanNotUse("order by", order.Node)
	}
	w.FormatName(n)
	switch order.Dir {
	case 0:
	case ast.AscOrder:
		w.WriteBlank()
		w.WriteKeyword("ASC")
	case ast.DescOrder:
		w.WriteBlank()
		w.WriteKeyword("DESC")
	default:
		return fmt.Errorf("invalid order direction")
	}
	if order.Nulls != "" {
		w.WriteBlank()
		w.WriteKeyword("NULLS")
		w.WriteBlank()
		w.WriteKeyword(order.Nulls)
	}
	return nil
}

func (w *Writer) FormatLimit(stmt ast.Node) error {
	if stmt == nil {
		return nil
	}
	var limit ast.Node
	if n, ok := stmt.(ast.CommentedNode); ok {
		limit = n.Node
	} else {
		limit = stmt
	}
	lim, ok := limit.(ast.Limit)
	if !ok {
		return w.FormatOffset(stmt)
	}
	w.writeCommentBefore(stmt)
	w.WriteKeyword("LIMIT")
	w.WriteBlank()
	w.WriteString(strconv.Itoa(lim.Count))
	if lim.Offset > 0 {
		w.WriteNL()
		w.WritePrefix()
		w.WriteKeyword("OFFSET")
		w.WriteBlank()
		w.WriteString(strconv.Itoa(lim.Offset))
	}
	w.writeCommentAfter(stmt)
	return nil
}

func (w *Writer) FormatOffset(limit ast.Node) error {
	lim, ok := limit.(ast.Offset)
	if !ok {
		return w.CanNotUse("fetch", limit)
	}
	if lim.Offset > 0 {
		w.WriteKeyword("OFFSET")
		w.WriteBlank()
		w.WriteString(strconv.Itoa(lim.Offset))
		w.WriteBlank()
		w.WriteKeyword("ROWS")
		w.WriteBlank()
	}
	w.WriteKeyword("FETCH")
	w.WriteBlank()
	if lim.Next {
		w.WriteKeyword("NEXT")
	} else {
		w.WriteKeyword("FIRST")
	}
	w.WriteBlank()
	w.WriteString(strconv.Itoa(lim.Count))
	w.WriteBlank()
	w.WriteKeyword("ROWS ONLY")
	return nil
}

func (w *Writer) FormatWith(stmt ast.WithStatement) error {
	kw, _ := stmt.Keyword()
	w.WriteKeyword(kw)
	if stmt.Recursive {
		w.WriteBlank()
		w.WriteString("RECURSIVE")
	}
	w.WriteNL()

	for i, q := range stmt.Queries {
		if i > 0 {
			w.WriteNL()
		}
		w.writeCommentBefore(stmt.Queries[i])
		if err := w.FormatStatement(q); err != nil {
			return err
		}
		if i < len(stmt.Queries)-1 {
			w.WriteString(",")
		}
		w.writeCommentAfter(stmt.Queries[i])
	}
	w.WriteNL()
	return w.FormatStatement(stmt.Node)
}

func (w *Writer) FormatCte(stmt ast.CteStatement) error {
	ident := stmt.Ident
	if w.Upperize.Identifier() {
		ident = strings.ToUpper(ident)
	}
	w.WriteString(ident)
	if !w.Compact.Cte() && len(stmt.Columns) > 0 {
		w.WriteString("(")
		for i, s := range stmt.Columns {
			if i > 0 {
				w.WriteString(",")
				if !w.Compact.All() {
					w.WriteBlank()
				}
			}
			if w.Upperize.Identifier() {
				s = strings.ToUpper(s)
			}
			if w.UseQuote {
				s = w.Quote(s)
			}
			w.WriteString(s)
		}
		w.WriteString(")")
	}
	w.WriteBlank()
	w.WriteKeyword("AS")
	w.WriteBlank()
	w.WriteString("(")
	if !w.Compact.All() {
		w.WriteNL()
	}

	w.Enter()
	defer w.Leave()
	if err := w.FormatStatement(stmt.Node); err != nil {
		return err
	}
	if !w.Compact.All() {
		w.WriteNL()
	}
	w.WriteString(")")
	return nil
}
