// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package shape describes the structure of a payload without its values: key
// and element names, value kinds and protobuf field numbers. It is meant for
// logging payloads that code couldn't handle, when the payload itself can't be
// logged because it carries message content or credentials.
//
// A name is only listed when it looks like a schema name rather than data: at
// most 64 ASCII letters, digits, '_', '-', '.' and '$', with no more than four
// digits, and split on those separators into words that read like identifiers.
// A word may be lowercase, capitalized camel case with fewer capitals than half
// its lowercase letters, or all capitals, and a digit may only be followed by
// a capital that starts a new word, except in short lowercase words like
// "e2ee". This rejects nearly all random tokens over 16 characters, but short
// tokens, usernames and file names can still pass, so it doesn't make it safe
// to describe payloads whose keys are mostly user data. Rejected names, such as
// user IDs, phone numbers or message IDs used as keys, are counted in a final
// "N other" entry and their subtree is skipped. Results are sorted and
// deduplicated, so equal shapes log identically. At most 64 names are listed;
// the rest are counted in "N other" too.
//
// Use the redact package instead when values need to be correlated between log
// lines.
package shape

import (
	"fmt"
	"maps"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const (
	maxNames       = 64
	maxTracked     = 1024
	maxNameLen     = 64
	maxNameDigits  = 4
	maxShortWord   = 5
	maxSchemeLen   = 16
	nameSeparators = "_-.$"
)

func isSafeName(name string) bool {
	if len(name) == 0 || len(name) > maxNameLen {
		return false
	}
	digits := 0
	for i := 0; i < len(name); i++ {
		switch c := name[i]; {
		case isDigit(c):
			digits++
		case isLower(c), isUpper(c), strings.IndexByte(nameSeparators, c) >= 0:
		default:
			return false
		}
	}
	if digits > maxNameDigits {
		return false
	}
	for word := range strings.FieldsFuncSeq(name, func(r rune) bool { return strings.ContainsRune(nameSeparators, r) }) {
		if !isSafeWord(word) {
			return false
		}
	}
	return true
}

func isSafeWord(word string) bool {
	upper, lower, badDigit := 0, 0, false
	for i := 0; i < len(word); i++ {
		switch c := word[i]; {
		case isUpper(c):
			upper++
		case isLower(c):
			lower++
		case i+1 < len(word):
			next := word[i+1]
			badDigit = badDigit || isLower(next) || (isUpper(next) && i > 0 && isUpper(word[i-1]))
		}
	}
	if upper == 0 && len(word) <= maxShortWord {
		return true
	}
	return !badDigit && (lower == 0 || upper*2 <= lower+1)
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isLower(c byte) bool { return c >= 'a' && c <= 'z' }
func isUpper(c byte) bool { return c >= 'A' && c <= 'Z' }

type nameSet struct {
	names map[string]struct{}
	other map[string]struct{}
	extra int
}

func (s *nameSet) add(name string) {
	if s.names == nil {
		s.names = make(map[string]struct{})
	}
	if _, ok := s.names[name]; ok {
		return
	} else if len(s.names) >= maxTracked {
		s.addOther(name)
	} else {
		s.names[name] = struct{}{}
	}
}

func (s *nameSet) addOther(raw string) {
	if s.other == nil {
		s.other = make(map[string]struct{})
	}
	if _, ok := s.other[raw]; ok {
		return
	} else if len(s.other) >= maxTracked {
		s.extra++
	} else {
		s.other[raw] = struct{}{}
	}
}

func (s *nameSet) result() []string {
	out := slices.Sorted(maps.Keys(s.names))
	other := len(s.other) + s.extra
	if len(out) > maxNames {
		other += len(out) - maxNames
		out = out[:maxNames]
	}
	if other > 0 {
		out = append(out, fmt.Sprintf("%d other", other))
	}
	return out
}

// Keys returns the sorted keys of a map, such as a [net/http.Header], a
// [url.Values] form or a decoded JSON object.
func Keys[M ~map[K]V, K ~string, V any](m M) []string {
	safe := make([]string, 0, len(m))
	for key := range m {
		if isSafeName(string(key)) {
			safe = append(safe, string(key))
		}
	}
	slices.Sort(safe)
	if len(safe) > maxNames {
		safe = safe[:maxNames]
	}
	if other := len(m) - len(safe); other > 0 {
		safe = append(safe, fmt.Sprintf("%d other", other))
	}
	return safe
}

// URL describes a URL as its scheme, host, port, path and sorted query keys,
// e.g. "https://example.com/l/{22}?lang&uuid". Path segments, query keys, host
// labels and schemes that don't look like schema names are replaced by their
// length in braces, as are the opaque part and the fragment. IP addresses
// become "{ip}" and user info becomes "{userinfo}".
func URL(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Sprintf("invalid (%d bytes)", len(raw))
	}
	var sb strings.Builder
	if parsed.Scheme != "" {
		if isSafeScheme(parsed.Scheme) {
			sb.WriteString(parsed.Scheme)
		} else {
			writeLength(&sb, parsed.Scheme)
		}
		sb.WriteByte(':')
	}
	if parsed.Opaque != "" {
		writeLength(&sb, parsed.Opaque)
	} else if (parsed.Scheme != "" || parsed.Host != "" || parsed.User != nil) &&
		(!parsed.OmitHost || parsed.Host != "" || parsed.User != nil) &&
		(parsed.Host != "" || parsed.Path != "" || parsed.User != nil) {
		sb.WriteString("//")
		if parsed.User != nil {
			sb.WriteString("{userinfo}@")
		}
		writeHost(&sb, parsed.Hostname())
		if port := parsed.Port(); port != "" {
			sb.WriteByte(':')
			sb.WriteString(port)
		}
	}
	for i, segment := range strings.Split(parsed.EscapedPath(), "/") {
		if i > 0 {
			sb.WriteByte('/')
		}
		if decoded, err := url.PathUnescape(segment); err == nil && (decoded == "" || isSafeName(decoded)) {
			sb.WriteString(decoded)
		} else {
			writeLength(&sb, segment)
		}
	}
	if parsed.RawQuery != "" {
		query, _ := url.ParseQuery(parsed.RawQuery)
		keys := make([]string, 0, len(query))
		for key := range query {
			if isSafeName(key) {
				keys = append(keys, key)
			} else {
				keys = append(keys, "{"+strconv.Itoa(len(key))+"}")
			}
		}
		slices.Sort(keys)
		sb.WriteByte('?')
		sb.WriteString(strings.Join(slices.Compact(keys), "&"))
	}
	if parsed.Fragment != "" {
		sb.WriteByte('#')
		writeLength(&sb, parsed.Fragment)
	}
	return sb.String()
}

func isSafeScheme(scheme string) bool {
	if len(scheme) > maxSchemeLen {
		return false
	}
	for i := 0; i < len(scheme); i++ {
		if c := scheme[i]; !isLower(c) && !(i > 0 && (isDigit(c) || c == '+' || c == '-')) {
			return false
		}
	}
	return true
}

func writeHost(sb *strings.Builder, host string) {
	if host == "" {
		return
	}
	ip, _, _ := strings.Cut(host, "%")
	if net.ParseIP(ip) != nil {
		sb.WriteString("{ip}")
		return
	}
	for i, label := range strings.Split(host, ".") {
		if i > 0 {
			sb.WriteByte('.')
		}
		if label == "" || isSafeName(label) {
			sb.WriteString(label)
		} else {
			writeLength(sb, label)
		}
	}
}

func writeLength(sb *strings.Builder, s string) {
	sb.WriteByte('{')
	sb.WriteString(strconv.Itoa(len(s)))
	sb.WriteByte('}')
}
