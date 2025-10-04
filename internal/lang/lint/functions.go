package lint

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

type callFunc struct {
	ast.Visitor
	*rule
}
