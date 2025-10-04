package lint

import (
	"github.com/midbel/sweet/internal/lang/ast"
)

type caseIdent struct {
	ast.Visitor
	*rule
}

type caseConstant struct {
	ast.Visitor
	*rule
}
