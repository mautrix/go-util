// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exsync

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForRefCount[Key comparable](t *testing.T, km *KeyedMutex[Key], key Key, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		km.lock.Lock()
		lock := km.locks[key]
		got := 0
		if lock != nil {
			got = lock.c
		}
		km.lock.Unlock()
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for key %v to have ref count %d (got %d)", key, want, got)
		}
		runtime.Gosched()
	}
}

func TestNewKeyedMutex(t *testing.T) {
	km := NewKeyedMutex[string]()
	require.NotNil(t, km)
	assert.NotNil(t, km.locks)
	assert.Empty(t, km.locks)
}

func TestKeyedMutexZeroValue(t *testing.T) {
	var km KeyedMutex[int]

	km.Lock(1)
	assert.Len(t, km.locks, 1)
	km.Unlock(1)
	assert.Empty(t, km.locks)
}

func TestKeyedMutexSerializesSameKey(t *testing.T) {
	km := NewKeyedMutex[string]()
	km.Lock("shared")

	acquired := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		km.Lock("shared")
		close(acquired)
		<-release
		km.Unlock("shared")
		close(done)
	}()

	// Waiting until the second reference has been registered ensures the
	// goroutine has reached Lock before checking that it is blocked.
	waitForRefCount(t, km, "shared", 2)
	select {
	case <-acquired:
		t.Fatal("the same key was acquired while it was already locked")
	default:
	}

	km.Unlock("shared")
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("waiting lock was not acquired after the key was unlocked")
	}
	assert.Len(t, km.locks, 1)

	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("waiting goroutine did not finish")
	}
	assert.Empty(t, km.locks)
}

func TestKeyedMutexAllowsDifferentKeys(t *testing.T) {
	km := NewKeyedMutex[string]()
	km.Lock("first")
	require.True(t, km.TryLock("second"))

	km.Unlock("second")
	km.Unlock("first")
	assert.Empty(t, km.locks)
}

func TestKeyedMutexTryLock(t *testing.T) {
	km := NewKeyedMutex[string]()
	require.True(t, km.TryLock("key"))
	assert.False(t, km.TryLock("key"))

	km.Unlock("key")
	assert.Empty(t, km.locks)
}

func TestKeyedMutexWithLock(t *testing.T) {
	km := NewKeyedMutex[string]()
	unlock := km.WithLock("key")

	assert.False(t, km.TryLock("key"))
	require.True(t, km.TryLock("other"))
	km.Unlock("other")
	assert.Len(t, km.locks, 1)

	unlock()
	assert.Empty(t, km.locks)

	require.True(t, km.TryLock("key"))
	km.Unlock("key")
}

func TestKeyedMutexUnlockUnknownKeyPanics(t *testing.T) {
	km := NewKeyedMutex[string]()

	assert.PanicsWithError(t, "exsync/multilock: unlock of unlocked key missing", func() {
		km.Unlock("missing")
	})
}
