package lint

import (
	"strings"
	"testing"
)

type TestCase struct {
	Query  string
	Issues int
}

func runTests(t *testing.T, tests []TestCase, rule Rule) {
	t.Helper()
	for _, c := range tests {
		issues, err := Lint(strings.NewReader(c.Query), []Rule{rule})
		if err != nil {
			continue
		}
		if len(issues) != c.Issues {
			t.Errorf("[%s]: want %d issues! got %d", rule.Name(), c.Issues, len(issues))
		}
	}
}
