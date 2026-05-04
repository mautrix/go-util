package glob

import (
	"strings"
)

// ExactGlob is the result of [Compile] when the pattern contains no glob characters.
// It uses a simple string comparison to match. The pattern is case-insensitive, Match uses EqualFold.
type ExactGlob string

func (eg ExactGlob) Match(s string) bool {
	return strings.EqualFold(string(eg), s)
}

// SuffixGlob is the result of [Compile] when the pattern only has one `*` at the beginning.
// It uses [strings.HasSuffix] to match. The pattern is case-insensitive, Match uses EqualFold.
type SuffixGlob string

func (sg SuffixGlob) Match(s string) bool {
	return len(s) >= len(sg) &&
		strings.EqualFold(string(sg), s[len(s)-len(sg):])
}

// PrefixGlob is the result of [Compile] when the pattern only has one `*` at the end.
// It uses [strings.HasPrefix] to match. The pattern is case-insensitive, Match uses EqualFold.
type PrefixGlob string

func (pg PrefixGlob) Match(s string) bool {
	return len(s) >= len(pg) &&
		strings.EqualFold(string(pg), s[:len(pg)])
}

// ContainsGlob is the result of [Compile] when the pattern has two `*`s, one at the beginning and one at the end.
// It uses [strings.Contains] to match. The pattern must be lowercased, Match will always lowercase the input.
//
// When there are exactly two `*`s, but they're not surrounding the string, the pattern is compiled as a [PrefixSuffixAndContainsGlob] instead.
type ContainsGlob string

func (cg ContainsGlob) Match(s string) bool {
	return strings.Contains(strings.ToLower(s), string(cg))
}

// PrefixAndSuffixGlob is the result of [Compile] when the pattern only has one `*` in the middle.
type PrefixAndSuffixGlob struct {
	Prefix string
	Suffix string
}

func (psg PrefixAndSuffixGlob) Match(s string) bool {
	return len(s) >= len(psg.Prefix)+len(psg.Suffix) &&
		strings.EqualFold(psg.Prefix, s[:len(psg.Prefix)]) &&
		strings.EqualFold(psg.Suffix, s[len(s)-len(psg.Suffix):])
}

// PrefixSuffixAndContainsGlob is the result of [Compile] when the pattern has two `*`s which are not surrounding the rest of the pattern.
// Each part of the pattern must be lowercased, Match will always lowercase the input.
type PrefixSuffixAndContainsGlob struct {
	Prefix   string
	Suffix   string
	Contains string
}

func (psacg PrefixSuffixAndContainsGlob) Match(s string) bool {
	s = strings.ToLower(s)
	return strings.HasPrefix(s, psacg.Prefix) &&
		strings.HasSuffix(s[len(psacg.Prefix):], psacg.Suffix) &&
		strings.Contains(s[len(psacg.Prefix):len(s)-len(psacg.Suffix)], psacg.Contains)
}
