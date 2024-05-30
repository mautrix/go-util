// Copyright (c) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package progress provides wrappers for [io.Writer] and [io.Reader] that
// report the progress of the read or write operation via a callback.
package progress

const defaultUpdateInterval = 256 * 1024
