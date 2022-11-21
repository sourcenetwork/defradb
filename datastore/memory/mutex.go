// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package memory

import (
	"sync"

	"sync/atomic"
)

type keyMutex struct {
	querymu sync.RWMutex
	keys    map[string]*lock
}

type lock struct {
	queue *int32
	mu    sync.RWMutex
}

func newKeyMutex() *keyMutex {
	km := &keyMutex{
		keys: make(map[string]*lock),
	}
	return km
}

func (km *keyMutex) lock(key string) {
	if km.keys[key] == nil {
		var q int32 = 0
		km.keys[key] = &lock{
			queue: &q,
		}
	}
	atomic.AddInt32(km.keys[key].queue, 1)
	km.keys[key].mu.Lock()
	km.querymu.Lock()
}

func (km *keyMutex) unlock(key string) {
	if km.keys[key] == nil {
		return
	}
	km.keys[key].mu.Unlock()
	atomic.AddInt32(km.keys[key].queue, -1)
	km.querymu.Unlock()
	if atomic.LoadInt32(km.keys[key].queue) == 0 {
		km.keys[key] = nil
	}
}

func (km *keyMutex) rlock(key string) {
	if km.keys[key] == nil {
		var q int32 = 0
		km.keys[key] = &lock{
			queue: &q,
		}
	}

	atomic.AddInt32(km.keys[key].queue, 1)
	km.keys[key].mu.RLock()
}

func (km *keyMutex) runlock(key string) {
	if km.keys[key] == nil {
		return
	}
	km.keys[key].mu.RUnlock()
	atomic.AddInt32(km.keys[key].queue, -1)
	if atomic.LoadInt32(km.keys[key].queue) == 0 {
		km.keys[key] = nil
	}
}
