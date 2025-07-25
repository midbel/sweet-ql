package format

import (
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
			c.Accept(w)
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
	w.Enter()
	defer w.Leave()
	where.Accept(w)
}

func (w *Writer) VisitValues(stmt ast.ValuesStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("values")
	w.WriteBlank()

	w.Enter()
	defer w.Leave()
	for i, v := range stmt.List {
		if i > 0 {
			w.WriteNL()
			w.WritePrefix()
		}
		if i > 0 && w.PrependComma {
			w.WriteComma()
		}
		v.Accept(w)
		if i < len(stmt.List)-1 && !w.PrependComma {
			w.WriteComma()
		}
	}
}
