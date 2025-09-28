package format

import (
	"strconv"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitGroup(group *ast.Group) error {
	if _, ok := group.Node.(*ast.SelectStatement); ok && w.Compact.Subqueries() {
		compact := w.Compact
		w.Compact = compactAll
		defer func() {
			w.Compact = compact
		}()
	}
	w.WriteString("(")
	group.Node.Accept(w)
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitPlaceholder(placeholder *ast.Placeholder) error {
	if placeholder.Node == nil {
		w.WriteString("?")
		return nil
	}
	switch placeholder.Node.(type) {
	case *ast.Name:
		w.WriteString(":")
	case *ast.Value:
		w.WriteString("$")
	default:
	}
	return placeholder.Node.Accept(w)
}

func (w *Writer) VisitValue(value *ast.Value) error {
	if value.Constant() {
		if w.withColor() {
			w.WriteString(keywordColor)
		}
		w.WriteKeyword(value.Literal)
		if w.withColor() {
			w.WriteString(resetCode)
		}
		return nil
	}
	if value.Number() {
		if w.withColor() {
			w.WriteString(numberColor)
		}
		w.WriteString(value.Literal)
		if w.withColor() {
			w.WriteString(resetCode)
		}
		return nil
	}
	w.WriteQuoted(value.Literal)
	return nil
}

func (w *Writer) VisitCollate(collate *ast.Collate) error {
	collate.Ident.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("collate")
	w.WriteBlank()
	collate.Value.Accept(w)
	return nil
}

func (w *Writer) VisitCallFunc(call *ast.Call) error {
	call.Ident.Accept(w)
	w.WriteString("(")
	if call.Distinct {
		w.WriteKeyword("distinct")
		w.WriteBlank()
	}
	for i := range call.Args {
		if i > 0 {
			w.WriteComma()
		}
		call.Args[i].Accept(w)
	}
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitName(name *ast.Name) error {
	for i := range name.Parts {
		if i > 0 {
			w.WriteString(".")
		}
		str := name.Parts[i].Name
		if name.Parts[i].Star() {
			str = "*"
		}
		if w.Upperize.Identifier() || w.Upperize.All() {
			str = strings.ToUpper(str)
		}
		if name.Parts[i].Quoted || (w.UseQuote && !name.Parts[i].Star()) {
			str = w.Quote(str)
		}
		w.WriteString(str)
	}
	return nil
}

func (w *Writer) VisitAlias(alias *ast.Alias) error {
	alias.Node.Accept(w)
	w.WriteBlank()
	if !w.Compact.NoAs() {
		w.WriteKeyword("as")
		w.WriteBlank()
	}
	str := alias.Name
	if w.UseQuote || alias.Quoted {
		str = w.Quote(str)
	}
	w.WriteIdent(str)
	if len(alias.Columns) > 0 {
		w.WriteString("(")
		for i, c := range alias.Columns {
			if i > 0 {
				w.WriteComma()
			}
			c.Accept(w)
		}
		w.WriteString(")")
	}
	return nil
}

func (w *Writer) VisitBinary(bin *ast.Binary) error {
	bin.Left.Accept(w)
	if bin.IsRelation() {
		w.WriteNL()
		w.WritePrefix()
	} else if (w.Compact.KeepSpacesAround() && !bin.IsRelation()) || w.Compact.Expression() {
		w.WriteBlank()
	}
	w.WriteKeyword(bin.Op)
	if w.Compact.KeepSpacesAround() || bin.IsRelation() {
		w.WriteBlank()
	}
	bin.Right.Accept(w)
	return nil
}

func (w *Writer) VisitUnary(unary *ast.Unary) error {
	w.WriteKeyword(unary.Op)
	unary.Right.Accept(w)
	return nil
}

func (w *Writer) VisitList(list *ast.List) error {
	w.WriteString("(")
	for i, v := range list.Values {
		if i > 0 {
			w.WriteComma()
		}
		v.Accept(w)
	}
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitBody(body *ast.Body) error {
	for _, v := range body.Values {
		v.Accept(w)
		w.WriteEOL()
		w.WriteNL()
	}
	return nil
}

func (w *Writer) VisitCase(cas *ast.Case) error {
	w.WriteKeyword("case")
	if cas.Cdt != nil {
		w.WriteBlank()
		cas.Cdt.Accept(w)
	}
	if w.Compact.Expression() {
		w.WriteBlank()
	} else {
		w.WriteNL()
	}
	for _, n := range cas.Body {
		if !w.Compact.Expression() {
			w.WritePrefix()
		}
		n.Accept(w)
		if w.Compact.Expression() {
			w.WriteBlank()
		} else {
			w.WriteNL()
		}
	}
	if cas.Else != nil {
		if !w.Compact.Expression() {
			w.WritePrefix()
		}
		w.WriteKeyword("else")
		w.WriteBlank()
		cas.Else.Accept(w)
	}
	if w.Compact.Expression() {
		w.WriteBlank()
	} else {
		w.WriteNL()
		w.WritePrefix()
	}
	w.WriteKeyword("end")
	return nil
}

func (w *Writer) VisitWhen(when *ast.When) error {
	w.WriteKeyword("when")
	w.WriteBlank()
	when.Cdt.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("then")
	w.WriteBlank()
	when.Body.Accept(w)
	return nil
}

func (w *Writer) VisitCast(cast *ast.Cast) error {
	w.WriteKeyword("cast")
	w.WriteString("(")
	cast.Node.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("as")
	w.WriteBlank()
	w.visitType(cast.Type)
	w.WriteString(")")
	return nil
}

func (w *Writer) visitType(typ ast.Type) {
	if w.Upperize.All() || w.Upperize.Type() {
		typ.Name = strings.ToUpper(typ.Name)
	}
	w.WriteString(typ.Name)
	if typ.Length > 0 {
		w.WriteString("(")
		w.WriteString(strconv.Itoa(typ.Length))
		if typ.Precision > 0 {
			w.WriteComma()
			w.WriteString(strconv.Itoa(typ.Length))
		}
		w.WriteString(")")
	}
}

func (w *Writer) VisitExists(exists *ast.Exists) error {
	w.WriteKeyword("exists")
	w.WriteString("(")
	exists.Node.Accept(w)
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitBetween(between *ast.Between) error {
	w.visitBetween(between, false)
	return nil
}

func (w *Writer) VisitAll(all *ast.All) error {
	w.WriteKeyword("all")
	w.WriteString("(")
	all.Node.Accept(w)
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitAny(any *ast.Any) error {
	w.WriteKeyword("any")
	w.WriteString("(")
	any.Node.Accept(w)
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitIs(is *ast.Is) error {
	w.visitIs(is, false)
	return nil
}

func (w *Writer) VisitIn(in *ast.In) error {
	w.visitIn(in, false)
	return nil
}

func (w *Writer) VisitNot(not *ast.Not) error {
	switch n := not.Node.(type) {
	case *ast.Is:
		w.visitIs(n, true)
	case *ast.In:
		w.visitIn(n, true)
	case *ast.Between:
		w.visitBetween(n, true)
	default:
	}
	return nil
}

func (w *Writer) visitIs(is *ast.Is, not bool) {
	is.Ident.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("is")
	if not {
		w.WriteBlank()
		w.WriteKeyword("not")
	}
	w.WriteBlank()
	is.Value.Accept(w)
}

func (w *Writer) visitIn(in *ast.In, not bool) {
	in.Ident.Accept(w)
	if not {
		w.WriteBlank()
		w.WriteKeyword("not")
	}
	w.WriteBlank()
	w.WriteKeyword("in")
	w.WriteBlank()
	in.Value.Accept(w)
}

func (w *Writer) visitBetween(between *ast.Between, not bool) {
	between.Ident.Accept(w)
	if not {
		w.WriteBlank()
		w.WriteKeyword("not")
	}
	w.WriteBlank()
	w.WriteKeyword("between")
	w.WriteBlank()
	between.Lower.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("and")
	w.WriteBlank()
	between.Upper.Accept(w)
}

func (w *Writer) VisitRow(stmt *ast.Row) error {
	w.WriteKeyword("row")
	w.WriteString("(")
	for i, v := range stmt.Values {
		if i > 0 {
			w.WriteComma()
		}
		v.Accept(w)
	}
	w.WriteString(")")
	return nil
}
