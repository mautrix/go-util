// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package redact anonymises arbitrary JSON-shaped values so payloads can be
// logged for debugging without leaking their contents. Opaque values are
// replaced with a marker that preserves their length and guessed type plus a
// keyed hash, so the shape of a payload survives and equal values correlate
// while the data cannot be recovered. The key is random per process, so
// markers do not correlate across restarts.
//
// What is kept versus hashed is governed entirely by a [Policy] value, making
// each policy auditable at a glance. [MakeDefaultPolicy] hashes everything that
// is not structural. Callers that need more readable output relax the Keep
// fields; [Policy.HashPatterns] still catches PII inside whatever they keep.
package redact

import (
	"bytes"
	"cmp"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"go.mau.fi/util/random"
)

// Decision is the fate of a single value, returned by the [Policy.ClassifyKey]
// and [Policy.ClassifyString] hooks.
type Decision int

const (
	// Auto defers to the policy's built-in heuristics.
	Auto Decision = iota
	// Keep skips the heuristics and passes the value through. HashPatterns
	// still apply to it and to any strings nested within it.
	Keep
	// Hash always redacts the value.
	Hash
)

// Policy controls which values are kept and which are redacted. The zero
// value fails closed: it keeps nothing on prose grounds and hashes every
// non-zero number.
type Policy struct {
	// KeepProse keeps strings that read as natural language or as an
	// attribute name (kebab/snake identifiers).
	KeepProse bool
	// MaxDigits is the most digits a number may have, ignoring sign, decimal
	// point and leading zeros, and still be kept verbatim. Longer numbers, and
	// any in exponent notation, are replaced with a marker. Ten keeps
	// second-precision unix timestamps.
	MaxDigits int
	// KeepPatterns keeps any string matching one of these, checked before the
	// prose heuristics. Use it for internal identifiers.
	KeepPatterns []*regexp.Regexp
	// KeepURLPathPatterns keeps the path of an http(s) URL when it matches one
	// of these. The scheme and host of a URL are always kept.
	KeepURLPathPatterns []*regexp.Regexp
	// HashPatterns redacts every match of these inside a string that is
	// otherwise kept, whichever rule kept it, so PII embedded in prose and
	// allowlisted values is still hashed. Each match is replaced in place and
	// the surrounding text survives. Strings hashed whole are unaffected.
	HashPatterns []*regexp.Regexp
	// SensitiveKeys always redacts the value under a matching object key or
	// HTML attribute name, whatever its type. Keys are matched
	// case-insensitively.
	SensitiveKeys map[string]bool

	// ClassifyKey overrides the fate of the value under an object key or HTML
	// attribute name. A return of Auto falls through to SensitiveKeys and the
	// value's own heuristics.
	ClassifyKey func(key string) Decision
	// ClassifyString overrides the fate of a string value. A return of Auto
	// falls through to the built-in heuristics.
	ClassifyString func(value string) Decision
}

// Patterns for [Policy.HashPatterns]. They favour catching PII over precision:
// PhoneNumberPattern also matches card numbers and separated dates, and
// IPv4Pattern matches dotted version strings.
var (
	EmailPattern       = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	PhoneNumberPattern = regexp.MustCompile(`\+?\(?\d(?:[ .()-]{0,2}\d){7,}`)
	DigitRunPattern    = regexp.MustCompile(`\d{6,}`)
	IPv4Pattern        = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
)

// MakeDefaultPolicy returns a policy that hashes everything not structural: no
// string is kept on prose grounds, numbers longer than four digits are hashed,
// values under common secret/PII keys are always redacted, and PII-shaped text
// is hashed inside anything that is kept.
// Each call returns independent maps and slices, so the result can be tuned
// freely.
func MakeDefaultPolicy() Policy {
	return Policy{
		MaxDigits:     4,
		HashPatterns:  []*regexp.Regexp{EmailPattern, PhoneNumberPattern, DigitRunPattern, IPv4Pattern},
		SensitiveKeys: defaultSensitiveKeys(),
	}
}

func defaultSensitiveKeys() map[string]bool {
	keys := map[string]bool{}
	for _, k := range []string{
		"email", "email_address", "phone", "phone_number", "phonenumber",
		"username", "first_name", "last_name", "full_name", "display_name",
		"password", "pass", "passwd", "secret", "client_secret",
		"token", "access_token", "refresh_token", "session_token", "id_token",
		"auth_token", "oauth_token", "csrf_token", "fb_dtsg", "lsd",
		"session", "session_id", "sessionid", "session_key",
		"cookie", "cookies", "set_cookie", "xs",
		"api_key", "apikey", "authorization", "auth",
		"security_code", "verification_code", "confirmation_code",
		"otp", "pin", "ssn", "credit_card", "card_number", "cvv",
	} {
		keys[k] = true
	}
	return keys
}

// JSON redacts a JSON document and returns the re-marshaled result. Numbers
// are decoded as [json.Number] so their exact text survives redaction.
func (p Policy) JSON(data []byte) ([]byte, error) {
	v, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("parsing payload for redaction: %w", err)
	}
	p.Value(&v)
	out, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshaling redacted payload: %w", err)
	}
	return out, nil
}

func decode(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("trailing data after JSON value")
	}
	return v, nil
}

// Value redacts an already-parsed JSON value in place. Strings, numbers
// ([json.Number], int64 or float64), bools and nil are handled as leaves,
// []any and map[string]any are walked, and a value of any other type is
// hashed whole.
func (p Policy) Value(ptr *any) {
	switch val := (*ptr).(type) {
	case nil, bool:
	case string:
		*ptr = p.String(val)
	case json.Number:
		*ptr = p.number(string(val), val)
	case int64:
		*ptr = p.number(strconv.FormatInt(val, 10), val)
	case float64:
		*ptr = p.number(strconv.FormatFloat(val, 'f', -1, 64), val)
	case []any:
		for idx := range val {
			p.Value(&val[idx])
		}
	case map[string]any:
		for key, child := range val {
			val[key] = p.keyed(key, child)
		}
	default:
		*ptr = p.forceRedact(val)
	}
}

func (p Policy) number(text string, num any) any {
	significant := strings.TrimLeft(strings.TrimPrefix(text, "-"), "0.")
	if strings.ContainsAny(text, "eE") || len(significant)-strings.Count(significant, ".") > p.MaxDigits {
		return hashMarker(text)
	}
	return num
}

func (p Policy) keyed(key string, v any) any {
	switch p.keyDecision(key) {
	case Keep:
		return p.mask(v)
	case Hash:
		return p.forceRedact(v)
	default:
		p.Value(&v)
		return v
	}
}

func (p Policy) keyDecision(key string) Decision {
	if p.ClassifyKey != nil {
		if d := p.ClassifyKey(key); d != Auto {
			return d
		}
	}
	if p.SensitiveKeys[strings.ToLower(key)] {
		return Hash
	}
	return Auto
}

func (p Policy) forceRedact(v any) any {
	switch val := v.(type) {
	case nil, bool:
		return val
	case string:
		return hashMarker(val)
	default:
		marshaled, err := json.Marshal(v)
		if err != nil {
			return hashMarker(fmt.Sprint(v))
		}
		return hashMarker(string(marshaled))
	}
}

func (p Policy) mask(v any) any {
	switch val := v.(type) {
	case string:
		return p.maskPatterns(val)
	case []any:
		for idx := range val {
			val[idx] = p.mask(val[idx])
		}
	case map[string]any:
		for key, child := range val {
			val[key] = p.mask(child)
		}
	}
	return v
}

var (
	marker    = regexp.MustCompile(`^redacted_\d+char_[a-z_]+_[0-9a-f]{8}$`)
	likelyURL = regexp.MustCompile(`^([a-z]{2,10}://[a-z0-9.-]+/)(.+)$`)
)

// String redacts a single string value.
func (p Policy) String(s string) string {
	if marker.MatchString(s) {
		return s
	}
	if p.ClassifyString != nil {
		switch p.ClassifyString(s) {
		case Keep:
			return p.maskPatterns(s)
		case Hash:
			return hashMarker(s)
		}
	}
	if trimmed := strings.TrimLeft(s, " \t\r\n"); strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if obj, err := decode([]byte(s)); err == nil {
			p.Value(&obj)
			if out, err := json.Marshal(obj); err == nil {
				return string(out)
			}
		}
	}
	if p.keepsWhole(s) {
		return p.maskPatterns(s)
	}
	prefix := ""
	if match := likelyURL.FindStringSubmatch(s); match != nil {
		prefix = match[1]
		s = match[2]
	}
	for _, pattern := range p.KeepURLPathPatterns {
		if pattern.MatchString(s) {
			return prefix + p.maskPatterns(s)
		}
	}
	return prefix + hashMarker(s)
}

func (p Policy) keepsWhole(s string) bool {
	for _, pattern := range p.KeepPatterns {
		if pattern.MatchString(s) {
			return true
		}
	}
	return p.KeepProse && (likelyAttributeName(s) || likelyEnglishSentence(s))
}

func (p Policy) maskPatterns(s string) string {
	var spans [][]int
	for _, pattern := range p.HashPatterns {
		for _, span := range pattern.FindAllStringIndex(s, -1) {
			if span[0] < span[1] {
				spans = append(spans, span)
			}
		}
	}
	if len(spans) == 0 {
		return s
	}
	slices.SortFunc(spans, func(a, b []int) int { return cmp.Compare(a[0], b[0]) })
	merged := [][]int{spans[0]}
	for _, span := range spans[1:] {
		if last := merged[len(merged)-1]; span[0] <= last[1] {
			last[1] = max(last[1], span[1])
		} else {
			merged = append(merged, span)
		}
	}
	var out strings.Builder
	end := 0
	for _, span := range merged {
		out.WriteString(s[end:span[0]])
		out.WriteString(hashMarker(s[span[0]:span[1]]))
		end = span[1]
	}
	out.WriteString(s[end:])
	return out.String()
}

var stringTypes = []struct {
	name  string
	regex *regexp.Regexp
}{
	{"uuid_lowercase", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)},
	{"uuid_uppercase", regexp.MustCompile(`^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$`)},
	{"hex", regexp.MustCompile(`^[0-9a-f]*[a-f][0-9a-f]*$`)},
	{"number", regexp.MustCompile(`^[0-9]+$`)},
}

var hashKey = random.Bytes(32)

func hashMarker(s string) string {
	stringType := "str"
	for _, guess := range stringTypes {
		if guess.regex.MatchString(s) {
			stringType = guess.name
			break
		}
	}
	mac := hmac.New(sha256.New, hashKey)
	mac.Write([]byte(s))
	return fmt.Sprintf("redacted_%dchar_%s_%s", len(s), stringType, hex.EncodeToString(mac.Sum(nil)[:4]))
}

var (
	englishWord           = regexp.MustCompile(`([A-Za-z0-9'-]+)[,.:;]? *`)
	digitRegexp           = regexp.MustCompile(`[0-9]`)
	lowercaseLetterRegexp = regexp.MustCompile(`[a-z]`)
	uppercaseLetterRegexp = regexp.MustCompile(`[A-Z]`)
	alnumsRegexp          = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

func likelyEnglishSentence(s string) bool {
	if len(s) == 0 {
		return true
	}
	words := englishWord.FindAllStringSubmatch(s, -1)
	numReasonableChars := 0
	for _, wordMatch := range words {
		wordAndExtra := wordMatch[0]
		word := wordMatch[1]
		if len(word) > 24 {
			continue
		}
		if strings.Count(word, "-") > 1 {
			continue
		}
		if strings.Count(word, "'") > 1 {
			continue
		}
		if len(digitRegexp.FindAllString(word, -1)) > 6 {
			continue
		}
		if len(lowercaseLetterRegexp.FindAllString(word, -1)) > 2 && len(uppercaseLetterRegexp.FindAllString(word, -1)) > 2 {
			continue
		}
		numReasonableChars += len(wordAndExtra)
	}
	return numReasonableChars > 0 && (float64(numReasonableChars)/float64(len(s)) > 0.8 || len(s)-numReasonableChars < 10)
}

func likelyAttributeName(s string) bool {
	for _, delimiter := range []string{"-", "_"} {
		bad := false
		for part := range strings.SplitSeq(s, delimiter) {
			if len(part) > 24 {
				bad = true
			}
			if !alnumsRegexp.MatchString(part) {
				bad = true
			}
			if len(digitRegexp.FindAllString(part, -1)) > 2 {
				bad = true
			}
			if len(lowercaseLetterRegexp.FindAllString(part, -1)) > 2 && len(uppercaseLetterRegexp.FindAllString(part, -1)) > 2 {
				bad = true
			}
		}
		if !bad {
			return true
		}
	}
	return false
}
