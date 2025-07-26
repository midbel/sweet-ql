package lang

import (
	"slices"
	"strings"

	"github.com/midbel/sweet/internal/keywords"
	"github.com/midbel/sweet/internal/lang/ast"
)

type Formatter interface {
	Quote(string) string
}

type Parser interface {
	Parse() (ast.Node, error)
	Query() string
}

var AggregateFunctions = []string{
	"MAX",
	"MIN",
	"AVG",
	"SUM",
	"COUNT",
	"STDDEV_POP",
	"STDDEV_SAMP",
	"VAR_POP",
	"VAR_SAMP",
}

func IsAggregateFunc(ident string) bool {
	return slices.Contains(AggregateFunctions, strings.ToUpper(ident))
}

var BuiltinFunctions = []string{
	"MAX",
	"MIN",
	"AVG",
	"SUM",
	"COUNT",
	"STDDEV_POP",
	"STDDEV_SAMP",
	"VAR_POP",
	"VAR_SAMP",
	"UPPER",
	"LOWER",
	"CONCAT",
	"CHAR_LENGTH",
	"CHARACTER_LENGTH",
	"POSITION",
	"SUBSTR",
	"TRIM",
}

func IsBuiltinFunc(ident string) bool {
	return slices.Contains(BuiltinFunctions, strings.ToUpper(ident))
}

func ExpandKeyword(kw string) string {
	switch strings.ToUpper(kw) {
	case "JOIN":
		kw = "INNER JOIN"
	case "LEFT JOIN":
		kw = "LEFT OUTER JOIN"
	case "RIGHT JOIN":
		kw = "RIGHT OUTER JOIN"
	case "FULL JOIN":
		kw = "FULL OUTER JOIN"
	}
	return kw
}

func CompactKeyword(kw string) string {
	switch strings.ToUpper(kw) {
	case "INNER JOIN":
		kw = "JOIN"
	case "LEFT OUTER JOIN":
		kw = "LEFT JOIN"
	case "RIGHT OUTER JOIN":
		kw = "RIGHT JOIN"
	case "FULL OUTER JOIN":
		kw = "FULL JOIN"
	}
	return kw
}

var ansi = [][]string{
	{"create", "procedure"},
	{"create", "table"},
	{"create", "view"},
	{"declare"},
	{"default"},
	{"exists"},
	{"null"},
	{"select"},
	{"from"},
	{"where"},
	{"having"},
	{"limit"},
	{"offset"},
	{"fetch"},
	{"row"},
	{"rows"},
	{"next"},
	{"only"},
	{"group", "by"},
	{"order", "by"},
	{"as"},
	{"in"},
	{"inout"},
	{"out"},
	{"join"},
	{"on"},
	{"full", "join"},
	{"full", "outer", "join"},
	{"outer", "join"},
	{"left", "join"},
	{"left", "outer", "join"},
	{"right", "join"},
	{"right", "outer", "join"},
	{"inner", "join"},
	{"union"},
	{"intersect"},
	{"except"},
	{"all"},
	{"any"},
	{"distinct"},
	{"and"},
	{"or"},
	{"asc"},
	{"desc"},
	{"nulls"},
	{"first"},
	{"last"},
	{"similar"},
	{"like"},
	{"ilike"},
	{"delete"},
	{"delete", "from"},
	{"truncate"},
	{"truncate", "table"},
	{"update"},
	{"merge"},
	{"merge", "into"},
	{"when", "matched"},
	{"when", "not", "matched"},
	{"set"},
	{"insert"},
	{"insert", "into"},
	{"values"},
	{"case"},
	{"when"},
	{"then"},
	{"end"},
	{"using"},
	{"begin"},
	{"read", "write"},
	{"read", "only"},
	{"repeatable", "read"},
	{"read", "committed"},
	{"read", "uncommitted"},
	{"serializable"},
	{"isolation", "level"},
	{"start", "transaction"},
	{"set", "transaction"},
	{"savepoint"},
	{"release"},
	{"release", "savepoint"},
	{"rollback", "to", "savepoint"},
	{"commit"},
	{"rollback"},
	{"while"},
	{"end", "while"},
	{"do"},
	{"if"},
	{"end", "if"},
	{"else"},
	{"elseif"},
	{"with"},
	{"recursive"},
	{"materialized"},
	{"return"},
	{"is"},
	{"isnull"},
	{"notnull"},
	{"not"},
	{"collate"},
	{"between"},
	{"cast"},
	{"filter"},
	{"window"},
	{"over"},
	{"partition", "by"},
	{"range"},
	{"groups"},
	{"preceding"},
	{"following"},
	{"unbounded", "preceding"},
	{"unbounded", "following"},
	{"current", "row"},
	{"exclude", "no", "others"},
	{"exclude", "current", "row"},
	{"exclude", "group"},
	{"exclude", "ties"},
	{"call"},
	{"constraint"},
	{"primary", "key"},
	{"foreign", "key"},
	{"references"},
	{"autoincrement"},
	{"unique"},
	{"check"},
	{"generated"},
	{"generated", "always"},
	{"generated", "by", "default"},
	{"language"},
	{"alter", "table"},
	{"alter"},
	{"alter", "column"},
	{"add"},
	{"add", "column"},
	{"add", "constraint"},
	{"drop"},
	{"drop", "table"},
	{"drop", "view"},
	{"drop", "column"},
	{"drop", "constraint"},
	{"to"},
	{"true"},
	{"false"},
	{"unknown"},
	{"cascade"},
	{"restrict"},
	{"restart", "identity"},
	{"continue", "identity"},
	{"grant"},
	{"with", "grant", "option"},
	{"revoke"},
	{"all", "privileges"},
}

func GetKeywords() *keywords.Trie {
	return keywords.NewTrieFrom(ansi)
}
