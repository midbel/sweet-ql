package ast

import "strings"

type StaticType int16

const (
	TypeAny StaticType = 1 << iota
	TypeNumber
	TypeText
	TypeBin
	TypeDate
	TypeInterval
	TypeBool
	TypeXml
	TypeJson
	TypeNull
	TypeSet
	TypeRow
	TypeVoid
)

func (t StaticType) String() string {
	var str string
	switch t {
	case TypeAny:
		str = "any"
	case TypeNumber:
		str = "number"
	case TypeText:
		str = "text"
	case TypeBin:
		str = "bin"
	case TypeDate:
		str = "date"
	case TypeInterval:
		str = "interval"
	case TypeBool:
		str = "bool"
	case TypeXml:
		str = "xml"
	case TypeJson:
		str = "json"
	case TypeNull:
		str = "null"
	case TypeSet:
		str = "set"
	case TypeRow:
		str = "row"
	case TypeVoid:
		str = "void"
	}
	return str
}

func (t StaticType) IsCompatible(other StaticType) bool {
	if t == TypeAny || other == TypeAny {
		return true
	}
	return t == other
}

func getTypeFromBinary(left, right Node) StaticType {
	if t := getNodeType(left); t == getNodeType(right) {
		return t
	}
	return TypeAny
}

func getTypeFromName(name string) StaticType {
	switch strings.ToUpper(name) {
	case "CHAR", "VARCHAR", "CLOB":
		return TypeText
	case "BLOB":
		return TypeBin
	case "INT", "INTEGER", "DOUBLE", "REAL":
		return TypeNumber
	case "DATE", "TIME":
		return TypeDate
	case "INTERVAL":
		return TypeInterval
	case "BOOL", "BOOLEAN":
		return TypeBool
	case "ROW":
		return TypeRow
	default:
		return TypeAny
	}
}

func getNodeType(node Node) StaticType {
	if node == nil {
		return 0
	}
	t, ok := node.(TypedNode)
	if ok {
		return t.Type()
	}
	return 0
}

type OrderDir uint8

const (
	AscOrder OrderDir = 1 << iota
	DescOrder
)

type OnNull int8

const (
	NullOnNull OnNull = 1 << iota
	AbsentOnNull
)

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

type MaterializedMode int

const (
	MaterializedCte MaterializedMode = iota + 1
	NotMaterializedCte
)

type ParameterMode int

const (
	ModeIn ParameterMode = 1 << (iota + 1)
	ModeOut
	ModeInOut
)

type CascadeMode int

const (
	Cascade CascadeMode = 1 << iota
	Restrict
)

type IdentityMode int

const (
	RestartIdentity IdentityMode = 1 << iota
	ContinueIdentity
)

type TransactionMode int

const (
	ModeReadWrite TransactionMode = 1 << (iota + 1)
	ModeReadOnly
)

type TransactionLevel int

const (
	LevelReadRepeat TransactionLevel = 1 << (iota + 1)
	LevelReadCommit
	LevelReadUncommit
	LevelSerializable
)
