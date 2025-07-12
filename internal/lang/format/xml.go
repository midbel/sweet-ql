package format

import "github.com/midbel/sweet/internal/lang/ast"

func (w *Writer) FormatXmlElement(elem ast.XmlElement) error {
	n, ok := elem.Ident.(ast.Name)
	if !ok {
		return w.CanNotUse("xmlelement", elem.Ident)
	}
	w.WriteCall(n.Ident())
	w.WriteString("(")
	w.WriteNL()

	w.Enter()

	w.WritePrefix()
	w.WriteString("NAME")
	w.WriteBlank()
	if err := w.FormatExpr(elem.Name, false); err != nil {
		return err
	}
	if len(elem.Namespaces) > 0 {
		w.WriteString(",")
		w.WriteNL()
		w.WritePrefix()
		w.WriteCall("xmlnamespaces")
		w.WriteString("(")
		w.Enter()
		for i, a := range elem.Namespaces {
			x, ok := a.(ast.XmlNamespace)
			if !ok {
				return w.CanNotUse("xmlelement", a)
			}
			w.WriteNL()
			w.WritePrefix()
			if x.Name == nil {
				w.WriteKeyword("DEFAULT")
				w.WriteBlank()
			}
			if err := w.FormatExpr(x.Uri, false); err != nil {
				return err
			}
			if x.Name != nil {
				w.WriteBlank()
				w.WriteKeyword("AS")
				w.WriteBlank()
				if err := w.FormatExpr(x.Name, false); err != nil {
					return err
				}
			}
			if i < len(elem.Attributes)-1 {
				w.WriteString(",")
			}
		}
		w.Leave()
		w.WriteNL()
		w.WritePrefix()
		w.WriteString(")")
	}
	if len(elem.Attributes) > 0 {
		w.WriteString(",")
		w.WriteNL()
		w.WritePrefix()
		w.WriteCall("xmlattributes")
		w.WriteString("(")
		w.Enter()
		for i, a := range elem.Attributes {
			x, ok := a.(ast.XmlAttribute)
			if !ok {
				return w.CanNotUse("xmlelement", a)
			}
			w.WriteNL()
			w.WritePrefix()
			if err := w.FormatExpr(x.Value, false); err != nil {
				return err
			}
			if x.Name != nil {
				w.WriteBlank()
				w.WriteKeyword("AS")
				w.WriteBlank()
				if err := w.FormatExpr(x.Name, false); err != nil {
					return err
				}
			}
			if i < len(elem.Attributes)-1 {
				w.WriteString(",")
			}
		}
		w.Leave()
		w.WriteNL()
		w.WritePrefix()
		w.WriteString(")")
	}
	if len(elem.Children) > 0 {
		w.WriteString(",")
		for i, e := range elem.Children {
			w.WriteNL()
			w.WritePrefix()
			if err := w.FormatExpr(e, false); err != nil {
				return err
			}
			if i < len(elem.Children)-1 {
				w.WriteString(",")
			}
		}
	}
	w.WriteNL()
	w.Leave()
	w.WritePrefix()
	w.WriteString(")")
	return nil
}

func (w *Writer) FormatXmlText(elem ast.XmlText) error {
	n, ok := elem.Ident.(ast.Name)
	if !ok {
		return w.CanNotUse("xmltext", elem.Ident)
	}
	w.WriteCall(n.Ident())
	w.WriteString("(")
	if err := w.FormatExpr(elem.Text, false); err != nil {
		return err
	}
	w.WriteString(")")
	return nil
}

func (w *Writer) FormatXmlComment(elem ast.XmlComment) error {
	n, ok := elem.Ident.(ast.Name)
	if !ok {
		return w.CanNotUse("xmlcomment", elem.Ident)
	}
	w.WriteCall(n.Ident())
	w.WriteString("(")
	if err := w.FormatExpr(elem.Text, false); err != nil {
		return err
	}
	w.WriteString(")")
	return nil
}
