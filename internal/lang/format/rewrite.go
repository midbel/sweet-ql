package format

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

func (w *Writer) Rewrite(stmt ast.Statement) (ast.Statement, error) {
	return stmt, nil
}

func (w *Writer) rewrite(stmt ast.Statement) (ast.Statement, error) {
	return stmt, nil
}
