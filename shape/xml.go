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
// containing them, followed by "invalid". Input with no elements or with text
// outside every element, such as a message with stray angle brackets, is just
// "invalid".
func XML(data []byte, maxDepth int) []string {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(bytes.TrimPrefix(data, []byte("\ufeff"))))
	dec.Entity = xml.HTMLEntity
	dec.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	maxDepth = clampDepth(maxDepth)
	var set nameSet
	var path []string
	var pending []xmlNames
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
			pending = append(pending, xmlNames{})
			if unsafeDepth > 0 || len(path) > maxDepth {
				continue
			}
			frame := &pending[len(pending)-1]
			elementPath := strings.Join(path, "/")
			if !isSafeName(tok.Name.Local) {
				unsafeDepth = len(path)
				frame.other = append(frame.other, elementPath)
				continue
			}
			frame.names = append(frame.names, elementPath)
			for _, attr := range tok.Attr {
				if attr.Name.Space == "xmlns" || (attr.Name.Space == "" && attr.Name.Local == "xmlns") {
					continue
				} else if isSafeName(attr.Name.Local) {
					frame.names = append(frame.names, elementPath+"@"+attr.Name.Local)
				} else {
					frame.other = append(frame.other, elementPath+"@"+attr.Name.Local)
				}
			}
		case xml.CharData:
			if len(path) == 0 && len(bytes.TrimSpace(tok)) > 0 {
				return []string{"invalid"}
			}
		case xml.EndElement:
			if len(path) == 0 {
				continue
			}
			for _, frame := range pending[confirmed:] {
				for _, name := range frame.names {
					set.add(name)
				}
				for _, raw := range frame.other {
					set.addOther(raw)
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

type xmlNames struct {
	names []string
	other []string
}
