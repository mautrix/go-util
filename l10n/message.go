// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package l10n contains the localizable message type used for user-facing
// strings that bridges and other services send to clients.
package l10n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
)

// Params are the values substituted into a template. Values must be strings,
// integers or Messages. All integer types are normalized to int64 when rendering.
type Params map[string]any

// Message is a user-facing string.
type Message struct {
	// Key is "<namespace>:<id>", or empty if the message isn't localizable.
	Key string `json:"key,omitempty"`
	// Text is the rendered English text.
	Text string `json:"text"`
	// Params are the values used to render Text.
	Params Params `json:"params,omitempty"`
}

// Unlocalized wraps text that has no template, e.g. from a remote network.
func Unlocalized(text string) Message {
	return Message{Text: text}
}

func (m Message) String() string {
	return m.Text
}

func (m Message) IsZero() bool {
	return m.Key == "" && m.Text == "" && len(m.Params) == 0
}

func (m Message) IsLocalizable() bool {
	return m.Key != ""
}

func (m Message) MarshalZerologObject(e *zerolog.Event) {
	e.Str("text", m.Text)
	if m.Key != "" {
		e.Str("key", m.Key)
	}
}

// UnmarshalJSON also accepts a bare string, which becomes an unlocalized message.
func (m *Message) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*m = Unlocalized(text)
		return nil
	} else if bytes.Equal(data, []byte("null")) {
		return nil
	}
	if jsonObjectDepth(data) > maxObjectDepth {
		return fmt.Errorf("message params are nested more than %d levels deep", maxNestingDepth)
	}
	var raw struct {
		Key    string                     `json:"key"`
		Text   string                     `json:"text"`
		Params map[string]json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = Message{Key: raw.Key, Text: raw.Text}
	if len(raw.Params) > 0 {
		m.Params = make(Params, len(raw.Params))
		for name, val := range raw.Params {
			parsed, err := parseParamValue(val)
			if err != nil {
				return fmt.Errorf("invalid value for param %q: %w", name, err)
			}
			m.Params[name] = parsed
		}
	}
	return nil
}

const maxNestingDepth = 8

// Each nested message adds two objects: the params map and the message itself.
const maxObjectDepth = 1 + 2*maxNestingDepth

func jsonObjectDepth(data []byte) int {
	var depth, maxDepth int
	inString := false
	for i := 0; i < len(data); i++ {
		switch c := data[i]; {
		case inString && c == '\\':
			i++
		case c == '"':
			inString = !inString
		case inString:
		case c == '{':
			depth++
			maxDepth = max(maxDepth, depth)
		case c == '}':
			depth--
		}
	}
	return maxDepth
}

func parseParamValue(raw json.RawMessage) (any, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) > 0 && raw[0] == '{' {
		var msg Message
		err := json.Unmarshal(raw, &msg)
		return msg, err
	} else if len(raw) > 0 && raw[0] == '"' {
		var str string
		err := json.Unmarshal(raw, &str)
		return str, err
	}
	val, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("must be a string, an integer or a message")
	}
	return val, nil
}
