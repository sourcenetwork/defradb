// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package ttl

import (
	"sync"
	"time"
)

type Wheel[K any] struct {
	tick       time.Duration
	slotCount  int
	cur        int
	slots      [][]K
	mu         sync.Mutex
	stop       chan struct{}
	expireFunc func(e K)
}

func NewWheel[K any](tick time.Duration, slotCount int, onExpire func(e K)) *Wheel[K] {
	return &Wheel[K]{
		tick:       tick,
		slotCount:  slotCount,
		slots:      make([][]K, slotCount),
		stop:       make(chan struct{}),
		expireFunc: onExpire,
	}
}

func (w *Wheel[K]) Add(key K, ttl time.Duration) {
	if ttl < 0 {
		ttl = 0
	}
	ticks := int(ttl / w.tick)
	if ticks == 0 && ttl > 0 {
		ticks = 1
	}
	pos := (w.cur + ticks) % w.slotCount

	w.mu.Lock()
	w.slots[pos] = append(w.slots[pos], key)
	w.mu.Unlock()
}

func (w *Wheel[K]) Start() {
	t := time.NewTicker(w.tick)
	go func() {
		for {
			select {
			case <-t.C:
				w.tickOnce()
			case <-w.stop:
				t.Stop()
				return
			}
		}
	}()
}

func (w *Wheel[K]) Stop() { close(w.stop) }

func (w *Wheel[K]) tickOnce() {
	w.mu.Lock()
	bucket := w.slots[w.cur]
	w.slots[w.cur] = nil
	w.cur = (w.cur + 1) % w.slotCount
	w.mu.Unlock()

	// process outside the lock
	for _, k := range bucket {
		w.expireFunc(k)
	}
}
