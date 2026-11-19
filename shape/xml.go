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
// the root, at most 32. Namespaces are ignored. A document that fails to parse
// returns the elements that were closed before the error and the elements
// containing them, followed by "invalid", so that text with stray angle
// brackets doesn't list its words. Input with no elements is "invalid" too.
func XML(data []byte, maxDepth int) []string {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Entity = xml.HTMLEntity
	dec.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	maxDepth = clampDepth(maxDepth)
	var set nameSet
	var path []string
	var pending [][]string
	confirmed, unsafeDepth := 0, 0
	sawElement := false
	var err error
	for {
		var tok xml.Token
		tok, err = dec.Token()
		if err != nil {
			break
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			sawElement = true
			path = append(path, tok.Name.Local)
			pending = append(pending, nil)
			if unsafeDepth > 0 || len(path) > maxDepth {
				continue
			}
			if !isSafeName(tok.Name.Local) {
				unsafeDepth = len(path)
				set.addOther(strings.Join(path, "/"))
				continue
			}
			elementPath := strings.Join(path, "/")
			names := []string{elementPath}
			for _, attr := range tok.Attr {
				if attr.Name.Space == "xmlns" || (attr.Name.Space == "" && attr.Name.Local == "xmlns") {
					continue
				} else if isSafeName(attr.Name.Local) {
					names = append(names, elementPath+"@"+attr.Name.Local)
				} else {
					set.addOther(elementPath + "@" + attr.Name.Local)
				}
			}
			pending[len(pending)-1] = names
		case xml.EndElement:
			if len(path) == 0 {
				continue
			}
			for _, names := range pending[confirmed:] {
				for _, name := range names {
					set.add(name)
				}
			}
			if len(path) == unsafeDepth {
				unsafeDepth = 0
			}
			path = path[:len(path)-1]
			pending = pending[:len(pending)-1]
			confirmed = len(pending)
		}
	}
	out := set.result()
	if !errors.Is(err, io.EOF) || !sawElement {
		out = append(out, "invalid")
	}
	return out
}
