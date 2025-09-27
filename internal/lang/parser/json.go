package parser

import (
	"fmt"
	"strings"

	"github.com/midbel/sweet/internal/lang/ast"
)

func (p *Parser) ParseJSON(name ast.Node) (ast.Node, error) {
	n, ok := name.(*ast.Name)
	if !ok {
		return nil, p.Unexpected("xml", defaultReason)
	}
	var (
		stmt ast.Node
		err  error
	)
	switch strings.ToUpper(n.Name()) {
	case "JSON_OBJECT":
	case "JSON_OBJECTAGG":
	case "JSON_ARRAY":
	case "JSON_ARRAYAGG":
	case "JSON_VALUE":
	case "JSON_QUERY":
	case "JSON_EXISTS":
	default:
		return nil, p.Unexpected("json", fmt.Sprintf("%s: undefined/unsupported xml functions"))
	}
	return stmt, err
}
