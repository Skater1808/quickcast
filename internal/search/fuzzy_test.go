package search

import (
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		pattern string
		target  string
		matched bool
	}{
		{"code", "Visual Studio Code", true},
		{"vsc", "Visual Studio Code", true},
		{"term", "Alacritty Terminal", true},
		{"ff", "Firefox", true},
		{"xyz", "Firefox", false},
		{"", "Anything", true},
	}

	for _, c := range cases {
		res := FuzzyMatch(c.pattern, c.target)
		if res.Matched != c.matched {
			t.Errorf("FuzzyMatch(%q, %q) got matched=%v, expected=%v", c.pattern, c.target, res.Matched, c.matched)
		}
	}
}
