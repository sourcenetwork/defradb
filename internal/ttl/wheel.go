// Copyright 2026 Democratized Data Foundation
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
	"context"
	"sync"
	"time"
)

// Wheel tracks comparable keys in a fixed-size timing wheel.
type Wheel[K comparable] struct {
	tick       time.Duration
	slotCount  int64
	cur        int64
	slots      [][]*entry[K]
	index      map[K]*entry[K]
	maxTTL     int64
	mu         sync.Mutex
	running    bool
	stop       chan struct{}
	ctx        context.Context
	expireFunc func(K)
}

type entry[K comparable] struct {
	key  K
	slot int64
	idx  int64
}

// NewWheel returns a timing wheel with the given tick resolution and slot count.
func NewWheel[K comparable](
	ctx context.Context,
	tick time.Duration,
	slotCount int64,
	onExpire func(K),
) (*Wheel[K], error) {
	if tick < 1 {
		return nil, ErrInvalidTick
	}
	if slotCount < 1 {
		return nil, ErrInvalidSlotCount
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return &Wheel[K]{
		tick:       tick,
		slotCount:  slotCount,
		maxTTL:     int64(tick) * slotCount,
		slots:      make([][]*entry[K], slotCount),
		index:      make(map[K]*entry[K]),
		stop:       make(chan struct{}),
		ctx:        ctx,
		expireFunc: onExpire,
	}, nil
}

// Add tracks key until ttl has elapsed.
func (w *Wheel[K]) Add(key K, ttl time.Duration) error {
	if err := w.validTTL(ttl); err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	pos := w.slotForTTL(ttl)
	if existing, ok := w.index[key]; ok {
		w.unsafeDeleteFromSlot(existing)
	}

	e := &entry[K]{key: key, slot: pos}
	w.slots[pos] = append(w.slots[pos], e)
	e.idx = int64(len(w.slots[pos]) - 1)
	w.index[key] = e
	return nil
}

// Delete stops tracking key.
func (w *Wheel[K]) Delete(key K) {
	w.mu.Lock()
	defer w.mu.Unlock()

	e, ok := w.index[key]
	if !ok {
		return
	}

	w.unsafeDeleteFromSlot(e)
	delete(w.index, key)
}

// UpdateTTL moves an existing key to a new expiration time.
func (w *Wheel[K]) UpdateTTL(key K, ttl time.Duration) error {
	if err := w.validTTL(ttl); err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	e, ok := w.index[key]
	if !ok {
		return nil
	}

	newSlot := w.slotForTTL(ttl)
	if newSlot == e.slot {
		return nil
	}

	w.unsafeDeleteFromSlot(e)
	e.slot = newSlot
	w.slots[newSlot] = append(w.slots[newSlot], e)
	e.idx = int64(len(w.slots[newSlot]) - 1)
	w.index[key] = e
	return nil
}

func (w *Wheel[K]) slotForTTL(ttl time.Duration) int64 {
	if ttl == 0 {
		return w.cur
	}
	ticks := int64(ttl / w.tick)
	if ttl%w.tick != 0 {
		ticks++
	}
	return (w.cur + ticks - 1) % w.slotCount
}

func (w *Wheel[K]) validTTL(ttl time.Duration) error {
	if ttl < 0 {
		return ErrNegativeTTL
	}
	if int64(ttl) > w.maxTTL {
		return ErrBeyondMaxTTL
	}
	return nil
}

func (w *Wheel[K]) unsafeDeleteFromSlot(e *entry[K]) {
	bucket := w.slots[e.slot]
	lastIdx := int64(len(bucket) - 1)
	if e.idx != lastIdx {
		bucket[e.idx] = bucket[lastIdx]
		bucket[e.idx].idx = e.idx
	}
	w.slots[e.slot] = bucket[:lastIdx]
}

// Start starts the wheel if it is not already running.
func (w *Wheel[K]) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	stop := make(chan struct{})
	w.stop = stop
	tickDur := w.tick
	w.mu.Unlock()

	ticker := time.NewTicker(tickDur)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.tickOnce()
			case <-w.ctx.Done():
				w.setStopped(stop)
				return
			case <-stop:
				w.setStopped(stop)
				return
			}
		}
	}()
}

// Stop stops the wheel if it is running.
func (w *Wheel[K]) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	stop := w.stop
	w.stop = nil
	w.running = false
	close(stop)
	w.mu.Unlock()
}

func (w *Wheel[K]) setStopped(stop chan struct{}) {
	w.mu.Lock()
	if w.stop == stop {
		w.stop = nil
		w.running = false
	}
	w.mu.Unlock()
}

func (w *Wheel[K]) tickOnce() {
	w.mu.Lock()
	bucket := w.slots[w.cur]
	w.slots[w.cur] = nil
	expired := make([]K, 0, len(bucket))
	for _, e := range bucket {
		if w.index[e.key] == e {
			delete(w.index, e.key)
			expired = append(expired, e.key)
		}
	}
	w.cur = (w.cur + 1) % w.slotCount
	w.mu.Unlock()

	for _, key := range expired {
		if w.expireFunc != nil {
			w.expireFunc(key)
		}
	}
}
