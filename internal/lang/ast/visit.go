package ast

type VisitableNode interface {
	Accept(Visitor) error
}

type ControlVisitor interface {
	VisitBody(Body) error
	VisitIf(If) error
	VisitWhile(While) error
	VisitSet(Set) error
	VisitDeclare(Declare) error
	VisitReturn(Return) error
}

type StmtVisitor interface {
	VisitValues(ValuesStatement) error
	VisitSelect(SelectStatement) error
	VisitUnion(UnionStatement) error
	VisitIntersect(IntersectStatement) error
	VisitExcept(ExceptStatement) error
	VisitInsert(InsertStatement) error
	VisitUpdate(UpdateStatement) error
	VisitDelete(DeleteStatement) error
	VisitTruncate(TruncateStatement) error
	VisitWith(WithStatement) error
	VisitCte(CteStatement) error
	VisitMerge(MergeStatement) error

	VisitCall(CallStatement) error

	VisitGrant(GrantStatement) error
	VisitRevoke(RevokeStatement) error

	VisitCommit(Commit) error
	VisitRollback(Rollback) error
	VisitSetTransaction(SetTransaction) error
	VisitStartTransaction(StartTransaction) error
	VisitSavepoint(Savepoint) error
	VisitReleaseSavepoint(ReleaseSavepoint) error
	VisitRollbackSavepoint(RollbackSavepoint) error

	VisitMatch(MatchStatement) error
	VisitJoin(Join) error
	VisitOrder(Order) error
	VisitLimit(Limit) error
	VisitOffset(Offset) error
}

type ExprVisitor interface {
	VisitBinary(Binary) error
	VisitUnary(Unary) error
	VisitList(List) error
	VisitCollate(Collate) error
	VisitIn(In) error
	VisitIs(Is) error
	VisitExists(Exists) error
	VisitBetween(Between) error
	VisitAll(All) error
	VisitAny(Any) error
	VisitNot(Not) error
	VisitCast(Cast) error
	VisitCallFunc(Call) error

	VisitValue(Value) error
	VisitAlias(Alias) error
	VisitName(Name) error
	VisitGroup(Group) error
	VisitCase(Case) error
	VisitWhen(When) error
	VisitAssignment(Assignment) error
}

type XmlVisitor interface {
	VisitXmlElement(XmlElement) error
	VisitXmlAttribute(XmlAttribute) error
	VisitXmlNamespace(XmlNamespace) error
	VisitXmlText(XmlText) error
	VisitXmlComment(XmlComment) error
	VisitXmlPi(XmlPi) error
	VisitXmlConcat(XmlConcat) error
	VisitXmlAgg(XmlAgg) error
	VisitXmlRoot(XmlRoot) error
	VisitXmlForest(XmlForest) error
}

type DefinitionVisitor interface {
	VisitCreateProcedure(CreateProcedureStatement) error

	VisitCreateTable(CreateTableStatement) error
	VisitDropTable(DropTableStatement) error
	VisitAlterTable(AlterTableStatement) error
	VisitCreateView(CreateViewStatement) error
	VisitDropView(DropViewStatement) error

	VisitColumnDef(ColumnDef) error
	VisitAddColumn(AddColumnAction) error
	VisitAlterColumn(AlterColumnAction) error
	VisitDropColumn(DropColumnAction) error
	VisitAddConstraint(AddConstraintAction) error
	VisitDropConstraint(DropConstraintAction) error
	VisitRenameTable(RenameTableAction) error
	VisitRenameColumn(RenameColumnAction) error
	VisitRenameConstraint(RenameConstraintAction) error

	VisitSetDefaultConstraint(SetDefaultConstraint) error
	VisitDropDefaultConstraint(DropDefaultConstraint) error
	VisitSetNotNullConstraint(SetNotNullConstraint) error
	VisitDropNotNullConstraint(DropNotNullConstraint) error
	VisitSetTypeConstraint(SetTypeConstraint) error

	VisitConstraint(Constraint) error
	VisitPrimaryKey(PrimaryKeyConstraint) error
	VisitForeignKey(ForeignKeyConstraint) error
	VisitNotNull(NotNullConstraint) error
	VisitUnique(UniqueConstraint) error
	VisitCheck(CheckConstraint) error
	VisitDefault(DefaultConstraint) error
	VisitGenerated(GeneratedConstraint) error
}

type Visitor interface {
	StmtVisitor
	ExprVisitor
	XmlVisitor
	ControlVisitor
	DefinitionVisitor
}

type noopVisitor struct{}

func Noop() Visitor {
	var noop noopVisitor
	return noop
}

func (noopVisitor) VisitValues(ValuesStatement) error {
	return nil
}

func (noopVisitor) VisitSelect(SelectStatement) error {
	return nil
}

func (noopVisitor) VisitUnion(UnionStatement) error {
	return nil
}

func (noopVisitor) VisitIntersect(IntersectStatement) error {
	return nil
}

func (noopVisitor) VisitExcept(ExceptStatement) error {
	return nil
}

func (noopVisitor) VisitInsert(InsertStatement) error {
	return nil
}

func (noopVisitor) VisitUpdate(UpdateStatement) error {
	return nil
}

func (noopVisitor) VisitDelete(DeleteStatement) error {
	return nil
}

func (noopVisitor) VisitTruncate(TruncateStatement) error {
	return nil
}

func (noopVisitor) VisitWith(WithStatement) error {
	return nil
}

func (noopVisitor) VisitCte(CteStatement) error {
	return nil
}

func (noopVisitor) VisitMerge(MergeStatement) error {
	return nil
}

func (noopVisitor) VisitMatch(MatchStatement) error {
	return nil
}

func (noopVisitor) VisitCall(CallStatement) error {
	return nil
}

func (noopVisitor) VisitGrant(GrantStatement) error {
	return nil
}

func (noopVisitor) VisitRevoke(RevokeStatement) error {
	return nil
}

func (noopVisitor) VisitCommit(Commit) error {
	return nil
}

func (noopVisitor) VisitRollback(Rollback) error {
	return nil
}

func (noopVisitor) VisitSetTransaction(SetTransaction) error {
	return nil
}

func (noopVisitor) VisitStartTransaction(StartTransaction) error {
	return nil
}

func (noopVisitor) VisitSavepoint(Savepoint) error {
	return nil
}

func (noopVisitor) VisitReleaseSavepoint(ReleaseSavepoint) error {
	return nil
}

func (noopVisitor) VisitRollbackSavepoint(RollbackSavepoint) error {
	return nil
}

func (noopVisitor) VisitJoin(Join) error {
	return nil
}

func (noopVisitor) VisitOrder(Order) error {
	return nil
}

func (noopVisitor) VisitLimit(Limit) error {
	return nil
}

func (noopVisitor) VisitOffset(Offset) error {
	return nil
}

func (noopVisitor) VisitBinary(Binary) error {
	return nil
}

func (noopVisitor) VisitUnary(Unary) error {
	return nil
}

func (noopVisitor) VisitCallFunc(Call) error {
	return nil
}

func (noopVisitor) VisitList(List) error {
	return nil
}

func (noopVisitor) VisitCollate(Collate) error {
	return nil
}

func (noopVisitor) VisitIn(In) error {
	return nil
}

func (noopVisitor) VisitIs(Is) error {
	return nil
}

func (noopVisitor) VisitExists(Exists) error {
	return nil
}

func (noopVisitor) VisitBetween(Between) error {
	return nil
}

func (noopVisitor) VisitAll(All) error {
	return nil
}

func (noopVisitor) VisitAny(Any) error {
	return nil
}

func (noopVisitor) VisitNot(Not) error {
	return nil
}

func (noopVisitor) VisitCast(Cast) error {
	return nil
}

func (noopVisitor) VisitValue(Value) error {
	return nil
}

func (noopVisitor) VisitAlias(Alias) error {
	return nil
}

func (noopVisitor) VisitName(Name) error {
	return nil
}

func (noopVisitor) VisitGroup(Group) error {
	return nil
}

func (noopVisitor) VisitAssignment(Assignment) error {
	return nil
}

func (noopVisitor) VisitBody(Body) error {
	return nil
}

func (noopVisitor) VisitIf(If) error {
	return nil
}

func (noopVisitor) VisitWhile(While) error {
	return nil
}

func (noopVisitor) VisitSet(Set) error {
	return nil
}

func (noopVisitor) VisitDeclare(Declare) error {
	return nil
}

func (noopVisitor) VisitReturn(Return) error {
	return nil
}

func (noopVisitor) VisitCase(Case) error {
	return nil
}

func (noopVisitor) VisitWhen(When) error {
	return nil
}

func (noopVisitor) VisitCreateProcedure(CreateProcedureStatement) error {
	return nil
}

func (noopVisitor) VisitCreateTable(CreateTableStatement) error {
	return nil
}

func (noopVisitor) VisitDropTable(DropTableStatement) error {
	return nil
}

func (noopVisitor) VisitAlterTable(AlterTableStatement) error {
	return nil
}

func (noopVisitor) VisitCreateView(CreateViewStatement) error {
	return nil
}

func (noopVisitor) VisitDropView(DropViewStatement) error {
	return nil
}

func (noopVisitor) VisitColumnDef(ColumnDef) error {
	return nil
}

func (noopVisitor) VisitAddColumn(AddColumnAction) error {
	return nil
}

func (noopVisitor) VisitAlterColumn(AlterColumnAction) error {
	return nil
}

func (noopVisitor) VisitDropColumn(DropColumnAction) error {
	return nil
}

func (noopVisitor) VisitAddConstraint(AddConstraintAction) error {
	return nil
}

func (noopVisitor) VisitDropConstraint(DropConstraintAction) error {
	return nil
}

func (noopVisitor) VisitRenameTable(RenameTableAction) error {
	return nil
}

func (noopVisitor) VisitRenameColumn(RenameColumnAction) error {
	return nil
}

func (noopVisitor) VisitRenameConstraint(RenameConstraintAction) error {
	return nil
}

func (noopVisitor) VisitConstraint(Constraint) error {
	return nil
}

func (noopVisitor) VisitSetDefaultConstraint(SetDefaultConstraint) error {
	return nil
}

func (noopVisitor) VisitDropDefaultConstraint(DropDefaultConstraint) error {
	return nil
}

func (noopVisitor) VisitSetNotNullConstraint(SetNotNullConstraint) error {
	return nil
}

func (noopVisitor) VisitDropNotNullConstraint(DropNotNullConstraint) error {
	return nil
}

func (noopVisitor) VisitSetTypeConstraint(SetTypeConstraint) error {
	return nil
}

func (noopVisitor) VisitPrimaryKey(PrimaryKeyConstraint) error {
	return nil
}

func (noopVisitor) VisitForeignKey(ForeignKeyConstraint) error {
	return nil
}

func (noopVisitor) VisitNotNull(NotNullConstraint) error {
	return nil
}

func (noopVisitor) VisitUnique(UniqueConstraint) error {
	return nil
}

func (noopVisitor) VisitCheck(CheckConstraint) error {
	return nil
}

func (noopVisitor) VisitDefault(DefaultConstraint) error {
	return nil
}

func (noopVisitor) VisitGenerated(GeneratedConstraint) error {
	return nil
}

func (noopVisitor) VisitXmlElement(XmlElement) error {
	return nil
}

func (noopVisitor) VisitXmlAttribute(XmlAttribute) error {
	return nil
}

func (noopVisitor) VisitXmlNamespace(XmlNamespace) error {
	return nil
}

func (noopVisitor) VisitXmlText(XmlText) error {
	return nil
}

func (noopVisitor) VisitXmlComment(XmlComment) error {
	return nil
}

func (noopVisitor) VisitXmlPi(XmlPi) error {
	return nil
}

func (noopVisitor) VisitXmlConcat(XmlConcat) error {
	return nil
}

func (noopVisitor) VisitXmlAgg(XmlAgg) error {
	return nil
}

func (noopVisitor) VisitXmlRoot(XmlRoot) error {
	return nil
}

func (noopVisitor) VisitXmlForest(XmlForest) error {
	return nil
}
