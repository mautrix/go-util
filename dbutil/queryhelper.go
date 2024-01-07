// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/exp/constraints"
)

// DataStruct is an interface for structs that represent a single database row.
type DataStruct[T any] interface {
	Scan(row Scannable) (T, error)
}

// QueryHelper is a generic helper struct for SQL query execution boilerplate.
//
// After implementing the Scan and Init methods in a data struct, the query
// helper allows writing query functions in a single line.
type QueryHelper[T DataStruct[T]] struct {
	db      *Database
	newFunc func(qh *QueryHelper[T]) T
}

func MakeQueryHelper[T DataStruct[T]](db *Database, new func(qh *QueryHelper[T]) T) *QueryHelper[T] {
	return &QueryHelper[T]{db: db, newFunc: new}
}

// ValueOrErr is a helper function that returns the value if err is nil, or
// returns nil and the error if err is not nil. It can be used to avoid
// `if err != nil { return nil, err }` boilerplate in certain cases like
// DataStruct.Scan implementations.
func ValueOrErr[T any](val *T, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	return val, nil
}

// StrPtr returns a pointer to the given string, or nil if the string is empty.
func StrPtr[T ~string](val T) *string {
	if val == "" {
		return nil
	}
	strVal := string(val)
	return &strVal
}

// NumPtr returns a pointer to the given number, or nil if the number is zero.
func NumPtr[T constraints.Integer | constraints.Float](val T) *T {
	if val == 0 {
		return nil
	}
	return &val
}

func (qh *QueryHelper[T]) GetDB() *Database {
	return qh.db
}

func (qh *QueryHelper[T]) New() T {
	return qh.newFunc(qh)
}

// Exec executes a query with ExecContext and returns the error.
//
// It omits the sql.Result return value, as it is rarely used. When the result
// is wanted, use `qh.GetDB().Exec(...)` instead, which is
// otherwise equivalent.
func (qh *QueryHelper[T]) Exec(ctx context.Context, query string, args ...any) error {
	_, err := qh.db.Exec(ctx, query, args...)
	return err
}

func (qh *QueryHelper[T]) scanNew(row Scannable) (T, error) {
	return qh.New().Scan(row)
}

// QueryOne executes a query with QueryRowContext, uses the associated DataStruct
// to scan it, and returns the value. If the query returns no rows, it returns nil
// and no error.
func (qh *QueryHelper[T]) QueryOne(ctx context.Context, query string, args ...any) (val T, err error) {
	val, err = qh.scanNew(qh.db.QueryRow(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return val, err
}

// QueryMany executes a query with QueryContext, uses the associated DataStruct
// to scan each row, and returns the values. If the query returns no rows, it
// returns a non-nil zero-length slice and no error.
func (qh *QueryHelper[T]) QueryMany(ctx context.Context, query string, args ...any) ([]T, error) {
	rows, err := qh.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return NewRowIter(rows, qh.scanNew).AsList()
}
