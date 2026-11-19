// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build goexperiment.jsonv2 || go1.27

package shape

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"io"
)

// JSON returns the distinct paths of a JSON document's values with their
// kind, e.g. "content.body:string" or "attachments[].size:number". Keys are
// joined with '.', array elements are "[]" and the kinds are string, number,
// bool, null, object and array. Objects and arrays are only listed by kind when
// they're empty or nested deeper than maxDepth; a maxDepth of 1 describes the
// top-level keys only, and depths over 32 are treated as 32. A document that
// fails to parse or has data after its first value returns what was read
// before the error followed by "invalid".
//
// JSON needs encoding/json/jsontext, so it returns nil unless built with Go
// 1.27 or newer, or with GOEXPERIMENT=jsonv2.
func JSON(data []byte, maxDepth int) []string {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	walker := jsonWalker{
		dec: jsontext.NewDecoder(
			bytes.NewReader(data),
			jsontext.AllowDuplicateNames(true),
			jsontext.AllowInvalidUTF8(true),
		),
		maxDepth: clampDepth(maxDepth),
	}
	err := walker.value("", 1)
	if err == nil {
		if _, err = walker.dec.ReadToken(); errors.Is(err, io.EOF) {
			err = nil
		} else if err == nil {
			err = errTrailingData
		}
	}
	out := walker.set.result()
	if err != nil {
		out = append(out, "invalid")
	}
	return out
}

var errTrailingData = errors.New("trailing data after JSON value")

type jsonWalker struct {
	dec      *jsontext.Decoder
	set      nameSet
	maxDepth int
}

func (w *jsonWalker) leaf(path, kind string) {
	if path == "" {
		w.set.add(kind)
	} else {
		w.set.add(path + ":" + kind)
	}
}

func (w *jsonWalker) value(path string, depth int) error {
	tok, err := w.dec.ReadToken()
	if err != nil {
		return err
	}
	switch kind := tok.Kind(); kind {
	case jsontext.KindBeginObject, jsontext.KindBeginArray:
		end, name, elemPath := jsontext.KindEndObject, "object", path+"."
		if kind == jsontext.KindBeginArray {
			end, name, elemPath = jsontext.KindEndArray, "array", path+"[]"
		} else if path == "" {
			elemPath = ""
		}
		if w.dec.PeekKind() == end || depth > w.maxDepth {
			w.leaf(path, name)
			for w.dec.PeekKind() != end {
				if err = w.dec.SkipValue(); err != nil {
					return err
				}
			}
		}
		for w.dec.PeekKind() != end {
			childPath := elemPath
			if kind == jsontext.KindBeginObject {
				keyTok, err := w.dec.ReadToken()
				if err != nil {
					return err
				}
				key := keyTok.String()
				childPath += key
				if !isSafeName(key) {
					w.set.addOther(childPath)
					if err = w.dec.SkipValue(); err != nil {
						return err
					}
					continue
				}
			}
			if err = w.value(childPath, depth+1); err != nil {
				return err
			}
		}
		_, err = w.dec.ReadToken()
		return err
	case jsontext.KindString:
		w.leaf(path, "string")
	case jsontext.KindNumber:
		w.leaf(path, "number")
	case jsontext.KindTrue, jsontext.KindFalse:
		w.leaf(path, "bool")
	case jsontext.KindNull:
		w.leaf(path, "null")
	}
	return nil
}
