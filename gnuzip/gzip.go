// Copyright (C) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gnuzip

import (
	"bytes"
	"compress/gzip"
)

func GZip(body []byte) ([]byte, error) {
	var compressedBuffer bytes.Buffer
	writer := gzip.NewWriter(&compressedBuffer)
	if _, err := writer.Write(body); err != nil {
		return nil, err
	}
	return compressedBuffer.Bytes(), writer.Close()
}
