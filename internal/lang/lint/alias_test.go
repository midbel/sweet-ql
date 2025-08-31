package lint

import (
	"testing"
)

func TestSelfAlias(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id from foobar",
			Issues: 0,
		},
		{
			Query:  "select id as ident from foobar",
			Issues: 0,
		},
		{
			Query:  "select id as id from foobar",
			Issues: 1,
		},
		{
			Query:  "select id as id, foo foo from foobar",
			Issues: 2,
		},
	}
	runTests(t, tests, SelfAlias(Warning))
}

func TestAmbiguousAlias(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select upper(foo) as name from foobar",
			Issues: 0,
		},
		{
			Query:  "select upper(foo) as upper from foobar",
			Issues: 1,
		},
	}
	runTests(t, tests, AmbiguousAlias(Warning))
}

func TestRecommandedAlias(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, foo as name from foobar",
			Issues: 0,
		},
		{
			Query:  "select upper(foo) as name from foobar",
			Issues: 0,
		},
		{
			Query:  "select id, xp + 2 from foobar",
			Issues: 1,
		},
		{
			Query:  "select id, xp + 2 as level from foobar",
			Issues: 0,
		},
	}
	runTests(t, tests, RecommandedAlias(Warning))
}

func TestMissingAlias(t *testing.T) {
	t.Skip()
}

func TestNoAlias(t *testing.T) {
	t.Skip()
}

func TestInvalidAlias(t *testing.T) {
	t.Skip()
}

func TestUndefinedAlias(t *testing.T) {
	t.Skip()
}
