// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package variationselector_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mau.fi/util/exerrors"
	"go.mau.fi/util/variationselector"
)

func TestAdd_Full(t *testing.T) {
	resp := get(t, "https://raw.githubusercontent.com/milesj/emojibase/master/packages/data/en/data.raw.json")
	var emojis []emojibaseEmoji
	exerrors.PanicIfNotNil(json.NewDecoder(resp.Body).Decode(&emojis))
	for _, e := range emojis {
		compareEmoji(t, e.Emoji, variationselector.Add)
		for _, s := range e.Skins {
			compareEmoji(t, s.Emoji, variationselector.Add)
		}
	}
}

func TestFullyQualify_Full(t *testing.T) {
	resp := get(t, "https://raw.githubusercontent.com/iamcal/emoji-data/master/emoji.json")
	var emojis []iamcalEmoji
	exerrors.PanicIfNotNil(json.NewDecoder(resp.Body).Decode(&emojis))
	for _, e := range emojis {
		compareEmoji(t, unifiedToUnicode(e.Unified), variationselector.FullyQualify)
		for _, s := range e.SkinVariations {
			compareEmoji(t, unifiedToUnicode(s.Unified), variationselector.FullyQualify)
		}
	}
}

func get(t *testing.T, url string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "GitHub actions @ https://github.com/mautrix/go-util/blob/main/variationselector/variationselector_test.go")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

type emojibaseEmoji struct {
	Emoji   string           `json:"emoji"`
	Hexcode string           `json:"hexcode"`
	Skins   []emojibaseEmoji `json:"skins"`
}

type iamcalEmoji struct {
	Unified        string                 `json:"unified"`
	SkinVariations map[string]iamcalEmoji `json:"skin_variations"`
}

func unifiedToUnicode(input string) string {
	parts := strings.Split(input, "-")
	output := make([]rune, len(parts))
	for i, part := range parts {
		output[i] = rune(exerrors.Must(strconv.ParseInt(part, 16, 32)))
	}
	return string(output)
}

func unicodeToUnified(input string) string {
	runes := []rune(input)
	output := make([]string, len(runes))
	for i, r := range runes {
		output[i] = fmt.Sprintf("%X", r)
	}
	return strings.Join(output, "-")
}

func compareEmoji(t *testing.T, orig string, fn func(string) string) {
	proc := fn(orig)
	if proc != orig {
		t.Errorf("emoji: %s\nexpected: %s\ngot:      %s", orig, unicodeToUnified(orig), unicodeToUnified(proc))
	}
}

func TestAdd(t *testing.T) {
	assert.Equal(t, "\U0001f44d\U0001f3fd", variationselector.Add("\U0001f44d\U0001f3fd"))
	assert.Equal(t, "\U0001f44d\ufe0f", variationselector.Add("\U0001f44d"))
	assert.Equal(t, "\U0001f44d\ufe0f", variationselector.Add("\U0001f44d\ufe0f"))
	assert.Equal(t, "4\ufe0f\u20e3", variationselector.Add("4\u20e3"))
	assert.Equal(t, "4\ufe0f\u20e3", variationselector.Add("4\ufe0f\u20e3"))
	assert.Equal(t, "4", variationselector.Add("4"))
	assert.Equal(t, "\U0001f914", variationselector.Add("\U0001f914"))
	assert.Equal(t, "\U0001f408\u200d\u2b1b", variationselector.Add("\U0001f408\u200d\u2b1b"))
	assert.Equal(t, "\u2122\ufe0f", variationselector.Add("\u2122"))
	assert.Equal(t, "\u2122\ufe0e", variationselector.Add("\u2122\ufe0e"))
}

func TestFullyQualify(t *testing.T) {
	assert.Equal(t, "\U0001f44d", variationselector.FullyQualify("\U0001f44d"))
	assert.Equal(t, "\U0001f44d", variationselector.FullyQualify("\U0001f44d\ufe0f"))
	assert.Equal(t, "4\ufe0f\u20e3", variationselector.FullyQualify("4\u20e3"))
	assert.Equal(t, "4\ufe0f\u20e3", variationselector.FullyQualify("4\ufe0f\u20e3"))
	assert.Equal(t, "4", variationselector.FullyQualify("4"))
	assert.Equal(t, "\U0001f914", variationselector.FullyQualify("\U0001f914"))
	assert.Equal(t, "\u263a\ufe0f", variationselector.FullyQualify("\u263a"))
	assert.Equal(t, "\u263a\ufe0f", variationselector.FullyQualify("\u263a"))
	assert.Equal(t, "\U0001f3f3\ufe0f\u200D\U0001f308", variationselector.FullyQualify("\U0001f3f3\u200D\U0001f308"))
	assert.Equal(t, "\U0001f3f3\ufe0f\u200D\U0001f308", variationselector.FullyQualify("\U0001f3f3\ufe0f\u200D\U0001f308"))
	assert.Equal(t, "\U0001f408\u200d\u2b1b", variationselector.FullyQualify("\U0001f408\u200d\u2b1b"))
	assert.Equal(t, "\u2122\ufe0f", variationselector.FullyQualify("\u2122"))
	assert.Equal(t, "\u2122\ufe0e", variationselector.FullyQualify("\u2122\ufe0e"))
}

func TestRemove(t *testing.T) {
	assert.Equal(t, "\U0001f44d", variationselector.Remove("\U0001f44d"))
	assert.Equal(t, "\U0001f44d", variationselector.Remove("\U0001f44d\ufe0f"))
	assert.Equal(t, "4\u20e3", variationselector.Remove("4\u20e3"))
	assert.Equal(t, "4\u20e3", variationselector.Remove("4\ufe0f\u20e3"))
	assert.Equal(t, "4", variationselector.Remove("4"))
	assert.Equal(t, "\U0001f914", variationselector.Remove("\U0001f914"))
}

func ExampleAdd() {
	fmt.Println(strconv.QuoteToASCII(variationselector.Add("\U0001f44d")))           // thumbs up (needs selector)
	fmt.Println(strconv.QuoteToASCII(variationselector.Add("\U0001f44d\ufe0f")))     // thumbs up with variation selector (stays as-is)
	fmt.Println(strconv.QuoteToASCII(variationselector.Add("\U0001f44d\U0001f3fd"))) // thumbs up with skin tone (shouldn't get selector)
	fmt.Println(strconv.QuoteToASCII(variationselector.Add("\U0001f914")))           // thinking face (shouldn't get selector)
	// Output:
	// "\U0001f44d\ufe0f"
	// "\U0001f44d\ufe0f"
	// "\U0001f44d\U0001f3fd"
	// "\U0001f914"
}

func ExampleFullyQualify() {
	fmt.Println(strconv.QuoteToASCII(variationselector.FullyQualify("\U0001f44d")))                       // thumbs up (already fully qualified)
	fmt.Println(strconv.QuoteToASCII(variationselector.FullyQualify("\U0001f44d\ufe0f")))                 // thumbs up with variation selector (variation selector removed)
	fmt.Println(strconv.QuoteToASCII(variationselector.FullyQualify("\U0001f44d\U0001f3fd")))             // thumbs up with skin tone (already fully qualified)
	fmt.Println(strconv.QuoteToASCII(variationselector.FullyQualify("\u263a")))                           // smiling face (unqualified, should get selector)
	fmt.Println(strconv.QuoteToASCII(variationselector.FullyQualify("\U0001f3f3\u200d\U0001f308")))       // rainbow flag (unqualified, should get selector)
	fmt.Println(strconv.QuoteToASCII(variationselector.FullyQualify("\U0001f3f3\ufe0f\u200d\U0001f308"))) // rainbow flag with variation selector (already fully qualified)
	// Output:
	// "\U0001f44d"
	// "\U0001f44d"
	// "\U0001f44d\U0001f3fd"
	// "\u263a\ufe0f"
	// "\U0001f3f3\ufe0f\u200d\U0001f308"
	// "\U0001f3f3\ufe0f\u200d\U0001f308"
}

func ExampleRemove() {
	fmt.Println(strconv.QuoteToASCII(variationselector.Remove("\U0001f44d")))
	fmt.Println(strconv.QuoteToASCII(variationselector.Remove("\U0001f44d\ufe0f")))
	// Output:
	// "\U0001f44d"
	// "\U0001f44d"
}

func doBenchmarkAdd(b *testing.B, input string) {
	for i := 0; i < b.N; i++ {
		variationselector.Add(input)
	}
}

func BenchmarkAddNumber4(b *testing.B) {
	doBenchmarkAdd(b, "4️\u20e3")
}

func BenchmarkAddBlackCat(b *testing.B) {
	doBenchmarkAdd(b, "\U0001f408\u200d\u2b1b")
}

func BenchmarkAddString(b *testing.B) {
	doBenchmarkAdd(b, "This is a slightly longer 🤔 string 🐈️ that contains a few emojis 👍️")
}

func BenchmarkAddLongString(b *testing.B) {
	doBenchmarkAdd(b, strings.Repeat("This is a slightly longer 🤔 string 🐈️ that contains a few emojis 👍️", 1000))
}

func BenchmarkAddNoEmojis(b *testing.B) {
	doBenchmarkAdd(b, "This is a slightly longer string that does not contain any emojis")
}

func BenchmarkAddNoEmojisLong(b *testing.B) {
	doBenchmarkAdd(b, strings.Repeat("This is a slightly longer string that does not contain any emojis", 1000))
}

func doBenchmarkFullyQualify(b *testing.B, input string) {
	for i := 0; i < b.N; i++ {
		variationselector.FullyQualify(input)
	}
}

func BenchmarkFullyQualifyNumber4(b *testing.B) {
	doBenchmarkFullyQualify(b, "4\ufe0f\u20e3")
}

func BenchmarkFullyQualifyBlackCat(b *testing.B) {
	doBenchmarkFullyQualify(b, "\U0001f408\u200d\u2b1b")
}

func BenchmarkFullyQualifyString(b *testing.B) {
	doBenchmarkFullyQualify(b, "This is a slightly longer 🤔 string 🐈️ that contains a few emojis 👍️")
}

func BenchmarkFullyQualifyLongString(b *testing.B) {
	doBenchmarkFullyQualify(b, strings.Repeat("This is a slightly longer 🤔 string 🐈️ that contains a few emojis 👍️", 1000))
}

func BenchmarkFullyQualifyNoEmojis(b *testing.B) {
	doBenchmarkFullyQualify(b, "This is a slightly longer string that does not contain any emojis")
}

func BenchmarkFullyQualifyNoEmojisLong(b *testing.B) {
	doBenchmarkFullyQualify(b, strings.Repeat("This is a slightly longer string that does not contain any emojis", 1000))
}
