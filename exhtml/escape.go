// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exhtml

import (
	"io"
	"strings"
)

const escapedChars = "&'<>\"\r"

// EscapeWrite is a copy of [html.EscapeString] that writes directly to a [io.StringWriter]
// instead of having an internal buffer that's returned as a string.
func EscapeWrite(buf io.StringWriter, s string) error {
	i := strings.IndexAny(s, escapedChars)
	for i != -1 {
		if _, err := buf.WriteString(s[:i]); err != nil {
			return err
		}
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
			esc = "&#39;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			// "&#34;" is shorter than "&quot;".
			esc = "&#34;"
		case '\r':
			esc = "&#13;"
		default:
			panic("html: unrecognized escape character")
		}
		s = s[i+1:]
		if _, err := buf.WriteString(esc); err != nil {
			return err
		}
		i = strings.IndexAny(s, escapedChars)
	}
	_, err := buf.WriteString(s)
	return err
}
