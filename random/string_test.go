// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package random_test

import (
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

func TestToken(t *testing.T) {
	for i := 0; i < 256; i++ {
		// Format: prefix_random_checksum
		// Length: prefix (4) + 1 + random (i) + 1 + checksum (6)
		token := random.Token("meow", i)
		require.Len(t, token, i+5+7)
		require.Regexp(t, tokenRegex, token)
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
