package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) VisitGrant(stmt ast.GrantStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("grant")
	w.WriteBlank()
	if len(stmt.Privileges) == 0 {
		w.WriteKeyword("all")
		w.WriteBlank()
		w.WriteKeyword("privileges")
	} else {
		for i, p := range stmt.Privileges {
			if i > 0 {
				w.WriteComma()
			}
			w.WriteKeyword(p)
		}
	}
	if stmt.Object != nil {
		w.WriteNL()
		w.WriteKeyword("on")
		w.WriteBlank()
		stmt.Object.Accept(w)
	}
	w.WriteNL()
	w.WriteKeyword("to")
	w.WriteBlank()
	for i, u := range stmt.Users {
		if i > 0 {
			w.WriteString(",")
			w.WriteBlank()
		}
		w.WriteIdent(u)
	}
	if stmt.Grant {
		w.WriteNL()
		w.WriteKeyword("with")
		w.WriteBlank()
		w.WriteKeyword("grant")
		w.WriteBlank()
		w.WriteKeyword("option")
	}
}

func (w *Writer) VisitRevoke(stmt ast.RevokeStatement) {
	w.Enter()
	defer w.Leave()
	w.WritePrefix()
	w.WriteKeyword("revoke")
	w.WriteBlank()
	if len(stmt.Privileges) == 0 {
		w.WriteKeyword("all")
	} else {
		for i, p := range stmt.Privileges {
			if i > 0 {
				w.WriteComma()
			}
			w.WriteKeyword(p)
		}
	}
	if stmt.Object != nil {
		w.WriteNL()
		w.WriteKeyword("on")
		w.WriteBlank()
		stmt.Object.Accept(w)
	}
	w.WriteNL()
	w.WriteKeyword("from")
	w.WriteBlank()
	for i, u := range stmt.Users {
		if i > 0 {
			w.WriteComma()
		}
		w.WriteIdent(u)
	}
	if stmt.Cascade == ast.Cascade {
		w.WriteNL()
		w.WriteKeyword("cascade")
	} else if stmt.Cascade == ast.Restrict {
		w.WriteNL()
		w.WriteKeyword("restrict")
	}

}
