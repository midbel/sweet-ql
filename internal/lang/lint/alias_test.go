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
	tests := []TestCase{
		{
			Query:  "select f.id as ident from foobar f",
			Issues: 0,
		},
		{
			Query:  "select id as ident from foobar",
			Issues: 1,
		},
		{
			Query:  "select id, name from foobar",
			Issues: 3,
		},
		{
			Query:  "with foo as (select id from foo) select * from foo",
			Issues: 4,
		},
		{
			Query:  "select id from foo where active=(select active from bar)",
			Issues: 4,
		},
	}
	runTests(t, tests, MissingAlias(Warning))
}

func TestNoAlias(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, name from foobar",
			Issues: 0,
		},
		{
			Query:  "with foo as (select id from foo) select * from foo",
			Issues: 0,
		},
		{
			Query:  "select name as foo from foobar",
			Issues: 1,
		},
		{
			Query:  "select f.name as foo from foobar f",
			Issues: 2,
		},
		{
			Query:  "with foo as (select id as i from foo) select * from foo",
			Issues: 1,
		},
	}
	runTests(t, tests, NoAlias(Warning))
}

func TestInvalidAlias(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select foo f from foobar",
			Issues: 0,
		},
		{
			Query:  "select foo f from foobar where f.active is true",
			Issues: 1,
		},
	}
	runTests(t, tests, InvalidAlias(Warning))
}

func TestUndefinedAlias(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select f.id from foobar f",
			Issues: 0,
		},
		{
			Query:  "select b.id from foobar f",
			Issues: 1,
		},
	}
	runTests(t, tests, UndefinedAlias(Warning))
}

func TestUnusedAlias(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select id, name from foobar",
			Issues: 0,
		},
		{
			Query:  "select foo.id, foo.name from foobar foo",
			Issues: 0,
		},
		{
			Query:  "select id, name from foobar foo where foo.active is true",
			Issues: 0,
		},
		{
			Query:  "select id, name from foo f join bar b on f.id=b.id",
			Issues: 0,
		},
		{
			Query:  "select id, name from foobar foo",
			Issues: 1,
		},
	}
	runTests(t, tests, UnusedAlias(Warning))
}
