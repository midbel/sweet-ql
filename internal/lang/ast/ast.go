package ast

import (
	"fmt"

	"github.com/midbel/sweet/internal/slx"
	"github.com/midbel/sweet/internal/token"
)

// type Expr interface {
// 	IsExpr() bool
// }

// type Stmt interface {
// 	IsStmt() bool
// }

type Node interface{}

type CommentedNode struct {
	Node
	Before []string
	After  string
}

func (n CommentedNode) Get() Node {
	if len(n.Before) == 0 && n.After == "" {
		return n.Node
	}
	return n
}

type Limit struct {
	token.Position

	Count  int
	Offset int
}

type Offset struct {
	token.Position

	Limit
	Next bool
}

type OrderDir uint8

const (
	AscOrder OrderDir = 1 << iota
	DescOrder
)

type Order struct {
	token.Position

	Node
	Dir   OrderDir
	Nulls string
}

type Join struct {
	token.Position

	Type  string
	Table Node
	Where Node
}

type WindowDefinition struct {
	Ident  Node
	Window Node
}

type Window struct {
	Ident      Node
	Partitions []Node
	Orders     []Node
	Spec       FrameSpec
}

type FrameRow int

const (
	RowCurrent FrameRow = 1 << iota
	RowPreceding
	RowFollowing
	RowUnbounded
)

type FrameExclude int

const (
	ExcludeCurrent FrameExclude = 1 << (iota + 1)
	ExcludeNoOthers
	ExcludeGroup
	ExcludeTies
)

type FrameSpec struct {
	Row  FrameRow
	Expr Node
}

type BetweenFrameSpec struct {
	Left    FrameSpec
	Right   FrameSpec
	Exclude FrameExclude
}

type MaterializedMode int

const (
	MaterializedCte MaterializedMode = iota + 1
	NotMaterializedCte
)

type CteNode struct {
	token.Position

	Ident        string
	Materialized MaterializedMode
	Columns      []string
	Node
}

type WithNode struct {
	token.Position

	Recursive bool
	Queries   []Node
	Node
}

func (s WithNode) Keyword() (string, error) {
	return "WITH", nil
}

func (s WithNode) Get() Node {
	if len(s.Queries) == 0 {
		return s.Node
	}
	return s
}

type ValuesNode struct {
	token.Position

	List   []Node
	Orders []Node
	Limit  Node
}

func (s ValuesNode) Keyword() (string, error) {
	return "VALUES", nil
}

type SelectNode struct {
	token.Position

	Distinct bool
	Columns  []Node
	Tables   []Node
	Where    Node
	Groups   []Node
	Having   Node
	Windows  []Node
	Orders   []Node
	Limit    Node
}

func (s SelectNode) ColumnsCount() int {
	return -1
}

func (s SelectNode) Keyword() (string, error) {
	return "SELECT", nil
}

func getCompoundKeyword(kw string, all, distinct bool) (string, error) {
	var suffix string
	switch {
	default:
		return kw, nil
	case all:
		suffix = "ALL"
	case distinct:
		suffix = "DISTINCT"
	case all && distinct:
		return "", fmt.Errorf("%s: all and distinct can not be set at the same time", kw)
	}
	return fmt.Sprintf("%s %s", kw, suffix), nil
}

type UnionNode struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s UnionNode) GetNode() []Node {
	return slx.Make(s.Left, s.Right)
}

func (s UnionNode) Keyword() (string, error) {
	return getCompoundKeyword("UNION", s.All, s.Distinct)
}

type IntersectNode struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s IntersectNode) GetNode() []Node {
	return slx.Make(s.Left, s.Right)
}

func (s IntersectNode) Keyword() (string, error) {
	return getCompoundKeyword("INTERSECT", s.All, s.Distinct)
}

type ExceptNode struct {
	token.Position

	Left     Node
	Right    Node
	All      bool
	Distinct bool
}

func (s ExceptNode) GetNode() []Node {
	return slx.Make(s.Left, s.Right)
}

func (s ExceptNode) Keyword() (string, error) {
	return getCompoundKeyword("EXCEPT", s.All, s.Distinct)
}

type MatchNode struct {
	token.Position

	Condition Node
	Node
}

type MergeNode struct {
	token.Position

	Target  Node
	Source  Node
	Join    Node
	Actions []Node
}

func (s MergeNode) Keyword() (string, error) {
	return "MERGE", nil
}

type Upsert struct {
	Columns []string
	List    []Node
	Where   Node
}

type Assignment struct {
	Field Node
	Value Node
}

type InsertNode struct {
	token.Position

	Table   Node
	Columns []string
	Values  Node
	Upsert  Node
	Return  Node
}

func (s InsertNode) Keyword() (string, error) {
	return "INSERT INTO", nil
}

type UpdateNode struct {
	token.Position

	Table  Node
	List   []Node
	Tables []Node
	Where  Node
	Return Node
}

func (s UpdateNode) Keyword() (string, error) {
	return "UPDATE", nil
}

type TruncateNode struct {
	Tables   []string
	Cascade  CascadeMode
	Identity IdentityMode
}

func (s TruncateNode) Keyword() (string, error) {
	return "TRUNCATE", nil
}

type DeleteNode struct {
	token.Position

	Table  string
	Where  Node
	Return Node
}

func (s DeleteNode) Keyword() (string, error) {
	return "DELETE FROM", nil
}
