// Copyright (c) 2025 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exstrings

import (
	"crypto/sha256"
	"crypto/subtle"
	"unsafe"
)

// UnsafeBytes returns a byte slice that points to the same memory as the input string.
//
// The returned byte slice must not be modified.
func UnsafeBytes(str string) []byte {
	return unsafe.Slice(unsafe.StringData(str), len(str))
}

// SHA256 returns the SHA-256 hash of the input string without copying the string.
func SHA256(str string) [32]byte {
	return sha256.Sum256(UnsafeBytes(str))
}

// ConstantTimeEqual compares two strings using [subtle.ConstantTimeCompare] without copying the strings.
//
// Note that ConstantTimeCompare is not constant time if the strings are of different length.
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare(UnsafeBytes(a), UnsafeBytes(b)) == 1
}
