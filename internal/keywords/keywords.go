package keywords

import (
	"errors"
	"slices"
	"sort"
	"strings"
)

var ErrFound = errors.New("keyword found")

type Set [][]string

func (ks Set) Merge(other Set) Set {
	return append(ks, other...)
}

func (ks Set) Len() int {
	return len(ks)
}

func (ks Set) Find(str string) int {
	return sort.Search(ks.Len(), func(i int) bool {
		return str <= ks[i][0]
	})
}

// Is check if the given str is a keyword. A keyword can be a standalone keyword
// or a compound keyword
// Is returns a string with the full SQL keyword, a first boolean as flag to indicate
// it the keyword is a standalone keyword and a final bool to indicate if the given str
// is a SQL keyword
func (ks Set) Is(str []string) (string, bool, bool) {
	var (
		n = ks.Len()
		s = strings.ToLower(str[0])
		i = ks.Find(s)
	)
	if i >= n || ks[i][0] != s {
		return "", false, false
	}

	if len(ks[i]) == 1 && len(str) == 1 && ((i+1 < n && ks[i+1][0] != s) || i+1 == n) {
		return s, true, true
	}
	var (
		got  = strings.ToLower(strings.Join(str, " "))
		want string
	)
	for _, kw := range ks[i:] {
		if kw[0] != s {
			break
		}
		want = strings.Join(kw, " ")
		switch {
		case want == got:
			var final bool
			if i+1 == n || !slices.Equal(str, ks[i+1]) {
				final = true
			}
			return got, final, true
		case strings.HasPrefix(want, got):
			return got, false, false
		default:
		}
	}
	return "", false, false
}

func (ks Set) Prepare() {
	seen := make(map[string]struct{})
	for i := range ks {
		str := strings.Join(ks[i], "")
		if _, ok := seen[str]; ok {
			continue
		}
		seen[str] = struct{}{}
		for j := range ks[i] {
			ks[i][j] = strings.ToLower(ks[i][j])
		}
	}
	sort.Slice(ks, func(i, j int) bool {
		fst := strings.Join(ks[i], " ")
		lst := strings.Join(ks[j], " ")
		return fst < lst
	})
}
