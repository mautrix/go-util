package glob_test

import (
	"strings"
	"testing"

	"go.mau.fi/util/glob"
)

var simplifyTests = []struct {
	input  string
	output string
}{
	{"", ""},
	{"a", "a"},
	{"*", "*"},
	{"**", "*"},
	{strings.Repeat("*", 9999), "*"},
	{"a*", "a*"},
	{"a**", "a*"},
	{"a**b", "a*b"},
	{"a**b**", "a*b*"},
	{"a**b**c", "a*b*c"},
	{"*?*", "?*"},
	{"**????***", "????*"},
	{"**?*?***?*?***", "????*"},
	{"meow**?*?***?*?***hmm**?*?***?*?***asd**?*?***?*?***", "meow????*hmm????*asd????*"},
}

func TestSimplify(t *testing.T) {
	for _, test := range simplifyTests {
		if got := glob.Simplify(test.input); got != test.output {
			t.Errorf("Simplify(%q) = %q; want %q", test.input, got, test.output)
		}
	}
}

func simplifyBySplitting(input string) string {
	parts := glob.SplitPattern(input)
	for i, part := range parts {
		if strings.ContainsRune(part, '*') {
			parts[i] = strings.Repeat("?", strings.Count(part, "?")) + "*"
		}
	}
	return strings.Join(parts, "")
}

func FuzzSimplify(f *testing.F) {
	f.Add(simplifyTests[0].input)
	f.Fuzz(func(t *testing.T, input string) {
		simplified := glob.Simplify(input)
		simplifiedBySplitting := simplifyBySplitting(input)
		if simplified != simplifiedBySplitting {
			t.Errorf("Simplify(%q) = %q; want %q", input, simplified, simplifiedBySplitting)
		}
	})
}
