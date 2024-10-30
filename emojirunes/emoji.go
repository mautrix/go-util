// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package emojirunes

import (
	"slices"
)

//go:generate go run generate.go

// Is reports whether r is an emoji rune, a zero-width joiner (\u200D), or variation selector 16 (\uFE0F).
func Is(r rune) bool {
	_, found := slices.BinarySearch(EmojiRunes, r)
	return found
}

func isNumberEmoji(r rune, s string, i int) bool {
	return ((r >= '0' && r <= '9') || r == '#' || r == '*') &&
		len(s) > i+3 && s[i+1] == 0xef && s[i+2] == 0xb8 && s[i+3] == 0x8f
}

// IsOnlyEmojis reports whether s is a string containing only emoji runes as defined by [Is].
// Additionally, it accepts numbers (0-9) followed by variation selector 16 (\uFE0F) as emojis.
func IsOnlyEmojis(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if !Is(r) && !isNumberEmoji(r, s, i) {
			return false
		}
	}
	return true
}
