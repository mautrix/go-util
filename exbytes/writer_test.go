// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exbytes_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.mau.fi/util/exbytes"
	"go.mau.fi/util/exerrors"
)

func ExampleWriter() {
	x := make([]byte, 0, 11)
	w := (*exbytes.Writer)(&x)
	w.Write([]byte("hello"))
	w.WriteByte(' ')
	w.WriteString("world")
	fmt.Println(string(x))
	// Output: hello world
}

func TestWriter_HelloWorld(t *testing.T) {
	x := make([]byte, 0, 11)
	w := (*exbytes.Writer)(&x)
	exerrors.Must(w.Write([]byte("hello")))
	exerrors.PanicIfNotNil(w.WriteByte(' '))
	exerrors.Must(w.WriteString("world"))
	assert.Equal(t, "hello world", string(x))
}
