// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
)

func reflectScan[T any](row Scannable) (*T, error) {
	t := new(T)
	val := reflect.ValueOf(t).Elem()
	fields := reflect.VisibleFields(val.Type())
	scanInto := make([]any, len(fields))
	for i, field := range fields {
		scanInto[i] = val.FieldByIndex(field.Index).Addr().Interface()
	}
	err := row.Scan(scanInto...)
	return t, err
}

// NewSimpleReflectRowIter creates a new RowIter that uses reflection to scan rows into the given type.
//
// This is a simplified implementation that always scans to all struct fields. It does not support any kind of struct tags.
func NewSimpleReflectRowIter[T any](rows Rows, err error) RowIter[*T] {
	return ConvertRowFn[*T](reflectScan[T]).NewRowIter(rows, err)
}

var ErrUnknownField = errors.New("column has no associated struct field")

func reflectFieldScan[T any](columns []string) func(Scannable) (*T, error) {
	return func(row Scannable) (*T, error) {
		t := new(T)
		val := reflect.ValueOf(t).Elem()
		fields := reflect.VisibleFields(val.Type())
		fieldMap := make(map[string][]int, len(fields))
		for _, field := range fields {
			colname := field.Tag.Get("column")
			if colname == "" {
				continue
			}
			if !slices.Contains(columns, colname) {
				continue
			}
			fieldMap[colname] = field.Index
		}

		scanInto := make([]any, len(columns))
		for i, column := range columns {
			index, ok := fieldMap[column]
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrUnknownField, column)
			}
			scanInto[i] = val.FieldByIndex(index).Addr().Interface()
		}
		err := row.Scan(scanInto...)
		return t, err
	}
}

// NewComplicatedReflectRowIter creates a new RowIter that uses reflection to scan rows into the given type.
//
// It scans into the struct by determining which columns were returned in the query, searching struct fields for the
// `column` tag. Columns that have no associated struct field will return ErrUnknownField.
// Struct fields that have an empty column tag, or a tag value which is not in the column list, are skipped.
func NewComplicatedReflectRowIter[T any](rows Rows, err error) RowIter[*T] {
	var cols []string
	if err == nil {
		// Don't overwrite err
		cols, err = rows.Columns()
	}
	return ConvertRowFn[*T](reflectFieldScan[T](cols)).NewRowIter(rows, err)
}
