package format

import "github.com/midbel/sweet/internal/lang/ast"

func (w *Writer) VisitXmlElement(elem ast.XmlElement) error {
	w.WriteCall("xmlelement")
	w.WriteString("(")
	w.WriteNL()

	w.Enter()

	w.WritePrefix()
	w.WriteKeyword("name")
	w.WriteBlank()
	elem.Name.Accept(w)
	w.visitXmlNamespaces(elem)
	w.visitXmlAttributes(elem)
	if len(elem.Children) > 0 {
		w.WriteString(",")
		for i, e := range elem.Children {
			if i > 0 {
				w.WriteComma()
			}
			w.WriteNL()
			w.WritePrefix()
			e.Accept(w)
		}
	}
	w.WriteNL()
	w.Leave()
	w.WritePrefix()
	w.WriteString(")")
	return nil
}

func (w *Writer) visitXmlAttributes(elem ast.XmlElement) {
	if len(elem.Attributes) == 0 {
		return
	}
	w.WriteString(",")
	w.WriteNL()
	w.WritePrefix()
	w.WriteCall("xmlattributes")
	w.WriteString("(")
	w.Enter()
	for i, a := range elem.Attributes {
		w.WriteNL()
		w.WritePrefix()
		a.Accept(w)
		if i < len(elem.Attributes)-1 {
			w.WriteString(",")
		}
	}
	w.Leave()
	w.WriteNL()
	w.WritePrefix()
	w.WriteString(")")
}

func (w *Writer) visitXmlNamespaces(elem ast.XmlElement) {
	if len(elem.Namespaces) == 0 {
		return
	}
	w.WriteString(",")
	w.WriteNL()
	w.WritePrefix()
	w.WriteCall("xmlnamespaces")
	w.WriteString("(")
	w.Enter()
	for i, a := range elem.Namespaces {
		w.WriteNL()
		w.WritePrefix()
		a.Accept(w)
		if i < len(elem.Namespaces)-1 {
			w.WriteString(",")
		}
	}
	w.Leave()
	w.WriteNL()
	w.WritePrefix()
	w.WriteString(")")
}

func (w *Writer) VisitXmlAttribute(elem ast.XmlAttribute) error {
	elem.Value.Accept(w)
	if elem.Name != nil {
		w.WriteBlank()
		w.WriteKeyword("AS")
		w.WriteBlank()
		elem.Name.Accept(w)
	}
	return nil
}

func (w *Writer) VisitXmlNamespace(elem ast.XmlNamespace) error {
	if elem.Name == nil {
		w.WriteKeyword("DEFAULT")
		w.WriteBlank()
	}
	elem.Uri.Accept(w)
	if elem.Name != nil {
		w.WriteBlank()
		w.WriteKeyword("AS")
		w.WriteBlank()
		elem.Name.Accept(w)
	}
	return nil
}

func (w *Writer) VisitXmlPi(elem ast.XmlPi) error {
	w.WriteCall("xmlpi")
	w.WriteString("(")
	w.WriteNL()
	w.Enter()
	w.WritePrefix()
	w.WriteKeyword("name")
	w.WriteBlank()
	elem.Name.Accept(w)
	w.WriteNL()
	w.Leave()
	w.WritePrefix()
	w.WriteString(")")
	return nil
	return nil
}

func (w *Writer) VisitXmlRoot(elem ast.XmlRoot) error {
	w.WriteCall("xmlroot")
	w.WriteString("(")
	w.WriteNL()
	w.Enter()
	w.WritePrefix()
	elem.Root.Accept(w)
	if elem.Version != "" {
		w.WriteComma()
		w.WriteNL()
		w.WritePrefix()
		w.WriteKeyword("version")
		w.WriteBlank()
		if elem.Version == "NO VALUE" {
			w.WriteKeyword(elem.Version)
		} else {
			w.WriteQuoted(elem.Version)

		}
	}
	if elem.Standalone != "" {
		w.WriteComma()
		w.WriteNL()
		w.WritePrefix()
		w.WriteKeyword("standalone")
		w.WriteBlank()
		w.WriteKeyword(elem.Standalone)
	}
	w.WriteNL()
	w.Leave()
	w.WritePrefix()
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitXmlText(elem ast.XmlText) error {
	w.WriteCall("xmltext")
	w.WriteString("(")
	elem.Text.Accept(w)
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitXmlComment(elem ast.XmlComment) error {
	w.WriteCall("xmlcomment")
	w.WriteString("(")
	elem.Text.Accept(w)
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitXmlAgg(elem ast.XmlAgg) error {
	w.WriteCall("xmlagg")
	w.WriteString("(")
	w.WriteNL()
	w.Enter()
	w.WritePrefix()
	elem.Body.Accept(w)
	w.WriteNL()
	w.Leave()
	w.WritePrefix()
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitXmlConcat(elem ast.XmlConcat) error {
	w.WriteCall("xmlconcat")
	w.WriteString("(")
	w.Enter()
	for i, n := range elem.Args {
		if i > 0 {
			w.WriteComma()
		}
		w.WriteNL()
		w.WritePrefix()
		n.Accept(w)
	}
	w.Leave()
	w.WriteNL()
	w.WritePrefix()
	w.WriteString(")")
	return nil
}

func (w *Writer) VisitXmlForest(elem ast.XmlForest) error {
	w.WriteCall("xmlforest")
	w.WriteString("(")
	w.Enter()
	for j, n := range elem.Args {
		if j > 0 {
			w.WriteComma()
		}
		i, ok := n.(ast.XmlForestItem)
		if !ok {
			continue
		}
		w.WriteNL()
		w.WritePrefix()
		if i.Name != nil {
			w.WriteKeyword("element")
			w.WriteBlank()
			w.WriteKeyword("name")
			w.WriteBlank()
			i.Name.Accept(w)
			w.WriteBlank()
		}
		i.Node.Accept(w)
		switch i.OnNull {
		case ast.NullOnNull:
			w.WriteBlank()
			w.WriteKeyword("null")
			w.WriteBlank()
			w.WriteKeyword("on")
			w.WriteBlank()
			w.WriteKeyword("null")
		case ast.AbsentOnNull:
			w.WriteBlank()
			w.WriteKeyword("absent")
			w.WriteBlank()
			w.WriteKeyword("on")
			w.WriteBlank()
			w.WriteKeyword("null")
		default:
		}
	}
	w.Leave()
	w.WriteNL()
	w.WritePrefix()
	w.WriteString(")")
	return nil
}
