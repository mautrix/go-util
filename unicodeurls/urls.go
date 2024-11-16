// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package unicodeurls contains URLs for Unicode data files.
// It is meant to be used in code generators that parse the files.
package unicodeurls

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"go.mau.fi/util/exerrors"
)

const UnicodeVersion = "16.0"

const EmojiVariationSequences = "https://www.unicode.org/Public/" + UnicodeVersion + ".0/ucd/emoji/emoji-variation-sequences.txt"
const EmojiTest = "https://unicode.org/Public/emoji/" + UnicodeVersion + "/emoji-test.txt"
const Confusables = "https://www.unicode.org/Public/security/" + UnicodeVersion + ".0/confusables.txt"

// ReadDataFile fetches a data file from a URL and processes it line by line with the given processor function.
func ReadDataFile(url string, processor func(string)) {
	resp := exerrors.Must(http.Get(url))
	buf := bufio.NewReader(resp.Body)
	for {
		line, err := buf.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		} else if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		processor(line)
	}
}

// ReadDataFileList fetches a data file from a URL and converts lines into array items with the given function.
func ReadDataFileList[T any](url string, processor func(string) (T, bool)) (output []T) {
	ReadDataFile(url, func(s string) {
		if item, ok := processor(s); ok {
			output = append(output, item)
		}
	})
	return
}

// ReadDataFileMap fetches a data file from a URL and converts lines into a map with the given function.
func ReadDataFileMap[Key comparable, Value any](url string, processor func(string) (Key, Value, bool)) (output map[Key]Value) {
	output = make(map[Key]Value)
	ReadDataFile(url, func(s string) {
		if key, value, ok := processor(s); ok {
			output[key] = value
		}
	})
	return
}

// ParseHex parses a list of Unicode codepoints encoded as hex into a string
func ParseHex(parts []string) string {
	output := make([]rune, len(parts))
	for i, part := range parts {
		output[i] = rune(exerrors.Must(strconv.ParseInt(part, 16, 32)))
	}
	return string(output)
}
