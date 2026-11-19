// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build !go1.27

package shape

// JSON needs encoding/json/jsontext, so it returns nil when built with Go
// 1.26.
func JSON(data []byte, maxDepth int) []string {
	return nil
}
