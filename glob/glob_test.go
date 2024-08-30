package glob_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"go.mau.fi/util/glob"
)

type mparams struct {
	input   string
	pattern string
}

type MatchTest struct {
	mparams
	result bool
}

var matchTests = []MatchTest{
	{mparams{"", ""}, true},
	{mparams{"", "a"}, false},
	{mparams{"a", ""}, false},
	{mparams{"a", "a"}, true},
	{mparams{"a", "b"}, false},
	{mparams{"a", "a*"}, true},
	{mparams{"a", "b*"}, false},
	{mparams{"a", "*a"}, true},
	{mparams{"a", "*b"}, false},
	{mparams{"a", "*"}, true},
	{mparams{"a", "a*?"}, false},
	{mparams{"ab", "a*?"}, true},
	{mparams{"a", "b*?"}, false},
	{mparams{"a", "*a?"}, false},
	{mparams{"ab", "*a?"}, true},
	{mparams{"a", "*b?"}, false},
	{mparams{"a", "*?"}, true},
	{mparams{"a", "a*?b"}, false},
	{mparams{"a", "a*?*"}, false},
	{mparams{"ab", "a*?*"}, true},
	{mparams{"a", "a*?*b"}, false},
	{mparams{"a", "a*?*a"}, false},
	{mparams{"aba", "a*?*a*"}, true},
	{mparams{"hellomeowworld", "*meow*"}, true},
	{mparams{"hellomeowworld", "hello*world"}, true},
	{mparams{"hellomeowworld", "hellomeow*"}, true},
	{mparams{"hellomeowworld", "*meowworld"}, true},
	{mparams{"meowwoofhmmdsa!asddfdfhg", "meow**?*?***?*?***hmm**?*?***?*?***asd**?*?***?*?***"}, true},
}

func TestCompile(t *testing.T) {
	for _, test := range matchTests {
		g := glob.Compile(test.pattern)
		if g.Match(test.input) != test.result {
			t.Errorf("Compile(%q).Match(%q) = %v; want %v", test.pattern, test.input, !test.result, test.result)
		}
	}
}

func BenchmarkCompile(b *testing.B) {
	for _, test := range matchTests {
		b.Run(test.pattern, func(b *testing.B) {
			g := glob.Compile(test.pattern)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.Match(test.input)
			}
		})
	}
}

func FuzzCompile(f *testing.F) {
	for _, test := range matchTests {
		f.Add(test.input, test.pattern)
	}
	f.Fuzz(func(t *testing.T, input, pattern string) {
		if !utf8.ValidString(pattern) || strings.ContainsRune(pattern, '\n') || strings.ContainsRune(input, '\n') {
			return
		}
		g := glob.CompileSimple(pattern)
		if g == nil {
			return
		}
		r, err := glob.CompileRegex(pattern)
		if err != nil {
			t.Fatalf("CompileRegex(%q) failed: %v", pattern, err)
		}
		if g.Match(input) != r.Match(input) {
			t.Errorf("Compile(%q).Match(%q) = %v; want %v", pattern, input, !r.Match(input), r.Match(input))
		}
	})
}
