// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ptr

func Ptr[T any](val T) *T {
	return &val
}

func Val[T any](ptr *T) (val T) {
	if ptr != nil {
		val = *ptr
	}
	return
}
