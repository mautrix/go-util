// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.mau.fi/util/dbutil"
)

type AbstractMassInsertable[T dbutil.Array] struct {
	Data T
}

func (a AbstractMassInsertable[T]) GetMassInsertValues() T {
	return a.Data
}

type OneParamMassInsertable = AbstractMassInsertable[[1]any]

func TestNewMassInsertBuilder_InvalidParams(t *testing.T) {
	assert.PanicsWithError(t, "invalid insert query: placeholders not found", func() {
		dbutil.NewMassInsertBuilder[OneParamMassInsertable, [1]any]("", "")
	})
	assert.PanicsWithError(t, "invalid placeholder template: static placeholder $1 not found", func() {
		dbutil.NewMassInsertBuilder[OneParamMassInsertable, [1]any]("INSERT INTO foo VALUES ($1, $2)", "")
	})
	assert.PanicsWithError(t, "invalid placeholder template: non-static placeholder $2 found", func() {
		dbutil.NewMassInsertBuilder[OneParamMassInsertable, [1]any]("INSERT INTO foo VALUES ($1, $2)", "($1, $2)")
	})
	assert.PanicsWithError(t, "invalid placeholder template: extra string found", func() {
		dbutil.NewMassInsertBuilder[OneParamMassInsertable, [1]any]("INSERT INTO foo VALUES ($1, $2)", "($1)")
	})
}

func TestMassInsertBuilder_Build(t *testing.T) {
	builder := dbutil.NewMassInsertBuilder[OneParamMassInsertable, [1]any]("INSERT INTO foo VALUES ($1, $2)", "($1, $%d)")
	query, values := builder.Build([1]any{"hi"}, []OneParamMassInsertable{{[1]any{"hmm"}}, {[1]any{"meow"}}, {[1]any{"third"}}})
	assert.Equal(t, "INSERT INTO foo VALUES ($1, $2), ($1, $3), ($1, $4)", query)
	assert.Equal(t, []any{"hi", "hmm", "meow", "third"}, values)
}
