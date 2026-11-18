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
// digits. Anything else, such as user IDs, phone numbers or message IDs used as
// keys, is counted in a final "N other" entry and its subtree is skipped.
// Results are sorted and deduplicated, so equal shapes log identically. At most
// 64 names are listed; the rest are counted in "N other" too.
//
// Use the redact package instead when values need to be correlated between log
// lines.
package shape

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const (
	maxNames      = 64
	maxTracked    = 1024
	maxNameLen    = 64
	maxNameDigits = 4
)

func isSafeName(name string) bool {
	if len(name) == 0 || len(name) > maxNameLen {
		return false
	}
	digits := 0
	for i := 0; i < len(name); i++ {
		switch c := name[i]; {
		case c >= '0' && c <= '9':
			digits++
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_', c == '-', c == '.', c == '$':
		default:
			return false
		}
	}
	return digits <= maxNameDigits
}

type nameSet struct {
	names     map[string]struct{}
	other     map[string]struct{}
	untracked int
}

func (s *nameSet) add(name string) {
	if s.names == nil {
		s.names = make(map[string]struct{})
	}
	s.track(s.names, name)
}

func (s *nameSet) addOther(raw string) {
	if s.other == nil {
		s.other = make(map[string]struct{})
	}
	s.track(s.other, raw)
}

func (s *nameSet) track(set map[string]struct{}, name string) {
	if _, ok := set[name]; ok {
		return
	} else if len(s.names)+len(s.other) >= maxTracked {
		s.untracked++
	} else {
		set[name] = struct{}{}
	}
}

func (s *nameSet) result() []string {
	out := slices.Sorted(maps.Keys(s.names))
	other := len(s.other) + s.untracked
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
	var set nameSet
	for key := range m {
		if isSafeName(string(key)) {
			set.add(string(key))
		} else {
			set.addOther(string(key))
		}
	}
	return set.result()
}

// URL describes a URL as its scheme, host, path and sorted query keys. Path
// segments, query keys and opaque parts that don't look like schema names are
// replaced by their length in braces, as is the fragment, and user info becomes
// "{userinfo}", e.g. "https://example.com/l/{22}?lang&uuid".
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
		sb.WriteString(parsed.Scheme)
		sb.WriteByte(':')
	}
	if parsed.Opaque != "" {
		writeLength(&sb, parsed.Opaque)
	}
	if parsed.Host != "" || parsed.User != nil {
		sb.WriteString("//")
		if parsed.User != nil {
			sb.WriteString("{userinfo}@")
		}
		sb.WriteString(parsed.Host)
	}
	for i, segment := range strings.Split(parsed.Path, "/") {
		if i > 0 {
			sb.WriteByte('/')
		}
		if segment == "" || isSafeName(segment) {
			sb.WriteString(segment)
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

func writeLength(sb *strings.Builder, s string) {
	sb.WriteByte('{')
	sb.WriteString(strconv.Itoa(len(s)))
	sb.WriteByte('}')
}
