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
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.mau.fi/util/exerrors"
)

const UnicodeVersion = "17.0.0"
const BaseURL = "https://www.unicode.org/Public/" + UnicodeVersion

const EmojiVariationSequences = BaseURL + "/ucd/emoji/emoji-variation-sequences.txt"
const EmojiTest = BaseURL + "/emoji/emoji-test.txt"
const Confusables = BaseURL + "/security/confusables.txt"

type ReadParams struct {
	ProcessComments bool
}

var dataDir = os.Getenv("MAUTRIX_GO_UTIL_UNICODE_DATA_DIR")

// ReadDataFile fetches a data file from a URL and processes it line by line with the given processor function.
func ReadDataFile(url string, processor func(string), params ...ReadParams) {
	var param ReadParams
	if len(params) > 0 {
		param = params[0]
	}
	cachePath := filepath.Join(dataDir, filepath.Base(url))
	f, err := os.Open(cachePath)
	var buf *bufio.Reader
	if err != nil {
		req := exerrors.Must(http.NewRequest(http.MethodGet, url, nil))
		req.Header.Set("User-Agent", "Unicode data parser +https://github.com/mautrix/go-util")
		resp := exerrors.Must(http.DefaultClient.Do(req))
		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url))
		}
		buf = bufio.NewReader(resp.Body)
	} else {
		defer f.Close()
		buf = bufio.NewReader(f)
	}
	for {
		line, err := buf.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		} else if line == "" || (strings.HasPrefix(line, "#") && !param.ProcessComments) {
			continue
		}
		processor(line)
	}
}

// ReadDataFileList fetches a data file from a URL and converts lines into array items with the given function.
func ReadDataFileList[T any](url string, processor func(string) (T, bool), params ...ReadParams) (output []T) {
	ReadDataFile(url, func(s string) {
		if item, ok := processor(s); ok {
			output = append(output, item)
		}
	}, params...)
	return
}

// ReadDataFileMap fetches a data file from a URL and converts lines into a map with the given function.
func ReadDataFileMap[Key comparable, Value any](url string, processor func(string) (Key, Value, bool), params ...ReadParams) (output map[Key]Value) {
	output = make(map[Key]Value)
	ReadDataFile(url, func(s string) {
		if key, value, ok := processor(s); ok {
			output[key] = value
		}
	}, params...)
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
