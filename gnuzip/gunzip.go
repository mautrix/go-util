// Copyright (C) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gnuzip

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
)

func MaybeGUnzip(body []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		if errors.Is(err, gzip.ErrHeader) {
			return body, nil
		} else {
			return nil, err
		}
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
