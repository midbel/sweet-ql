package lint

import (
	"testing"
)

func TestNoCte(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "select * from foobar",
			Issues: 0,
		},
		{
			Query:  "with foo as (select * from foo) select * from foo",
			Issues: 1,
		},
	}
	runTests(t, tests, NoCte(Warning))
}

func TestCteUnused(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "with foo as (select * from foobar) select * from foo",
			Issues: 0,
		},
		{
			Query:  "with foo as (select * from foobar) select * from tests t join foo f on t.id=f.id",
			Issues: 0,
		},
		{
			Query:  "with foo as (select * from foobar) select * from tests t where active in (select id from foo)",
			Issues: 0,
		},
		{
			Query:  "with foo as (select * from foobar) select * from tests",
			Issues: 1,
		},
		{
			Query:  "with foo as (select * from foobar), bar as (select * from barista) select * from tests",
			Issues: 2,
		},
	}
	runTests(t, tests, CteUnused(Warning))
}

func TestCteShadow(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "with foo as (select * from foobar) select * from foo",
			Issues: 0,
		},
		{
			Query:  "with foo as (select * from foo) select * from foo",
			Issues: 1,
		},
		{
			Query:  "with foo as (select f.* from foo f) select * from foo",
			Issues: 1,
		},
	}
	runTests(t, tests, CteShadow(Warning))
}

func TestCteNames(t *testing.T) {
	tests := []TestCase{
		{
			Query:  "with foo as (select * from foobar) select id, name from foo",
			Issues: 0,
		},
		{
			Query:  "with foo as (select id, name from foobar) select * from foo",
			Issues: 0,
		},
		{
			Query:  "with foo as (select id, name from foobar) select id, name, active from foo",
			Issues: 1,
		},
	}
	runTests(t, tests, CteNames(Warning))
}
