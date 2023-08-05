// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package random_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"go.mau.fi/util/random"
)

func TestString_Length(t *testing.T) {
	for i := 0; i < 256; i++ {
		require.Len(t, random.String(i), i)
	}
}

var stringRegex = regexp.MustCompile(`^[0-9A-Za-z]*$`)
var tokenRegex = regexp.MustCompile(`^.+?_[0-9A-Za-z]*_[0-9A-Za-z]{6}$`)

func TestString_Content(t *testing.T) {
	for i := 0; i < 256; i++ {
		require.Regexp(t, stringRegex, random.String(i))
	}
}

var prefixes = []string{"ght", "hut", "meow", "FOOBAR", "🐈️"}

func TestToken(t *testing.T) {
	for _, prefix := range prefixes {
		for i := 0; i < 256; i++ {
			t.Run(fmt.Sprintf("%s-%d", prefix, i), func(t *testing.T) {
				// Format: prefix_random_checksum
				// Length: prefix (4) + 1 + random (i) + 1 + checksum (6)
				token := random.Token(prefix, i)
				require.Len(t, token, len(prefix)+1+i+1+6)
				require.Regexp(t, tokenRegex, token)
			})
		}
	}
}

func TestGetTokenPrefix(t *testing.T) {
	for _, prefix := range prefixes {
		for i := 0; i < 256; i++ {
			t.Run(fmt.Sprintf("%s-%d", prefix, i), func(t *testing.T) {
				token := random.Token(prefix, i)
				require.Equal(t, prefix, random.GetTokenPrefix(token))
			})
		}
	}
}

func TestGetTokenPrefix_Static(t *testing.T) {
	var tokens = []string{
		"meow_FXfcJomwUu9hVqmxiEqq_wZsw02",
		"meow_54aDbVIVDO4fQkB80uAkoXpnISggmVDVrV_0yIw64",
	}
	for _, token := range tokens {
		require.Equal(t, "meow", random.GetTokenPrefix(token))
	}
}

func TestGetTokenPrefix_Invalid(t *testing.T) {
	var tokens = []string{
		"meow_FXfcJomwUu9hVqmxiEqq_wZsw12",
		"meow_54aDbVIVDO4fQkB80uAkoXpnISggmVDVV_0yIw64",
	}
	for _, token := range tokens {
		require.Empty(t, random.GetTokenPrefix(token))
	}
}

func BenchmarkString8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		random.String(8)
	}
}

func BenchmarkString32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		random.String(32)
	}
}

func BenchmarkString50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		random.String(50)
	}
}

func BenchmarkString256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		random.String(256)
	}
}

func BenchmarkToken32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		random.Token("meow", 32)
	}
}

func BenchmarkGetTokenPrefix32(b *testing.B) {
	tok := random.Token("meow", 32)
	for i := 0; i < b.N; i++ {
		random.GetTokenPrefix(tok)
	}
}
