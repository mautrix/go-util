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
// most 64 ASCII letters, digits, '_', '-', '.' and '$', with at least one letter
// and no more than four digits, and split on those separators into words that
// read like identifiers.
// In a camel case word, the capitals may be at most half its lowercase letters,
// rounded up, with each run of capitals counting as at most two. A word in all
// capitals may have at most 12 letters, and no digits unless it's at most five
// characters long. A digit may only be followed by a capital that starts a new
// word, except in short lowercase words like "e2ee". This rejects nearly all
// random tokens over 16 characters, but short tokens, usernames and file names
// can still pass, so it doesn't make it safe to describe payloads whose keys
// are mostly user data. Rejected names, such as user IDs, phone numbers or
// message IDs used as keys, are counted in a final "N other" entry and their
// subtree is skipped.
//
// Results are sorted and deduplicated, so equal shapes log identically. At most
// the first 64 names in sorted order are listed and the rest are counted in "N
// other" too. When more than 1024 distinct names are seen, the count becomes
// "960+ other".
//
// Use the redact package instead when values need to be correlated between log
// lines.
package shape

import (
	"fmt"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const (
	maxNames       = 64
	maxTracked     = 1024
	depthLimit     = 32
	maxNameLen     = 64
	maxNameDigits  = 4
	maxShortWord   = 5
	maxCapsWord    = 12
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
	if digits > maxNameDigits || strings.Trim(name, "0123456789"+nameSeparators) == "" {
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
	upper, lower, capitals, badDigit := 0, 0, 0, false
	for i := 0; i < len(word); {
		start := i
		switch c := word[i]; {
		case isUpper(c):
			for i < len(word) && isUpper(word[i]) {
				i++
			}
			upper += i - start
			capitals += min(i-start, 2)
		case isLower(c):
			lower++
			i++
		default:
			for i < len(word) && isDigit(word[i]) {
				i++
			}
			if i < len(word) && (isLower(word[i]) || (start > 0 && isUpper(word[start-1]) && isUpper(word[i]))) {
				badDigit = true
			}
		}
	}
	switch {
	case upper == 0 && len(word) <= maxShortWord:
		return true
	case badDigit:
		return false
	case lower == 0:
		return upper <= maxCapsWord && (len(word) <= maxShortWord || upper == len(word))
	default:
		return capitals*2 <= lower+1
	}
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isLower(c byte) bool { return c >= 'a' && c <= 'z' }
func isUpper(c byte) bool { return c >= 'A' && c <= 'Z' }

func clampDepth(depth int) int {
	return min(max(depth, 1), depthLimit)
}

type nameSet struct {
	seen     map[string]struct{}
	listed   []string
	overflow bool
}

func (s *nameSet) see(key string) {
	if _, ok := s.seen[key]; ok {
		return
	} else if len(s.seen) >= maxTracked {
		s.overflow = true
		return
	} else if s.seen == nil {
		s.seen = make(map[string]struct{})
	}
	s.seen[key] = struct{}{}
}

func (s *nameSet) add(name string) {
	s.see(name)
	idx, found := slices.BinarySearch(s.listed, name)
	if !found && idx < maxNames {
		s.listed = slices.Insert(s.listed, idx, name)
		if len(s.listed) > maxNames {
			s.listed = s.listed[:maxNames]
		}
	}
}

func (s *nameSet) addOther(raw string) {
	s.see("\x00" + raw)
}

func (s *nameSet) result() []string {
	out := slices.Clone(s.listed)
	if s.overflow {
		out = append(out, fmt.Sprintf("%d+ other", maxTracked-maxNames))
	} else if other := len(s.seen) - len(s.listed); other > 0 {
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
// become "{ip}" and user info becomes "{userinfo}". At most 64 path segments
// and 64 query keys are described.
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
		if i > maxNames {
			sb.WriteString("/...")
			break
		} else if i > 0 {
			sb.WriteByte('/')
		}
		decoded, err := url.PathUnescape(segment)
		if err != nil {
			decoded = segment
		}
		if decoded == "" || isSafeName(decoded) {
			sb.WriteString(decoded)
		} else {
			writeLength(&sb, decoded)
		}
	}
	if parsed.RawQuery != "" {
		sb.WriteByte('?')
		writeQueryKeys(&sb, parsed.RawQuery)
	}
	if parsed.Fragment != "" {
		sb.WriteByte('#')
		writeLength(&sb, parsed.Fragment)
	}
	return sb.String()
}

func isSafeScheme(scheme string) bool {
	if len(scheme) > maxSchemeLen || !isSafeName(strings.ReplaceAll(scheme, "+", "-")) {
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
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return
	}
	labels := strings.Split(host, ".")
	last := labels[len(labels)-1]
	ip, _, _ := strings.Cut(host, "%")
	if net.ParseIP(ip) != nil || (last != "" && isDigit(last[0])) {
		sb.WriteString("{ip}")
		return
	}
	for i, label := range labels {
		if i > 0 {
			sb.WriteByte('.')
		}
		if isSafeName(label) {
			sb.WriteString(label)
		} else {
			writeLength(sb, label)
		}
	}
}

func writeQueryKeys(sb *strings.Builder, rawQuery string) {
	keys := make(map[string]struct{})
	for part := range strings.FieldsFuncSeq(rawQuery, func(r rune) bool { return r == '&' || r == ';' }) {
		key, _, _ := strings.Cut(part, "=")
		if decoded, err := url.QueryUnescape(key); err == nil {
			key = decoded
		}
		if isSafeName(key) {
			keys[key] = struct{}{}
		} else {
			keys["{"+strconv.Itoa(len(key))+"}"] = struct{}{}
		}
	}
	sorted := make([]string, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	slices.Sort(sorted)
	if len(sorted) > maxNames {
		sorted = append(sorted[:maxNames], strconv.Itoa(len(sorted)-maxNames)+" other")
	}
	sb.WriteString(strings.Join(sorted, "&"))
}

func writeLength(sb *strings.Builder, s string) {
	sb.WriteByte('{')
	sb.WriteString(strconv.Itoa(len(s)))
	sb.WriteByte('}')
}
