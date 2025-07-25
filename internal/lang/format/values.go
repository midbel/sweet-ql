package format

import (
	"strconv"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitGroup(group ast.Group) {
	if _, ok := group.Node.(ast.SelectStatement); ok && w.Compact.Subqueries() {
		compact := w.Compact
		w.Compact = compactAll
		defer func() {
			w.Compact = compact
		}()
	}
	w.WriteString("(")
	if !w.Compact.All() {
		w.WriteNL()
	}
	group.Node.Accept(w)
	if !w.Compact.All() {
		w.WriteNL()
	}
	w.WritePrefix()
	w.WriteString(")")
}

func (w *Writer) VisitValue(value ast.Value) {
	if value.Constant() {
		if w.withColor() {
			w.WriteString(keywordColor)
		}
		w.WriteKeyword(value.Literal)
		if w.withColor() {
			w.WriteString(resetCode)
		}
		return
	}
	if value.Number() {
		if w.withColor() {
			w.WriteString(numberColor)
		}
		w.WriteString(value.Literal)
		if w.withColor() {
			w.WriteString(resetCode)
		}
		return
	}
	w.WriteQuoted(value.Literal)
}

func (w *Writer) VisitCollate(collate ast.Collate) {
	collate.Ident.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("collate")
	w.WriteBlank()
	collate.Value.Accept(w)
}

func (w *Writer) VisitCall(call ast.Call) {
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
}

func (w *Writer) VisitName(name ast.Name) {
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
}

func (w *Writer) VisitAlias(alias ast.Alias) {
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
}

func (w *Writer) VisitBinary(bin ast.Binary) {
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
}

func (w *Writer) VisitUnary(una ast.Unary) {
	w.WriteKeyword(una.Op)
	una.Right.Accept(w)
}

func (w *Writer) VisitList(list ast.List) {
	w.WriteString("(")
	for i, v := range list.Values {
		if i > 0 {
			w.WriteComma()
		}
		v.Accept(w)
	}
	w.WriteString(")")
}

func (w *Writer) VisitCase(cas ast.Case) {
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
}

func (w *Writer) VisitWhen(when ast.When) {
	w.WriteKeyword("when")
	w.WriteBlank()
	when.Cdt.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("then")
	w.WriteBlank()
	when.Body.Accept(w)
}

func (w *Writer) VisitCast(cast ast.Cast) {
	w.WriteKeyword("cast")
	w.WriteString("(")
	cast.Node.Accept(w)
	w.WriteBlank()
	w.WriteKeyword("as")
	w.WriteBlank()
	w.visitType(cast.Type)
	w.WriteString(")")
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

func (w *Writer) VisitExists(exists ast.Exists) {
	w.WriteKeyword("exists")
	w.WriteString("(")
	exists.Node.Accept(w)
	w.WriteString(")")
}

func (w *Writer) VisitBetween(between ast.Between) {
	w.visitBetween(between, false)
}

func (w *Writer) VisitAll(all ast.All) {
	w.WriteKeyword("all")
	w.WriteString("(")
	all.Node.Accept(w)
	w.WriteString(")")
}

func (w *Writer) VisitAny(any ast.Any) {
	w.WriteKeyword("any")
	w.WriteString("(")
	any.Node.Accept(w)
	w.WriteString(")")
}

func (w *Writer) VisitIs(is ast.Is) {
	w.visitIs(is, false)
}

func (w *Writer) VisitIn(in ast.In) {
	w.visitIn(in, false)
}

func (w *Writer) VisitNot(not ast.Not) {
	switch n := not.Node.(type) {
	case ast.Is:
		w.visitIs(n, true)
	case ast.In:
		w.visitIn(n, true)
	case ast.Between:
		w.visitBetween(n, true)
	default:
	}
}

func (w *Writer) visitIs(is ast.Is, not bool) {
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

func (w *Writer) visitIn(in ast.In, not bool) {
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

func (w *Writer) visitBetween(between ast.Between, not bool) {
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

func (w *Writer) FormatName(name ast.Name) error {
	return nil
}

func (w *Writer) FormatAlias(alias ast.Alias) error {
	return nil
}

func (w *Writer) FormatLiteral(literal string) {

}

func (w *Writer) FormatRow(stmt ast.Row, nl bool) error {
	kw, _ := stmt.Keyword()
	w.WriteKeyword(kw)
	w.WriteString("(")
	for i, v := range stmt.Values {
		if i > 0 {
			w.WriteString(",")
			w.WriteBlank()
		}
		if nl {
			w.WriteNL()
		}
		if err := w.FormatExpr(v, false); err != nil {
			return err
		}
	}
	if nl {
		w.WriteNL()
	}
	w.WriteString(")")
	return nil
}

func (w *Writer) FormatCast(stmt ast.Cast, _ bool) error {
	return nil
}

func (w *Writer) FormatType(dt ast.Type) error {
	return nil
}
