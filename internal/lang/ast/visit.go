package ast

type VisitableNode interface {
	Accept(Visitor)
}

type ControlVisitor interface {
	VisitBody(Body)
	VisitIf(If)
	VisitWhile(While)
	VisitSet(Set)
	VisitDeclare(Declare)
	VisitReturn(Return)
}

type StmtVisitor interface {
	VisitValues(ValuesStatement)
	VisitSelect(SelectStatement)
	VisitUnion(UnionStatement)
	VisitIntersect(IntersectStatement)
	VisitExcept(ExceptStatement)
	VisitInsert(InsertStatement)
	VisitUpdate(UpdateStatement)
	VisitDelete(DeleteStatement)
	VisitTruncate(TruncateStatement)
	VisitWith(WithStatement)
	VisitCte(CteStatement)
	VisitMerge(MergeStatement)

	VisitCall(CallStatement)

	VisitGrant(GrantStatement)
	VisitRevoke(RevokeStatement)

	VisitCommit(Commit)
	VisitRollback(Rollback)
	VisitSetTransaction(SetTransaction)
	VisitStartTransaction(StartTransaction)
	VisitSavepoint(Savepoint)
	VisitReleaseSavepoint(ReleaseSavepoint)
	VisitRollbackSavepoint(RollbackSavepoint)

	VisitMatch(MatchStatement)
	VisitJoin(Join)
	VisitOrder(Order)
	VisitLimit(Limit)
	VisitOffset(Offset)
}

type ExprVisitor interface {
	VisitBinary(Binary)
	VisitUnary(Unary)
	VisitList(List)
	VisitCollate(Collate)
	VisitIn(In)
	VisitIs(Is)
	VisitExists(Exists)
	VisitBetween(Between)
	VisitAll(All)
	VisitAny(Any)
	VisitNot(Not)
	VisitCast(Cast)
	VisitCallFunc(Call)

	VisitValue(Value)
	VisitAlias(Alias)
	VisitName(Name)
	VisitGroup(Group)
	VisitCase(Case)
	VisitWhen(When)
	VisitAssignment(Assignment)
}

type XmlVisitor interface {
	VisitXmlElement(XmlElement)
	VisitXmlAttribute(XmlAttribute)
	VisitXmlNamespace(XmlNamespace)
	VisitXmlText(XmlText)
	VisitXmlComment(XmlComment)
	VisitXmlAgg(XmlAgg)
}

type DefinitionVisitor interface {
	VisitCreateProcedure(CreateProcedureStatement)

	VisitCreateTable(CreateTableStatement)
	VisitDropTable(DropTableStatement)
	VisitAlterTable(AlterTableStatement)
	VisitCreateView(CreateViewStatement)
	VisitDropView(DropViewStatement)

	VisitColumnDef(ColumnDef)
	VisitAddColumn(AddColumnAction)
	VisitAlterColumn(AlterColumnAction)
	VisitDropColumn(DropColumnAction)
	VisitAddConstraint(AddConstraintAction)
	VisitDropConstraint(DropConstraintAction)

	VisitConstraint(Constraint)
	VisitPrimaryKey(PrimaryKeyConstraint)
	VisitForeignKey(ForeignKeyConstraint)
	VisitNotNull(NotNullConstraint)
	VisitUnique(UniqueConstraint)
	VisitCheck(CheckConstraint)
	VisitDefault(DefaultConstraint)
	VisitGenerated(GeneratedConstraint)
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

func (_ noopVisitor) VisitValues(_ ValuesStatement) {}

func (_ noopVisitor) VisitSelect(_ SelectStatement) {}

func (_ noopVisitor) VisitUnion(_ UnionStatement) {}

func (_ noopVisitor) VisitIntersect(_ IntersectStatement) {}

func (_ noopVisitor) VisitExcept(_ ExceptStatement) {}

func (_ noopVisitor) VisitInsert(_ InsertStatement) {}

func (_ noopVisitor) VisitUpdate(_ UpdateStatement) {}

func (_ noopVisitor) VisitDelete(_ DeleteStatement) {}

func (_ noopVisitor) VisitTruncate(_ TruncateStatement) {}

func (_ noopVisitor) VisitWith(_ WithStatement) {}

func (_ noopVisitor) VisitCte(_ CteStatement) {}

func (_ noopVisitor) VisitMerge(_ MergeStatement) {}

func (_ noopVisitor) VisitMatch(_ MatchStatement) {}

func (_ noopVisitor) VisitCall(_ CallStatement) {}

func (_ noopVisitor) VisitGrant(_ GrantStatement) {}

func (_ noopVisitor) VisitRevoke(_ RevokeStatement) {}

func (_ noopVisitor) VisitCommit(_ Commit) {}

func (_ noopVisitor) VisitRollback(_ Rollback) {}

func (_ noopVisitor) VisitSetTransaction(_ SetTransaction) {}

func (_ noopVisitor) VisitStartTransaction(_ StartTransaction) {}

func (_ noopVisitor) VisitSavepoint(_ Savepoint) {}

func (_ noopVisitor) VisitReleaseSavepoint(_ ReleaseSavepoint) {}

func (_ noopVisitor) VisitRollbackSavepoint(_ RollbackSavepoint) {}

func (_ noopVisitor) VisitJoin(_ Join) {}

func (_ noopVisitor) VisitOrder(_ Order) {}

func (_ noopVisitor) VisitLimit(_ Limit) {}

func (_ noopVisitor) VisitOffset(_ Offset) {}

func (_ noopVisitor) VisitBinary(_ Binary) {}

func (_ noopVisitor) VisitUnary(_ Unary) {}

func (_ noopVisitor) VisitCallFunc(_ Call) {}

func (_ noopVisitor) VisitList(_ List) {}

func (_ noopVisitor) VisitCollate(_ Collate) {}

func (_ noopVisitor) VisitIn(_ In) {}

func (_ noopVisitor) VisitIs(_ Is) {}

func (_ noopVisitor) VisitExists(_ Exists) {}

func (_ noopVisitor) VisitBetween(_ Between) {}

func (_ noopVisitor) VisitAll(_ All) {}

func (_ noopVisitor) VisitAny(_ Any) {}

func (_ noopVisitor) VisitNot(_ Not) {}

func (_ noopVisitor) VisitCast(_ Cast) {}

func (_ noopVisitor) VisitValue(_ Value) {}

func (_ noopVisitor) VisitAlias(_ Alias) {}

func (_ noopVisitor) VisitName(_ Name) {}

func (_ noopVisitor) VisitGroup(_ Group) {}

func (_ noopVisitor) VisitAssignment(_ Assignment) {}

func (_ noopVisitor) VisitBody(_ Body) {}

func (_ noopVisitor) VisitIf(_ If) {}

func (_ noopVisitor) VisitWhile(_ While) {}

func (_ noopVisitor) VisitSet(_ Set) {}

func (_ noopVisitor) VisitDeclare(_ Declare) {}

func (_ noopVisitor) VisitReturn(_ Return) {}

func (_ noopVisitor) VisitCase(_ Case) {}

func (_ noopVisitor) VisitWhen(_ When) {}

func (_ noopVisitor) VisitCreateProcedure(CreateProcedureStatement) {}

func (_ noopVisitor) VisitCreateTable(_ CreateTableStatement) {}

func (_ noopVisitor) VisitDropTable(_ DropTableStatement) {}

func (_ noopVisitor) VisitAlterTable(_ AlterTableStatement) {}

func (_ noopVisitor) VisitCreateView(_ CreateViewStatement) {}

func (_ noopVisitor) VisitDropView(_ DropViewStatement) {}

func (_ noopVisitor) VisitColumnDef(_ ColumnDef) {}

func (_ noopVisitor) VisitAddColumn(_ AddColumnAction) {}

func (_ noopVisitor) VisitAlterColumn(_ AlterColumnAction) {}

func (_ noopVisitor) VisitDropColumn(_ DropColumnAction) {}

func (_ noopVisitor) VisitAddConstraint(_ AddConstraintAction) {}

func (_ noopVisitor) VisitDropConstraint(_ DropConstraintAction) {}

func (_ noopVisitor) VisitConstraint(_ Constraint) {}

func (_ noopVisitor) VisitPrimaryKey(_ PrimaryKeyConstraint) {}

func (_ noopVisitor) VisitForeignKey(_ ForeignKeyConstraint) {}

func (_ noopVisitor) VisitNotNull(_ NotNullConstraint) {}

func (_ noopVisitor) VisitUnique(_ UniqueConstraint) {}

func (_ noopVisitor) VisitCheck(_ CheckConstraint) {}

func (_ noopVisitor) VisitDefault(_ DefaultConstraint) {}

func (_ noopVisitor) VisitGenerated(_ GeneratedConstraint) {}

func (_ noopVisitor) VisitXmlElement(_ XmlElement) {}

func (_ noopVisitor) VisitXmlAttribute(_ XmlAttribute) {}

func (_ noopVisitor) VisitXmlNamespace(_ XmlNamespace) {}

func (_ noopVisitor) VisitXmlText(_ XmlText) {}

func (_ noopVisitor) VisitXmlComment(_ XmlComment) {}

func (_ noopVisitor) VisitXmlAgg(_ XmlAgg) {}
