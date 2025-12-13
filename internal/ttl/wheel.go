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
	"context"
	"sync"
	"time"
)

// Wheel is a specialzed TTL structure used to track
// expiration times on arbitrary entries `K`. It is
// used in the TTLCache as the time tracking system.
// It works similar to a ring buffer, but instead of
// used for allocating memory, it tracks the "next"
// slot of items to expire.
//
// It has two main parameters, tick rate and slot
// count. Tick rate is the smallest resolution of
// time the TTL supports. Slot count is the total
// number of consesucative ticks in the future we can
// track. Entries are added with a TTL value, which
// is truncated to the unit of resolution defined by
// the tick rate (Eg: A tick rate of 1 second will
// truncate 1.5 seconds to 1).
//
// The combination of tick and slots determines the
// max supported TTL. Example: A tick rate of 100ms
// with a slot count of 2000 will support TTLs up
// to 200 seconds in the future.
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
	expireFunc func(k K)
}

type entry[K comparable] struct {
	key  K
	slot int64 // which slot in the wheel
	idx  int64 // index in slots[slot]
}

func NewWheel[K comparable](ctx context.Context, tick time.Duration, slotCount int64, onExpire func(e K)) (*Wheel[K], error) {
	if tick < 1 {
		return nil, ErrInvalidTick
	} else if slotCount < 1 {
		return nil, ErrInvalidSlotCount
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

func (w *Wheel[K]) Add(key K, ttl time.Duration) error {
	err := w.validTTL(ttl)
	if err != nil {
		return err
	}

	ticks := int64(ttl / w.tick)
	if ticks == 0 && ttl > 0 {
		ticks = 1
	}
	pos := (w.cur + ticks) % w.slotCount

	w.mu.Lock()
	defer w.mu.Unlock()

	e := &entry[K]{key: key, slot: pos}
	w.slots[pos] = append(w.slots[pos], e)
	e.idx = int64(len(w.slots[pos]) - 1)
	w.index[key] = e
	return nil
}

func (w *Wheel[K]) Delete(key K) {
	w.mu.Lock()
	defer w.mu.Unlock()

	e, ok := w.index[key]
	if !ok {
		return
	}

	// delete by replacing with last element from the bucket
	// shrinking the bucket
	w.unsafeDeleteFromSlot(e)

	delete(w.index, key)
}

func (w *Wheel[K]) UpdateTTL(key K, ttl time.Duration) error {
	err := w.validTTL(ttl)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	e, ok := w.index[key]
	if !ok {
		return nil
	}
	curSlot := e.slot

	ticks := int64(ttl / w.tick)
	if ticks == 0 {
		ticks = 1
	}
	newSlot := (w.cur + ticks) % w.slotCount
	// we can exit early here since the target
	// slot is the same as the current slot. This
	// happens if the Update call occured before
	// the wheel has ticked forward.
	if newSlot == curSlot {
		return nil
	}

	// move to new slot
	w.unsafeDeleteFromSlot(e)
	e.slot = newSlot
	w.slots[newSlot] = append(w.slots[newSlot], e)
	e.idx = int64(len(w.slots[newSlot]) - 1)

	return nil
}

func (w *Wheel[K]) validTTL(ttl time.Duration) error {
	if ttl < 0 {
		return ErrNegativeTTL
	} else if int64(ttl) > w.maxTTL {
		return ErrBeyondMaxTTL
	}
	return nil
}

// unsafeDeleteFromSlot is only unsafe if the caller
// hasn't aquired the lock
func (w *Wheel[K]) unsafeDeleteFromSlot(e *entry[K]) {
	bucket := w.slots[e.slot]
	lastIdx := len(bucket) - 1
	bucket[e.idx] = bucket[lastIdx]
	bucket[e.idx].idx = e.idx
	w.slots[e.slot] = bucket[:lastIdx]
}

// Start will start the wheel if it isn't already running.
// Consecutive calls to Start without a cooresponding Stop
// will be ignored.
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

	t := time.NewTicker(tickDur)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-t.C:
				w.tickOnce()
			case <-w.ctx.Done():
				return
			case <-stop:
				return
			}
		}
	}()
}

// Stop the wheel, consecutive calls to Stop without
// a coorsponding Start will be ignored
func (w *Wheel[K]) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	close(w.stop)
	w.stop = nil
	w.running = false
	w.mu.Unlock()
}

// move the ticker one slot forward and
// expire any "now" entries.
func (w *Wheel[K]) tickOnce() {
	w.mu.Lock()
	bucket := w.slots[w.cur]
	w.slots[w.cur] = nil
	for _, e := range bucket {
		delete(w.index, e.key)
	}
	w.cur = (w.cur + 1) % w.slotCount
	w.mu.Unlock()

	// process outside the lock
	for _, e := range bucket {
		w.expireFunc(e.key)
	}
}
