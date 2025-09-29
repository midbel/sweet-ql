package ast

type VisitableNode interface {
	Accept(Visitor) error
}

type TransformableNode interface {
	Transform(Transformer) (Node, error)
}

type ControlVisitor interface {
	VisitBody(*Body) error
	VisitIf(*If) error
	VisitWhile(*While) error
	VisitSet(*Set) error
	VisitDeclare(*Declare) error
	VisitReturn(*Return) error
}

type StmtVisitor interface {
	VisitValues(*ValuesStatement) error
	VisitSelect(*SelectStatement) error
	VisitUnion(*UnionStatement) error
	VisitIntersect(*IntersectStatement) error
	VisitExcept(*ExceptStatement) error
	VisitInsert(*InsertStatement) error
	VisitUpdate(*UpdateStatement) error
	VisitDelete(*DeleteStatement) error
	VisitTruncate(*TruncateStatement) error
	VisitWith(*WithStatement) error
	VisitCte(*CteStatement) error
	VisitMerge(*MergeStatement) error

	VisitCall(*CallStatement) error

	VisitGrant(*GrantStatement) error
	VisitRevoke(*RevokeStatement) error

	VisitCommit(*Commit) error
	VisitRollback(*Rollback) error
	VisitSetTransaction(*SetTransaction) error
	VisitStartTransaction(*StartTransaction) error
	VisitSavepoint(*Savepoint) error
	VisitReleaseSavepoint(*ReleaseSavepoint) error
	VisitRollbackSavepoint(*RollbackSavepoint) error

	VisitMatch(*MatchStatement) error
	VisitJoin(*Join) error
	VisitOrder(*Order) error
	VisitLimit(*Limit) error
	VisitOffset(*Offset) error
}

type ExprVisitor interface {
	VisitBinary(*Binary) error
	VisitUnary(*Unary) error
	VisitList(*List) error
	VisitCollate(*Collate) error
	VisitIn(*In) error
	VisitIs(*Is) error
	VisitExists(*Exists) error
	VisitBetween(*Between) error
	VisitAll(*All) error
	VisitAny(*Any) error
	VisitNot(*Not) error
	VisitCast(*Cast) error
	VisitCallFunc(*Call) error

	VisitPlaceholder(*Placeholder) error
	VisitValue(*Value) error
	VisitAlias(*Alias) error
	VisitName(*Name) error
	VisitGroup(*Group) error
	VisitCase(*Case) error
	VisitWhen(*When) error
	VisitAssignment(*Assignment) error
}

type XmlVisitor interface {
	VisitXmlElement(*XmlElement) error
	VisitXmlAttribute(*XmlAttribute) error
	VisitXmlNamespace(*XmlNamespace) error
	VisitXmlText(*XmlText) error
	VisitXmlComment(*XmlComment) error
	VisitXmlPi(*XmlPi) error
	VisitXmlConcat(*XmlConcat) error
	VisitXmlAgg(*XmlAgg) error
	VisitXmlRoot(*XmlRoot) error
	VisitXmlForest(*XmlForest) error
}

type DefinitionVisitor interface {
	VisitCreateProcedure(*CreateProcedureStatement) error

	VisitCreateTable(*CreateTableStatement) error
	VisitDropTable(*DropTableStatement) error
	VisitAlterTable(*AlterTableStatement) error
	VisitCreateView(*CreateViewStatement) error
	VisitDropView(*DropViewStatement) error

	VisitColumnDef(*ColumnDef) error
	VisitAddColumn(*AddColumnAction) error
	VisitAlterColumn(*AlterColumnAction) error
	VisitDropColumn(*DropColumnAction) error
	VisitAddConstraint(*AddConstraintAction) error
	VisitDropConstraint(*DropConstraintAction) error
	VisitRenameTable(*RenameTableAction) error
	VisitRenameColumn(*RenameColumnAction) error
	VisitRenameConstraint(*RenameConstraintAction) error

	VisitSetDefaultConstraint(*SetDefaultConstraint) error
	VisitDropDefaultConstraint(*DropDefaultConstraint) error
	VisitSetNotNullConstraint(*SetNotNullConstraint) error
	VisitDropNotNullConstraint(*DropNotNullConstraint) error
	VisitSetTypeConstraint(*SetTypeConstraint) error

	VisitConstraint(*Constraint) error
	VisitPrimaryKey(*PrimaryKeyConstraint) error
	VisitForeignKey(*ForeignKeyConstraint) error
	VisitNotNull(*NotNullConstraint) error
	VisitUnique(*UniqueConstraint) error
	VisitCheck(*CheckConstraint) error
	VisitDefault(*DefaultConstraint) error
	VisitGenerated(*GeneratedConstraint) error
}

type StmtTransformer interface {
	TransformValues(*ValuesStatement) (Node, error)
	TransformSelect(*SelectStatement) (Node, error)
	TransformUnion(*UnionStatement) (Node, error)
	TransformIntersect(*IntersectStatement) (Node, error)
	TransformExcept(*ExceptStatement) (Node, error)
	TransformInsert(*InsertStatement) (Node, error)
	TransformUpdate(*UpdateStatement) (Node, error)
	TransformDelete(*DeleteStatement) (Node, error)
	TransformTruncate(*TruncateStatement) (Node, error)
	TransformWith(*WithStatement) (Node, error)
	TransformCte(*CteStatement) (Node, error)
	TransformMerge(*MergeStatement) (Node, error)

	TransformCall(*CallStatement) (Node, error)

	TransformMatch(*MatchStatement) (Node, error)
	TransformJoin(*Join) (Node, error)
	TransformOrder(*Order) (Node, error)
	TransformLimit(*Limit) (Node, error)
	TransformOffset(*Offset) (Node, error)
}

type ExprTransformer interface {
	TransformBinary(*Binary) (Node, error)
	TransformUnary(*Unary) (Node, error)
	TransformList(*List) (Node, error)
	TransformCollate(*Collate) (Node, error)
	TransformIn(*In) (Node, error)
	TransformIs(*Is) (Node, error)
	TransformExists(*Exists) (Node, error)
	TransformBetween(*Between) (Node, error)
	TransformAll(*All) (Node, error)
	TransformAny(*Any) (Node, error)
	TransformNot(*Not) (Node, error)
	TransformCast(*Cast) (Node, error)
	TransformCallFunc(*Call) (Node, error)
	TransformValue(*Value) (Node, error)
	TransformAlias(*Alias) (Node, error)
	TransformName(*Name) (Node, error)
	TransformGroup(*Group) (Node, error)
	TransformCase(*Case) (Node, error)
	TransformWhen(*When) (Node, error)
	TransformAssignment(*Assignment) (Node, error)
}

type Transformer interface {
	StmtTransformer
	ExprTransformer
}

type Visitor interface {
	StmtVisitor
	ExprVisitor
	XmlVisitor
	ControlVisitor
	DefinitionVisitor
}

type nameVisitor struct {
	Visitor
	do func(*Name) error
}

func VisitName(do func(*Name) error) Visitor {
	return nameVisitor{
		Visitor: Noop(),
		do:      do,
	}
}

func (n nameVisitor) VisitName(name *Name) error {
	return n.do(name)
}

type selectVisitor struct {
	Visitor
	do func(*SelectStatement) error
}

func VisitSelect(do func(*SelectStatement) error) Visitor {
	return &selectVisitor{
		Visitor: Noop(),
		do:      do,
	}
}

func (i *selectVisitor) VisitSelect(stmt *SelectStatement) error {
	return i.do(stmt)
}

type literalVisitor struct {
	Visitor
	check func(*Value) error
}

func VisitLiteral(check func(*Value) error) Visitor {
	return &literalVisitor{
		Visitor: Noop(),
		check:   check,
	}
}

func (i *literalVisitor) VisitValue(value *Value) error {
	return i.check(value)
}

type callFuncVisitor struct {
	Visitor
	check func(*Call) error
}

func VisitCallFunc(check func(*Call) error) Visitor {
	return &callFuncVisitor{
		Visitor: Noop(),
		check:   check,
	}
}

func (v *callFuncVisitor) VisitCallFunc(call *Call) error {
	return v.check(call)
}

type noopVisitor struct{}

func Noop() Visitor {
	var noop noopVisitor
	return noop
}

func (noopVisitor) VisitValues(*ValuesStatement) error {
	return nil
}

func (noopVisitor) VisitSelect(*SelectStatement) error {
	return nil
}

func (noopVisitor) VisitUnion(*UnionStatement) error {
	return nil
}

func (noopVisitor) VisitIntersect(*IntersectStatement) error {
	return nil
}

func (noopVisitor) VisitExcept(*ExceptStatement) error {
	return nil
}

func (noopVisitor) VisitInsert(*InsertStatement) error {
	return nil
}

func (noopVisitor) VisitUpdate(*UpdateStatement) error {
	return nil
}

func (noopVisitor) VisitDelete(*DeleteStatement) error {
	return nil
}

func (noopVisitor) VisitTruncate(*TruncateStatement) error {
	return nil
}

func (noopVisitor) VisitWith(*WithStatement) error {
	return nil
}

func (noopVisitor) VisitCte(*CteStatement) error {
	return nil
}

func (noopVisitor) VisitMerge(*MergeStatement) error {
	return nil
}

func (noopVisitor) VisitMatch(*MatchStatement) error {
	return nil
}

func (noopVisitor) VisitCall(*CallStatement) error {
	return nil
}

func (noopVisitor) VisitGrant(*GrantStatement) error {
	return nil
}

func (noopVisitor) VisitRevoke(*RevokeStatement) error {
	return nil
}

func (noopVisitor) VisitCommit(*Commit) error {
	return nil
}

func (noopVisitor) VisitRollback(*Rollback) error {
	return nil
}

func (noopVisitor) VisitSetTransaction(*SetTransaction) error {
	return nil
}

func (noopVisitor) VisitStartTransaction(*StartTransaction) error {
	return nil
}

func (noopVisitor) VisitSavepoint(*Savepoint) error {
	return nil
}

func (noopVisitor) VisitReleaseSavepoint(*ReleaseSavepoint) error {
	return nil
}

func (noopVisitor) VisitRollbackSavepoint(*RollbackSavepoint) error {
	return nil
}

func (noopVisitor) VisitJoin(*Join) error {
	return nil
}

func (noopVisitor) VisitOrder(*Order) error {
	return nil
}

func (noopVisitor) VisitLimit(*Limit) error {
	return nil
}

func (noopVisitor) VisitOffset(*Offset) error {
	return nil
}

func (noopVisitor) VisitBinary(*Binary) error {
	return nil
}

func (noopVisitor) VisitUnary(*Unary) error {
	return nil
}

func (noopVisitor) VisitCallFunc(*Call) error {
	return nil
}

func (noopVisitor) VisitList(*List) error {
	return nil
}

func (noopVisitor) VisitCollate(*Collate) error {
	return nil
}

func (noopVisitor) VisitIn(*In) error {
	return nil
}

func (noopVisitor) VisitIs(*Is) error {
	return nil
}

func (noopVisitor) VisitExists(*Exists) error {
	return nil
}

func (noopVisitor) VisitBetween(*Between) error {
	return nil
}

func (noopVisitor) VisitAll(*All) error {
	return nil
}

func (noopVisitor) VisitAny(*Any) error {
	return nil
}

func (noopVisitor) VisitNot(*Not) error {
	return nil
}

func (noopVisitor) VisitCast(*Cast) error {
	return nil
}

func (noopVisitor) VisitPlaceholder(*Placeholder) error {
	return nil
}

func (noopVisitor) VisitValue(*Value) error {
	return nil
}

func (noopVisitor) VisitAlias(*Alias) error {
	return nil
}

func (noopVisitor) VisitName(*Name) error {
	return nil
}

func (noopVisitor) VisitGroup(*Group) error {
	return nil
}

func (noopVisitor) VisitAssignment(*Assignment) error {
	return nil
}

func (noopVisitor) VisitBody(*Body) error {
	return nil
}

func (noopVisitor) VisitIf(*If) error {
	return nil
}

func (noopVisitor) VisitWhile(*While) error {
	return nil
}

func (noopVisitor) VisitSet(*Set) error {
	return nil
}

func (noopVisitor) VisitDeclare(*Declare) error {
	return nil
}

func (noopVisitor) VisitReturn(*Return) error {
	return nil
}

func (noopVisitor) VisitCase(*Case) error {
	return nil
}

func (noopVisitor) VisitWhen(*When) error {
	return nil
}

func (noopVisitor) VisitCreateProcedure(*CreateProcedureStatement) error {
	return nil
}

func (noopVisitor) VisitCreateTable(*CreateTableStatement) error {
	return nil
}

func (noopVisitor) VisitDropTable(*DropTableStatement) error {
	return nil
}

func (noopVisitor) VisitAlterTable(*AlterTableStatement) error {
	return nil
}

func (noopVisitor) VisitCreateView(*CreateViewStatement) error {
	return nil
}

func (noopVisitor) VisitDropView(*DropViewStatement) error {
	return nil
}

func (noopVisitor) VisitColumnDef(*ColumnDef) error {
	return nil
}

func (noopVisitor) VisitAddColumn(*AddColumnAction) error {
	return nil
}

func (noopVisitor) VisitAlterColumn(*AlterColumnAction) error {
	return nil
}

func (noopVisitor) VisitDropColumn(*DropColumnAction) error {
	return nil
}

func (noopVisitor) VisitAddConstraint(*AddConstraintAction) error {
	return nil
}

func (noopVisitor) VisitDropConstraint(*DropConstraintAction) error {
	return nil
}

func (noopVisitor) VisitRenameTable(*RenameTableAction) error {
	return nil
}

func (noopVisitor) VisitRenameColumn(*RenameColumnAction) error {
	return nil
}

func (noopVisitor) VisitRenameConstraint(*RenameConstraintAction) error {
	return nil
}

func (noopVisitor) VisitConstraint(*Constraint) error {
	return nil
}

func (noopVisitor) VisitSetDefaultConstraint(*SetDefaultConstraint) error {
	return nil
}

func (noopVisitor) VisitDropDefaultConstraint(*DropDefaultConstraint) error {
	return nil
}

func (noopVisitor) VisitSetNotNullConstraint(*SetNotNullConstraint) error {
	return nil
}

func (noopVisitor) VisitDropNotNullConstraint(*DropNotNullConstraint) error {
	return nil
}

func (noopVisitor) VisitSetTypeConstraint(*SetTypeConstraint) error {
	return nil
}

func (noopVisitor) VisitPrimaryKey(*PrimaryKeyConstraint) error {
	return nil
}

func (noopVisitor) VisitForeignKey(*ForeignKeyConstraint) error {
	return nil
}

func (noopVisitor) VisitNotNull(*NotNullConstraint) error {
	return nil
}

func (noopVisitor) VisitUnique(*UniqueConstraint) error {
	return nil
}

func (noopVisitor) VisitCheck(*CheckConstraint) error {
	return nil
}

func (noopVisitor) VisitDefault(*DefaultConstraint) error {
	return nil
}

func (noopVisitor) VisitGenerated(*GeneratedConstraint) error {
	return nil
}

func (noopVisitor) VisitXmlElement(*XmlElement) error {
	return nil
}

func (noopVisitor) VisitXmlAttribute(*XmlAttribute) error {
	return nil
}

func (noopVisitor) VisitXmlNamespace(*XmlNamespace) error {
	return nil
}

func (noopVisitor) VisitXmlText(*XmlText) error {
	return nil
}

func (noopVisitor) VisitXmlComment(*XmlComment) error {
	return nil
}

func (noopVisitor) VisitXmlPi(*XmlPi) error {
	return nil
}

func (noopVisitor) VisitXmlConcat(*XmlConcat) error {
	return nil
}

func (noopVisitor) VisitXmlAgg(*XmlAgg) error {
	return nil
}

func (noopVisitor) VisitXmlRoot(*XmlRoot) error {
	return nil
}

func (noopVisitor) VisitXmlForest(*XmlForest) error {
	return nil
}

type noopTransformer struct{}

func Keep() Transformer {
	var t noopTransformer
	return t
}

func (noopTransformer) TransformValues(stmt *ValuesStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformSelect(stmt *SelectStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformUnion(stmt *UnionStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformIntersect(stmt *IntersectStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformExcept(stmt *ExceptStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformInsert(stmt *InsertStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformUpdate(stmt *UpdateStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformDelete(stmt *DeleteStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformTruncate(stmt *TruncateStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformWith(stmt *WithStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformCte(stmt *CteStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformMerge(stmt *MergeStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformCall(stmt *CallStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformMatch(stmt *MatchStatement) (Node, error) {
	return stmt, nil
}

func (noopTransformer) TransformJoin(join *Join) (Node, error) {
	return join, nil
}

func (noopTransformer) TransformOrder(order *Order) (Node, error) {
	return order, nil
}

func (noopTransformer) TransformLimit(limit *Limit) (Node, error) {
	return limit, nil
}

func (noopTransformer) TransformOffset(offset *Offset) (Node, error) {
	return offset, nil
}

func (noopTransformer) TransformBinary(binary *Binary) (Node, error) {
	return binary, nil
}

func (noopTransformer) TransformUnary(unary *Unary) (Node, error) {
	return unary, nil
}

func (noopTransformer) TransformList(list *List) (Node, error) {
	return list, nil
}

func (noopTransformer) TransformCollate(collate *Collate) (Node, error) {
	return collate, nil
}

func (noopTransformer) TransformIn(in *In) (Node, error) {
	return in, nil
}

func (noopTransformer) TransformIs(is *Is) (Node, error) {
	return is, nil
}

func (noopTransformer) TransformExists(exists *Exists) (Node, error) {
	return exists, nil
}

func (noopTransformer) TransformBetween(between *Between) (Node, error) {
	return between, nil
}

func (noopTransformer) TransformAll(all *All) (Node, error) {
	return all, nil
}

func (noopTransformer) TransformAny(any *Any) (Node, error) {
	return any, nil
}

func (noopTransformer) TransformNot(not *Not) (Node, error) {
	return not, nil
}

func (noopTransformer) TransformCast(cast *Cast) (Node, error) {
	return cast, nil
}

func (noopTransformer) TransformCallFunc(call *Call) (Node, error) {
	return call, nil
}

func (noopTransformer) TransformValue(value *Value) (Node, error) {
	return value, nil
}

func (noopTransformer) TransformAlias(alias *Alias) (Node, error) {
	return alias, nil
}

func (noopTransformer) TransformName(name *Name) (Node, error) {
	return name, nil
}

func (noopTransformer) TransformGroup(group *Group) (Node, error) {
	return group, nil
}

func (noopTransformer) TransformCase(cas *Case) (Node, error) {
	return cas, nil
}

func (noopTransformer) TransformWhen(when *When) (Node, error) {
	return when, nil
}

func (noopTransformer) TransformAssignment(assign *Assignment) (Node, error) {
	return assign, nil
}
