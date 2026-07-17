// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exsync

import (
	"fmt"
	"sync"
)

type lockWithRefCount struct {
	sync.Mutex
	c int
}

type KeyedMutex[Key comparable] struct {
	lock  sync.Mutex
	locks map[Key]*lockWithRefCount
}

func (km *KeyedMutex[Key]) lockSelf() {
	km.lock.Lock()
	if km.locks == nil {
		km.locks = make(map[Key]*lockWithRefCount)
	}
}

func (km *KeyedMutex[Key]) getLock(k Key) *lockWithRefCount {
	km.lockSelf()
	defer km.lock.Unlock()
	l, ok := km.locks[k]
	if !ok {
		l = &lockWithRefCount{}
		km.locks[k] = l
	}
	l.c++
	return l
}

func (km *KeyedMutex[Key]) Lock(k Key) {
	km.getLock(k).Lock()
}

func (km *KeyedMutex[Key]) TryLock(k Key) bool {
	return km.getLock(k).TryLock()
}

func (km *KeyedMutex[Key]) Unlock(k Key) {
	km.lockSelf()
	defer km.lock.Unlock()
	l, ok := km.locks[k]
	if !ok {
		panic(fmt.Errorf("exsync/multilock: unlock of unlocked key %v", k))
	}
	// This can happen inside the main lock as it should be instant
	l.Unlock()
	l.c--
	if l.c == 0 {
		delete(km.locks, k)
	} else if l.c < 0 {
		// l.Unlock will already panic if the lock is not held, so this should never be hit
		panic(fmt.Errorf("exsync/multilock: impossible case: %v's ref count is %d", k, l.c))
	}
}
