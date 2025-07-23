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

type Statement interface{}

type Node struct {
	Statement
	Before []string
	After  string
}

func (n Node) Get() Statement {
	if len(n.Before) == 0 && n.After == "" {
		return n.Statement
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

	Statement
	Dir   OrderDir
	Nulls string
}

type Join struct {
	token.Position

	Type  string
	Table Statement
	Where Statement
}

type WindowDefinition struct {
	Ident  Statement
	Window Statement
}

type Window struct {
	Ident      Statement
	Partitions []Statement
	Orders     []Statement
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
	Expr Statement
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

type CteStatement struct {
	token.Position

	Ident        string
	Materialized MaterializedMode
	Columns      []string
	Statement
}

type WithStatement struct {
	token.Position

	Recursive bool
	Queries   []Statement
	Statement
}

func (s WithStatement) Keyword() (string, error) {
	return "WITH", nil
}

func (s WithStatement) Get() Statement {
	if len(s.Queries) == 0 {
		return s.Statement
	}
	return s
}

type ValuesStatement struct {
	token.Position

	List   []Statement
	Orders []Statement
	Limit  Statement
}

func (s ValuesStatement) Keyword() (string, error) {
	return "VALUES", nil
}

type SelectStatement struct {
	token.Position

	Distinct bool
	Columns  []Statement
	Tables   []Statement
	Where    Statement
	Groups   []Statement
	Having   Statement
	Windows  []Statement
	Orders   []Statement
	Limit    Statement
}

func (s SelectStatement) ColumnsCount() int {
	return -1
}

func (s SelectStatement) Keyword() (string, error) {
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

type UnionStatement struct {
	token.Position

	Left     Statement
	Right    Statement
	All      bool
	Distinct bool
}

func (s UnionStatement) GetStatement() []Statement {
	return slx.Make(s.Left, s.Right)
}

func (s UnionStatement) Keyword() (string, error) {
	return getCompoundKeyword("UNION", s.All, s.Distinct)
}

type IntersectStatement struct {
	token.Position

	Left     Statement
	Right    Statement
	All      bool
	Distinct bool
}

func (s IntersectStatement) GetStatement() []Statement {
	return slx.Make(s.Left, s.Right)
}

func (s IntersectStatement) Keyword() (string, error) {
	return getCompoundKeyword("INTERSECT", s.All, s.Distinct)
}

type ExceptStatement struct {
	token.Position

	Left     Statement
	Right    Statement
	All      bool
	Distinct bool
}

func (s ExceptStatement) GetStatement() []Statement {
	return slx.Make(s.Left, s.Right)
}

func (s ExceptStatement) Keyword() (string, error) {
	return getCompoundKeyword("EXCEPT", s.All, s.Distinct)
}

type MatchStatement struct {
	token.Position

	Condition Statement
	Statement
}

type MergeStatement struct {
	token.Position

	Target  Statement
	Source  Statement
	Join    Statement
	Actions []Statement
}

func (s MergeStatement) Keyword() (string, error) {
	return "MERGE", nil
}

type Upsert struct {
	Columns []string
	List    []Statement
	Where   Statement
}

type Assignment struct {
	Field Statement
	Value Statement
}

type InsertStatement struct {
	token.Position

	Table   Statement
	Columns []string
	Values  Statement
	Upsert  Statement
	Return  Statement
}

func (s InsertStatement) Keyword() (string, error) {
	return "INSERT INTO", nil
}

type UpdateStatement struct {
	token.Position

	Table  Statement
	List   []Statement
	Tables []Statement
	Where  Statement
	Return Statement
}

func (s UpdateStatement) Keyword() (string, error) {
	return "UPDATE", nil
}

type TruncateStatement struct {
	Tables   []string
	Cascade  CascadeMode
	Identity IdentityMode
}

func (s TruncateStatement) Keyword() (string, error) {
	return "TRUNCATE", nil
}

type DeleteStatement struct {
	token.Position

	Table  string
	Where  Statement
	Return Statement
}

func (s DeleteStatement) Keyword() (string, error) {
	return "DELETE FROM", nil
}
