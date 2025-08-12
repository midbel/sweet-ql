package ast

type StaticType int8

const (
	TypeAny StaticType = 1 << iota
	TypeNumber
	TypeText
	TypeDate
	TypeBool
	TypeNull
)

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
