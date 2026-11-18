// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package shape

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// XML returns the distinct paths of an XML document's elements and attributes,
// e.g. "msg/appmsg/title" or "msg/img@aeskey", down to maxDepth elements from
// the root. Namespaces are ignored. A document that fails to parse returns what
// was read before the error followed by "invalid".
func XML(data []byte, maxDepth int) []string {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	maxDepth = max(maxDepth, 1)
	var set nameSet
	var path []string
	unsafeDepth := 0
	var err error
	for {
		var tok xml.Token
		tok, err = dec.Token()
		if err != nil {
			break
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			path = append(path, tok.Name.Local)
			if unsafeDepth > 0 || len(path) > maxDepth {
				continue
			}
			if !isSafeName(tok.Name.Local) {
				unsafeDepth = len(path)
				set.addOther(strings.Join(path, "/"))
				continue
			}
			elementPath := strings.Join(path, "/")
			set.add(elementPath)
			for _, attr := range tok.Attr {
				if attr.Name.Space == "xmlns" || (attr.Name.Space == "" && attr.Name.Local == "xmlns") {
					continue
				} else if isSafeName(attr.Name.Local) {
					set.add(elementPath + "@" + attr.Name.Local)
				} else {
					set.addOther(elementPath + "@" + attr.Name.Local)
				}
			}
		case xml.EndElement:
			if len(path) == unsafeDepth {
				unsafeDepth = 0
			}
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
		}
	}
	out := set.result()
	if !errors.Is(err, io.EOF) {
		out = append(out, "invalid")
	}
	return out
}
