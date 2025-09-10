package lint

import (
	"testing"
)

func TestGroupbyColumns(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, nae from foobar",
			Issues: 0,
		},
		{
			Query:  "select name, count(*) from foobar group by name",
			Issues: 0,
		},
		{
			Query:  "select name as grp, count(*) from foobar group by name",
			Issues: 0,
		},
		{
			Query:  "select name, dtstart, dtend, count(*) from foobar group by name",
			Issues: 2,
		},
		{
			Query:  "select name, isBetween(dtstart, dtend) from foobar group by name",
			Issues: 1,
		},
	}
	runTests(t, tests, GroupbyColumns(Warning))
}

func TestGroupbyDistinct(t *testing.T) {
	tests := []TestCase{}
	runTests(t, tests, GroupbyDistinct(Warning))
}

func TestGroupByNoPosition(t *testing.T) {
	tests := []TestCase{}
	runTests(t, tests, NoPositionGroupby(Warning))
}

func TestGroupByNoLiteral(t *testing.T) {
	tests := []TestCase{}
	runTests(t, tests, NoLiteralGroupby(Warning))
}

func TestGroupbyAggr(t *testing.T) {
	tests := []TestCase{}
	runTests(t, tests, GroupbyAggrFunc(Warning))
}
