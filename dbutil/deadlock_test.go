// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mau.fi/util/dbutil"
	_ "go.mau.fi/util/dbutil/litestream"
)

func initTestDB(t *testing.T) *dbutil.Database {
	db, err := dbutil.NewFromConfig("", dbutil.Config{
		PoolConfig: dbutil.PoolConfig{
			Type:         "sqlite3-fk-wal",
			URI:          ":memory:?_txlock=immediate",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
		DeadlockDetection: true,
	}, nil)
	require.NoError(t, err)
	ctx := context.Background()
	_, err = db.Exec(ctx, `
		CREATE TABLE meow (id INTEGER PRIMARY KEY, value TEXT);
		INSERT INTO meow (id, value) VALUES (1, 'meow');
		INSERT INTO meow (id, value) VALUES (2, 'meow 2');
		INSERT INTO meow (value) VALUES ('meow 3');
	`)
	require.NoError(t, err)
	return db
}

func getMeow(ctx context.Context, db dbutil.Execable, id int) (value string, err error) {
	err = db.QueryRowContext(ctx, "SELECT value FROM meow WHERE id = ?", id).Scan(&value)
	return
}

func TestDatabase_NoDeadlock(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.DoTxn(ctx, nil, func(ctx context.Context) error {
		_, err := db.Exec(ctx, "INSERT INTO meow (value) VALUES ('meow 4');")
		require.NoError(t, err)
		return nil
	}))
	val, err := getMeow(ctx, db.Execable(ctx), 4)
	require.NoError(t, err)
	require.Equal(t, "meow 4", val)
}

func TestDatabase_NoDeadlock_Goroutine(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.DoTxn(ctx, nil, func(ctx context.Context) error {
		_, err := db.Exec(ctx, "INSERT INTO meow (value) VALUES ('meow 4');")
		require.NoError(t, err)
		go func() {
			_, err := db.Exec(context.Background(), "INSERT INTO meow (value) VALUES ('meow 5');")
			require.NoError(t, err)
		}()
		time.Sleep(50 * time.Millisecond)
		return nil
	}))
	val, err := getMeow(ctx, db.Execable(ctx), 4)
	require.NoError(t, err)
	require.Equal(t, "meow 4", val)
	val, err = getMeow(ctx, db.Execable(ctx), 5)
	require.NoError(t, err)
	require.Equal(t, "meow 5", val)
}

func TestDatabase_Deadlock(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	_ = db.DoTxn(ctx, nil, func(ctx context.Context) error {
		assert.PanicsWithError(t, dbutil.ErrQueryDeadlock.Error(), func() {
			_, _ = db.Exec(context.Background(), "INSERT INTO meow (value) VALUES ('meow 4');")
		})
		return fmt.Errorf("meow")
	})
}

func TestDatabase_Deadlock_Acquire(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	_ = db.DoTxn(ctx, nil, func(ctx context.Context) error {
		assert.PanicsWithError(t, dbutil.ErrAcquireDeadlock.Error(), func() {
			_, _ = db.AcquireConn(context.Background())
		})
		return fmt.Errorf("meow")
	})
}

func TestDatabase_Deadlock_Txn(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	_ = db.DoTxn(ctx, nil, func(ctx context.Context) error {
		assert.PanicsWithError(t, dbutil.ErrTransactionDeadlock.Error(), func() {
			_ = db.DoTxn(context.Background(), nil, func(ctx context.Context) error {
				return nil
			})
		})
		return fmt.Errorf("meow")
	})
}

func TestDatabase_Deadlock_Child(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	childDB := db.Child("", nil, nil)
	_ = db.DoTxn(ctx, nil, func(ctx context.Context) error {
		assert.PanicsWithError(t, dbutil.ErrQueryDeadlock.Error(), func() {
			_, _ = childDB.Exec(context.Background(), "INSERT INTO meow (value) VALUES ('meow 4');")
		})
		return fmt.Errorf("meow")
	})
}

func TestDatabase_Deadlock_Child2(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	childDB := db.Child("", nil, nil)
	_ = childDB.DoTxn(ctx, nil, func(ctx context.Context) error {
		assert.PanicsWithError(t, dbutil.ErrQueryDeadlock.Error(), func() {
			_, _ = db.Exec(context.Background(), "INSERT INTO meow (value) VALUES ('meow 4');")
		})
		return fmt.Errorf("meow")
	})
}
