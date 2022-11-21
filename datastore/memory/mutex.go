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
	"time"

	"sync/atomic"
)

type keyMutex struct {
	querymu sync.RWMutex
	keys    map[string]*lock
	close   chan struct{}
}

type lock struct {
	queue *int32
	mu    sync.RWMutex
}

func newKeyMutex() *keyMutex {
	km := &keyMutex{
		keys:  make(map[string]*lock),
		close: make(chan struct{}),
	}
	go km.clearLocks()
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
}

func (km *keyMutex) clearLocks() {
	for {
		select {
		case <-km.close:
			return
		case <-time.After(10 * time.Minute):
			for k, v := range km.keys {
				if atomic.LoadInt32(v.queue) == 0 {
					km.keys[k] = nil
				}
			}
		}
	}
}
