// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil_test

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.mau.fi/util/dbutil"
	"go.mau.fi/util/random"
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

func TestMassInsertBuilder_Build_MultiValue(t *testing.T) {
	ts := time.Now().UnixMilli()
	builder := dbutil.NewMassInsertBuilder[AbstractMassInsertable[[5]any], [3]any]("INSERT INTO foo VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", "($1, $2, $%d, $%d, $3, $%d, $%d, $%d)")
	query, values := builder.Build([3]any{"first", "second", 3}, []AbstractMassInsertable[[5]any]{
		{[5]any{"foo1", 123, true, "meow", ts}},
		{[5]any{"foo2", 666, false, "meow", ts + 1}},
		{[5]any{"foo3", 999, true, "no meow", ts + 2}},
		{[5]any{"foo4", 0, true, "meow!", 0}},
	})
	assert.Equal(t, "INSERT INTO foo VALUES ($1, $2, $4, $5, $3, $6, $7, $8), ($1, $2, $9, $10, $3, $11, $12, $13), ($1, $2, $14, $15, $3, $16, $17, $18), ($1, $2, $19, $20, $3, $21, $22, $23)", query)
	assert.Equal(t, []any{"first", "second", 3, "foo1", 123, true, "meow", ts, "foo2", 666, false, "meow", ts + 1, "foo3", 999, true, "no meow", ts + 2, "foo4", 0, true, "meow!", 0}, values)
}

func TestMassInsertBuilder_Build_CompareWithManual(t *testing.T) {
	builder := dbutil.NewMassInsertBuilder[AbstractMassInsertable[[5]any], [3]any]("INSERT INTO foo VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", "($1, $2, $%d, $%d, $3, $%d, $%d, $%d)")
	data := makeBenchmarkData[[5]any](100)
	manualQuery, manualParams := buildMassInsertManual(data)
	query, params := builder.Build([3]any{"first", "second", 3}, data)
	assert.Equal(t, manualQuery, query)
	assert.Equal(t, manualParams, params)
}

func makeBenchmarkData[T dbutil.Array](n int) []AbstractMassInsertable[T] {
	outArr := make([]AbstractMassInsertable[T], n)
	dataLen := len(outArr[0].Data)
	for i := 0; i < dataLen; i++ {
		var val any
		switch rand.Intn(4) {
		case 0:
			val = rand.Intn(1000)
		case 1:
			val = rand.Intn(1) == 0
		case 2:
			val = time.Now().UnixMilli()
		case 3:
			val = random.String(16)
		}
		for j := 0; j < len(outArr); j++ {
			outArr[j].Data[i] = val
		}
	}
	return outArr
}

func BenchmarkMassInsertBuilder_Build5x100(b *testing.B) {
	builder := dbutil.NewMassInsertBuilder[AbstractMassInsertable[[5]any], [3]any]("INSERT INTO foo VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", "($1, $2, $%d, $%d, $3, $%d, $%d, $%d)")
	data := makeBenchmarkData[[5]any](100)
	for i := 0; i < b.N; i++ {
		builder.Build([3]any{"first", "second", 3}, data)
	}
}

func buildMassInsertManual(data []AbstractMassInsertable[[5]any]) (string, []any) {
	const queryTemplate = `INSERT INTO foo VALUES %s`
	const placeholderTemplate = "($1, $2, $%d, $%d, $3, $%d, $%d, $%d)"
	placeholders := make([]string, len(data))
	params := make([]any, 3+len(data)*5)
	params[0] = "first"
	params[1] = "second"
	params[2] = 3
	for j, item := range data {
		baseIndex := j*5 + 3
		params[baseIndex] = item.Data[0]
		params[baseIndex+1] = item.Data[1]
		params[baseIndex+2] = item.Data[2]
		params[baseIndex+3] = item.Data[3]
		params[baseIndex+4] = item.Data[4]
		placeholders[j] = fmt.Sprintf(placeholderTemplate, baseIndex+1, baseIndex+2, baseIndex+3, baseIndex+4, baseIndex+5)
	}
	query := fmt.Sprintf(queryTemplate, strings.Join(placeholders, ", "))
	return query, params
}

func BenchmarkMassInsertBuilder_Build5x100_Manual(b *testing.B) {
	data := makeBenchmarkData[[5]any](100)
	for i := 0; i < b.N; i++ {
		buildMassInsertManual(data)
	}
}
