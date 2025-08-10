// Copyright (c) 2025 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exstrings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLongestSequenceOf(t *testing.T) {
	testCases := []struct {
		input    string
		runeVal  rune
		expected int
	}{
		{"", 'a', 0},
		{"aaaaa", 'a', 5},
		{"aaaaab", 'a', 5},
		{"bbbaaa", 'a', 3},
		{"bbbaaa", 'b', 3},
		{"a", 'a', 1},
		{"b", 'a', 0},
		{"aabbaa", 'a', 2},
		{"aabbaaa", 'a', 3},
		{"aabbaaa", 'b', 2},
	}
	for _, tc := range testCases {
		result := LongestSequenceOf(tc.input, tc.runeVal)
		assert.Equal(t, tc.expected, result)
	}
}

func TestPrefixByteRunLength(t *testing.T) {
	testCases := []struct {
		input    string
		byteVal  byte
		expected int
	}{
		{"", 'a', 0},
		{"aaaaa", 'a', 5},
		{"aaaaab", 'a', 5},
		{"bbbaaa", 'a', 0},
		{"bbbaaa", 'b', 3},
		{"a", 'a', 1},
		{"b", 'a', 0},
		{"aabbaa", 'a', 2},
		{"aabbaaa", 'a', 2},
		{"aabbaaa", 'b', 0},
		{"    ", ' ', 4},
		{"    a", ' ', 4},
	}
	for _, tc := range testCases {
		result := PrefixByteRunLength(tc.input, tc.byteVal)
		assert.Equal(t, tc.expected, result)
	}
}

func TestCollapseSpaces(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{" ", " "},
		{"  ", " "},
		{"   ", " "},
		{"a", "a"},
		{" a ", " a "},
		{"  a  ", " a "},
		{"a b", "a b"},
		{"a  b", "a b"},
		{"                a  b", " a b"},
		{"  a   b  ", " a b "},
		{"  a      b           ", " a b "},
	}
	for _, tc := range testCases {
		result := CollapseSpaces(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}
