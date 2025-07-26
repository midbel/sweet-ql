package format

import "github.com/midbel/sweet/internal/lang/ast"

func (w *Writer) VisitXmlElement(elem ast.XmlElement) {
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

func (w *Writer) VisitXmlAttribute(elem ast.XmlAttribute) {
	elem.Value.Accept(w)
	if elem.Name != nil {
		w.WriteBlank()
		w.WriteKeyword("AS")
		w.WriteBlank()
		elem.Name.Accept(w)
	}
}

func (w *Writer) VisitXmlNamespace(elem ast.XmlNamespace) {
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
}

func (w *Writer) VisitXmlText(elem ast.XmlText) {
	w.WriteCall("xmltext")
	w.WriteString("(")
	elem.Text.Accept(w)
	w.WriteString(")")
}

func (w *Writer) VisitXmlComment(elem ast.XmlComment) {
	w.WriteCall("xmlcomment")
	w.WriteString("(")
	elem.Text.Accept(w)
	w.WriteString(")")
}

func (w *Writer) VisitXmlAgg(elem ast.XmlComment) {
	w.WriteCall("xmlagg")
	w.WriteString("(")
	elem.Body.Accept(w)
	w.WriteString(")")
}
