// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package shape

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// JSON returns the distinct paths of a JSON document's values with their
// kind, e.g. "content.body:string" or "attachments[].size:number". Keys are
// joined with '.', array elements are "[]" and the kinds are string, number,
// bool, null, object and array. Objects and arrays are only listed by kind when
// they're empty or nested deeper than maxDepth; a maxDepth of 1 describes the
// top-level keys only. A document that fails to parse returns what was read
// before the error followed by "invalid".
func JSON(data []byte, maxDepth int) []string {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	walker := jsonWalker{dec: json.NewDecoder(bytes.NewReader(data)), maxDepth: max(maxDepth, 1)}
	walker.dec.UseNumber()
	err := walker.value("", 1)
	out := walker.set.result()
	if err != nil {
		out = append(out, "invalid")
	}
	return out
}

type jsonWalker struct {
	dec      *json.Decoder
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
	tok, err := w.dec.Token()
	if err != nil {
		return err
	}
	switch tok := tok.(type) {
	case json.Delim:
		kind, elemPath := "object", path+"."
		if tok == '[' {
			kind, elemPath = "array", path+"[]"
		} else if path == "" {
			elemPath = ""
		}
		if !w.dec.More() || depth > w.maxDepth {
			w.leaf(path, kind)
			return w.skipRest()
		}
		for w.dec.More() {
			childPath := elemPath
			if tok == '{' {
				keyTok, err := w.dec.Token()
				if err != nil {
					return err
				}
				key, _ := keyTok.(string)
				childPath += key
				if !isSafeName(key) {
					w.set.addOther(childPath)
					if err = w.skipValue(); err != nil {
						return err
					}
					continue
				}
			}
			if err = w.value(childPath, depth+1); err != nil {
				return err
			}
		}
		_, err = w.dec.Token()
		return err
	case string:
		w.leaf(path, "string")
	case json.Number:
		w.leaf(path, "number")
	case bool:
		w.leaf(path, "bool")
	case nil:
		w.leaf(path, "null")
	}
	return nil
}

func (w *jsonWalker) skipValue() error {
	tok, err := w.dec.Token()
	if err != nil {
		return err
	} else if delim, ok := tok.(json.Delim); ok && (delim == '{' || delim == '[') {
		return w.skipRest()
	}
	return nil
}

func (w *jsonWalker) skipRest() error {
	for nesting := 1; nesting > 0; {
		tok, err := w.dec.Token()
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		} else if err != nil {
			return err
		}
		switch tok {
		case json.Delim('{'), json.Delim('['):
			nesting++
		case json.Delim('}'), json.Delim(']'):
			nesting--
		}
	}
	return nil
}
