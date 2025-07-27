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
	VisitXmlAgg(XmlAgg) error
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

func Visit() Visitor {
	var noop noopVisitor
	return noop
}

func (_ noopVisitor) VisitValues(_ ValuesStatement) error {
	return nil
}

func (_ noopVisitor) VisitSelect(_ SelectStatement) error {
	return nil
}

func (_ noopVisitor) VisitUnion(_ UnionStatement) error {
	return nil
}

func (_ noopVisitor) VisitIntersect(_ IntersectStatement) error {
	return nil
}

func (_ noopVisitor) VisitExcept(_ ExceptStatement) error {
	return nil
}

func (_ noopVisitor) VisitInsert(_ InsertStatement) error {
	return nil
}

func (_ noopVisitor) VisitUpdate(_ UpdateStatement) error {
	return nil
}

func (_ noopVisitor) VisitDelete(_ DeleteStatement) error {
	return nil
}

func (_ noopVisitor) VisitTruncate(_ TruncateStatement) error {
	return nil
}

func (_ noopVisitor) VisitWith(_ WithStatement) error {
	return nil
}

func (_ noopVisitor) VisitCte(_ CteStatement) error {
	return nil
}

func (_ noopVisitor) VisitMerge(_ MergeStatement) error {
	return nil
}

func (_ noopVisitor) VisitMatch(_ MatchStatement) error {
	return nil
}

func (_ noopVisitor) VisitCall(_ CallStatement) error {
	return nil
}

func (_ noopVisitor) VisitGrant(_ GrantStatement) error {
	return nil
}

func (_ noopVisitor) VisitRevoke(_ RevokeStatement) error {
	return nil
}

func (_ noopVisitor) VisitCommit(_ Commit) error {
	return nil
}

func (_ noopVisitor) VisitRollback(_ Rollback) error {
	return nil
}

func (_ noopVisitor) VisitSetTransaction(_ SetTransaction) error {
	return nil
}

func (_ noopVisitor) VisitStartTransaction(_ StartTransaction) error {
	return nil
}

func (_ noopVisitor) VisitSavepoint(_ Savepoint) error {
	return nil
}

func (_ noopVisitor) VisitReleaseSavepoint(_ ReleaseSavepoint) error {
	return nil
}

func (_ noopVisitor) VisitRollbackSavepoint(_ RollbackSavepoint) error {
	return nil
}

func (_ noopVisitor) VisitJoin(_ Join) error {
	return nil
}

func (_ noopVisitor) VisitOrder(_ Order) error {
	return nil
}

func (_ noopVisitor) VisitLimit(_ Limit) error {
	return nil
}

func (_ noopVisitor) VisitOffset(_ Offset) error {
	return nil
}

func (_ noopVisitor) VisitBinary(_ Binary) error {
	return nil
}

func (_ noopVisitor) VisitUnary(_ Unary) error {
	return nil
}

func (_ noopVisitor) VisitCallFunc(_ Call) error {
	return nil
}

func (_ noopVisitor) VisitList(_ List) error {
	return nil
}

func (_ noopVisitor) VisitCollate(_ Collate) error {
	return nil
}

func (_ noopVisitor) VisitIn(_ In) error {
	return nil
}

func (_ noopVisitor) VisitIs(_ Is) error {
	return nil
}

func (_ noopVisitor) VisitExists(_ Exists) error {
	return nil
}

func (_ noopVisitor) VisitBetween(_ Between) error {
	return nil
}

func (_ noopVisitor) VisitAll(_ All) error {
	return nil
}

func (_ noopVisitor) VisitAny(_ Any) error {
	return nil
}

func (_ noopVisitor) VisitNot(_ Not) error {
	return nil
}

func (_ noopVisitor) VisitCast(_ Cast) error {
	return nil
}

func (_ noopVisitor) VisitValue(_ Value) error {
	return nil
}

func (_ noopVisitor) VisitAlias(_ Alias) error {
	return nil
}

func (_ noopVisitor) VisitName(_ Name) error {
	return nil
}

func (_ noopVisitor) VisitGroup(_ Group) error {
	return nil
}

func (_ noopVisitor) VisitAssignment(_ Assignment) error {
	return nil
}

func (_ noopVisitor) VisitBody(_ Body) error {
	return nil
}

func (_ noopVisitor) VisitIf(_ If) error {
	return nil
}

func (_ noopVisitor) VisitWhile(_ While) error {
	return nil
}

func (_ noopVisitor) VisitSet(_ Set) error {
	return nil
}

func (_ noopVisitor) VisitDeclare(_ Declare) error {
	return nil
}

func (_ noopVisitor) VisitReturn(_ Return) error {
	return nil
}

func (_ noopVisitor) VisitCase(_ Case) error {
	return nil
}

func (_ noopVisitor) VisitWhen(_ When) error {
	return nil
}

func (_ noopVisitor) VisitCreateProcedure(CreateProcedureStatement) error {
	return nil
}

func (_ noopVisitor) VisitCreateTable(_ CreateTableStatement) error {
	return nil
}

func (_ noopVisitor) VisitDropTable(_ DropTableStatement) error {
	return nil
}

func (_ noopVisitor) VisitAlterTable(_ AlterTableStatement) error {
	return nil
}

func (_ noopVisitor) VisitCreateView(_ CreateViewStatement) error {
	return nil
}

func (_ noopVisitor) VisitDropView(_ DropViewStatement) error {
	return nil
}

func (_ noopVisitor) VisitColumnDef(_ ColumnDef) error {
	return nil
}

func (_ noopVisitor) VisitAddColumn(_ AddColumnAction) error {
	return nil
}

func (_ noopVisitor) VisitAlterColumn(_ AlterColumnAction) error {
	return nil
}

func (_ noopVisitor) VisitDropColumn(_ DropColumnAction) error {
	return nil
}

func (_ noopVisitor) VisitAddConstraint(_ AddConstraintAction) error {
	return nil
}

func (_ noopVisitor) VisitDropConstraint(_ DropConstraintAction) error {
	return nil
}

func (_ noopVisitor) VisitConstraint(_ Constraint) error {
	return nil
}

func (_ noopVisitor) VisitPrimaryKey(_ PrimaryKeyConstraint) error {
	return nil
}

func (_ noopVisitor) VisitForeignKey(_ ForeignKeyConstraint) error {
	return nil
}

func (_ noopVisitor) VisitNotNull(_ NotNullConstraint) error {
	return nil
}

func (_ noopVisitor) VisitUnique(_ UniqueConstraint) error {
	return nil
}

func (_ noopVisitor) VisitCheck(_ CheckConstraint) error {
	return nil
}

func (_ noopVisitor) VisitDefault(_ DefaultConstraint) error {
	return nil
}

func (_ noopVisitor) VisitGenerated(_ GeneratedConstraint) error {
	return nil
}

func (_ noopVisitor) VisitXmlElement(_ XmlElement) error {
	return nil
}

func (_ noopVisitor) VisitXmlAttribute(_ XmlAttribute) error {
	return nil
}

func (_ noopVisitor) VisitXmlNamespace(_ XmlNamespace) error {
	return nil
}

func (_ noopVisitor) VisitXmlText(_ XmlText) error {
	return nil
}

func (_ noopVisitor) VisitXmlComment(_ XmlComment) error {
	return nil
}

func (_ noopVisitor) VisitXmlAgg(_ XmlAgg) error {
	return nil
}
