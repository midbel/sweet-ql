package ast

type VisitableNode interface {
	Accept(Visitor)
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

	VisitMatch(MatchStatement)
	VisitJoin(Join)
	VisitOrder(Order)
	VisitLimit(Limit)
	VisitOffset(Offset)
	VisitReturn(Return)
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
	VisitCall(Call)

	VisitValue(Value)
	VisitAlias(Alias)
	VisitName(Name)
	VisitGroup(Group)
	VisitCase(Case)
	VisitWhen(When)
	VisitAssignment(Assignment)
}

type XmlVisitor interface{}

type Visitor interface {
	StmtVisitor
	ExprVisitor
	XmlVisitor
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

func (_ noopVisitor) VisitJoin(_ Join) {}

func (_ noopVisitor) VisitOrder(_ Order) {}

func (_ noopVisitor) VisitLimit(_ Limit) {}

func (_ noopVisitor) VisitOffset(_ Offset) {}

func (_ noopVisitor) VisitBinary(_ Binary) {}

func (_ noopVisitor) VisitUnary(_ Unary) {}

func (_ noopVisitor) VisitCall(_ Call) {}

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

func (_ noopVisitor) VisitCase(_ Case) {}

func (_ noopVisitor) VisitWhen(_ When) {}

func (_ noopVisitor) VisitReturn(_ Return) {}

func (_ noopVisitor) VisitAssignment(_ Assignment) {}
