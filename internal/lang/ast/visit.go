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
	VisiteTruncate(TruncateStatement)
	VisitCall(CallStatement)
}

type ExprVisitor interface {
	VisitBinary(Binary)
	VisitUnary(Unary)
	VisitIn(In)
	VisitIs(Is)
	VisitBetween(Between)
	VisitAll(All)
	VisitAny(Any)
	VisitNot(Not)
	VisitCast(Cast)

	VisitAlias(Alias)
	VisitName(Name)
	VisitGroup(Group)
}

type Visitor interface {
	StmtVisitor
	ExprVisitor
}
