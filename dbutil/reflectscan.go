// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"fmt"
	"reflect"
)

func reflectScan[T any]() ConvertRowFn[*T] {
	fields := reflect.VisibleFields(reflect.TypeFor[T]())
	return func(row Scannable) (*T, error) {
		t := new(T)
		val := reflect.ValueOf(t).Elem()
		scanInto := make([]any, len(fields))
		for i, field := range fields {
			scanInto[i] = val.FieldByIndex(field.Index).Addr().Interface()
		}
		err := row.Scan(scanInto...)
		return t, err
	}
}

func getFieldMap[T any]() map[string][]int {
	fields := reflect.VisibleFields(reflect.TypeFor[T]())
	m := make(map[string][]int, len(fields))
	for _, field := range fields {
		sqlName := field.Tag.Get("sql")
		if sqlName == "" {
			sqlName = field.Name
		}
		m[sqlName] = field.Index
	}
	return m
}

func reflectScanComplicated[T any](rows Rows, err error) (ConvertRowFn[*T], error) {
	if err != nil {
		return nil, err
	}
	var fields [][]int
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("reflectscan: failed to get columns: %w", err)
	}
	fields = make([][]int, len(columns))
	fieldMap := getFieldMap[T]()
	var ok bool
	for i, col := range columns {
		fields[i], ok = fieldMap[col]
		if !ok {
			return nil, fmt.Errorf("reflectscan: column %q does not match any struct field", col)
		}
	}
	return func(row Scannable) (*T, error) {
		t := new(T)
		val := reflect.ValueOf(t).Elem()
		scanInto := make([]any, len(fields))
		for i, idx := range fields {
			scanInto[i] = val.FieldByIndex(idx).Addr().Interface()
		}
		err := row.Scan(scanInto...)
		return t, err
	}, nil
}

// NewSimpleReflectRowIter creates a new RowIter that uses reflection to scan rows into the given type.
//
// This is a simplified implementation that always scans to all struct fields. It does not support any kind of struct tags.
func NewSimpleReflectRowIter[T any](rows Rows, err error) RowIter[*T] {
	return reflectScan[T]().NewRowIter(rows, err)
}

// NewComplicatedReflectRowIter creates a new RowIter that uses reflection to scan rows into the given type.
//
// This will use the `sql` struct tag. The column names returned by the db must match an explicit struct tag exactly.
func NewComplicatedReflectRowIter[T any](rows Rows, err error) RowIter[*T] {
	fn, err := reflectScanComplicated[T](rows, err)
	return fn.NewRowIter(rows, err)
}
