// Copyright (c) 2023 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

// RowIter is a wrapper for [Rows] that allows conveniently iterating over rows
// with a predefined scanner function.
type RowIter[T any] interface {
	// Iter iterates over the rows and calls the given function for each row.
	//
	// If the function returns false, the iteration is stopped.
	// If the function returns an error, the iteration is stopped and the error is
	// returned.
	Iter(func(T) (bool, error)) error

	// AsList collects all rows into a slice.
	AsList() ([]T, error)
}

type rowIterImpl[T any] struct {
	rows Rows

	ConvertRow func(Rows) (T, error)
}

// NewRowIter creates a new RowIter from the given Rows and scanner function.
func NewRowIter[T any](rows Rows, convertFn func(Rows) (T, error)) RowIter[T] {
	return &rowIterImpl[T]{rows: rows, ConvertRow: convertFn}
}

func (i *rowIterImpl[T]) Iter(fn func(T) (bool, error)) error {
	if i == nil || i.rows == nil {
		return nil
	}
	defer i.rows.Close()

	for i.rows.Next() {
		if item, err := i.ConvertRow(i.rows); err != nil {
			return err
		} else if cont, err := fn(item); err != nil {
			return err
		} else if !cont {
			break
		}
	}
	return i.rows.Err()
}

func (i *rowIterImpl[T]) AsList() (list []T, err error) {
	err = i.Iter(func(item T) (bool, error) {
		list = append(list, item)
		return true, nil
	})
	return
}
