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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWheel_ValidParameters(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})

	require.NoError(t, err)
	assert.NotNil(t, w)
}

func TestNewWheel_InvalidTick_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 0, 10, func(k string) {})

	require.ErrorIs(t, err, ErrInvalidTick)
	assert.Nil(t, w)
}

func TestNewWheel_NegativeTick_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, -1*time.Second, 10, func(k string) {})

	require.ErrorIs(t, err, ErrInvalidTick)
	assert.Nil(t, w)
}

func TestNewWheel_InvalidSlotCount_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 0, func(k string) {})

	require.ErrorIs(t, err, ErrInvalidSlotCount)
	assert.Nil(t, w)
}

func TestNewWheel_NegativeSlotCount_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, -5, func(k string) {})

	require.ErrorIs(t, err, ErrInvalidSlotCount)
	assert.Nil(t, w)
}

func TestWheel_Add_SingleEntry(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)

	// Verify entry exists in index
	w.mu.Lock()
	_, exists := w.index["key1"]
	w.mu.Unlock()
	assert.True(t, exists)
}

func TestWheel_Add_MultipleEntries(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)
	err = w.Add("key2", 300*time.Millisecond)
	require.NoError(t, err)
	err = w.Add("key3", 500*time.Millisecond)
	require.NoError(t, err)

	w.mu.Lock()
	assert.Len(t, w.index, 3)
	w.mu.Unlock()
}

func TestWheel_Delete_ExistingEntry(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)

	w.Delete("key1")

	w.mu.Lock()
	_, exists := w.index["key1"]
	w.mu.Unlock()
	assert.False(t, exists)
}

func TestWheel_UpdateTTL_ExistingEntry(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 20, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)

	w.mu.Lock()
	originalSlot := w.index["key1"].slot
	w.mu.Unlock()

	// Update to a different TTL that would result in a different slot
	err = w.UpdateTTL("key1", 1000*time.Millisecond)
	require.NoError(t, err)

	w.mu.Lock()
	newSlot := w.index["key1"].slot
	w.mu.Unlock()

	assert.NotEqual(t, originalSlot, newSlot)
}

func TestWheel_StartStop(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	w.Start()
	w.mu.Lock()
	running := w.running
	w.mu.Unlock()
	assert.True(t, running)

	w.Stop()
	w.mu.Lock()
	running = w.running
	w.mu.Unlock()
	assert.False(t, running)
}

func TestWheel_Add_TTLLessThanTick_RoundsUpToOneTick(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	// TTL of 50ms with a 100ms tick should round to 1 tick
	err = w.Add("key1", 50*time.Millisecond)
	require.NoError(t, err)

	w.mu.Lock()
	e := w.index["key1"]
	// Should be in slot 1 (cur=0 + 1 tick = slot 1)
	assert.Equal(t, int64(1), e.slot)
	w.mu.Unlock()
}

func TestWheel_Add_ZeroTTL(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	// Zero TTL should be allowed and result in slot 0 (current slot)
	err = w.Add("key1", 0)
	require.NoError(t, err)

	w.mu.Lock()
	e := w.index["key1"]
	assert.Equal(t, int64(0), e.slot)
	w.mu.Unlock()
}

func TestWheel_UpdateTTL_SameSlot_NoOp(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)

	w.mu.Lock()
	originalSlot := w.index["key1"].slot
	originalIdx := w.index["key1"].idx
	w.mu.Unlock()

	// Update to same TTL - should be a no-op since slot doesn't change
	err = w.UpdateTTL("key1", 500*time.Millisecond)
	require.NoError(t, err)

	w.mu.Lock()
	newSlot := w.index["key1"].slot
	newIdx := w.index["key1"].idx
	w.mu.Unlock()

	assert.Equal(t, originalSlot, newSlot)
	assert.Equal(t, originalIdx, newIdx)
}

func TestWheel_Add_MultipleEntriesSameSlot(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	// Add multiple entries with same TTL - all should go to same slot
	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)
	err = w.Add("key2", 500*time.Millisecond)
	require.NoError(t, err)
	err = w.Add("key3", 500*time.Millisecond)
	require.NoError(t, err)

	w.mu.Lock()
	slot := w.index["key1"].slot
	bucket := w.slots[slot]
	w.mu.Unlock()

	assert.Len(t, bucket, 3)
}

func TestWheel_TickOnce_ExpiresEntries(t *testing.T) {
	ctx := context.Background()
	var expired []string
	var mu sync.Mutex

	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {
		mu.Lock()
		expired = append(expired, k)
		mu.Unlock()
	})
	require.NoError(t, err)

	// Add entry that will expire on next tick (TTL < tick rounds to 1)
	// But we want it in current slot (slot 0), so use TTL=0
	err = w.Add("key1", 0)
	require.NoError(t, err)

	// Manually tick the wheel
	w.tickOnce()

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, expired, "key1")
}

func TestWheel_Expiration_Integration(t *testing.T) {
	ctx := context.Background()
	expired := make(chan string, 1)

	w, err := NewWheel(ctx, 50*time.Millisecond, 10, func(k string) {
		expired <- k
	})
	require.NoError(t, err)

	err = w.Add("key1", 50*time.Millisecond)
	require.NoError(t, err)

	w.Start()
	defer w.Stop()

	select {
	case key := <-expired:
		assert.Equal(t, "key1", key)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected key to expire within timeout")
	}
}

func TestWheel_Delete_NonExistentKey(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	// Should not panic
	w.Delete("nonexistent")
}

func TestWheel_UpdateTTL_NonExistentKey(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	// Should not error, just return nil
	err = w.UpdateTTL("nonexistent", 500*time.Millisecond)
	require.NoError(t, err)
}

func TestWheel_Add_ExactlyMaxTTL(t *testing.T) {
	ctx := context.Background()
	tick := 100 * time.Millisecond
	slots := int64(10)
	maxTTL := time.Duration(int64(tick) * slots)

	w, err := NewWheel(ctx, tick, slots, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", maxTTL)
	require.NoError(t, err)
}

func TestWheel_Add_BeyondMaxTTL_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tick := 100 * time.Millisecond
	slots := int64(10)
	maxTTL := time.Duration(int64(tick) * slots)

	w, err := NewWheel(ctx, tick, slots, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", maxTTL+1)
	require.ErrorIs(t, err, ErrBeyondMaxTTL)
}

func TestWheel_Add_NegativeTTL_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", -100*time.Millisecond)
	require.ErrorIs(t, err, ErrNegativeTTL)
}

func TestWheel_UpdateTTL_NegativeTTL_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)

	err = w.UpdateTTL("key1", -100*time.Millisecond)
	require.ErrorIs(t, err, ErrNegativeTTL)
}

func TestWheel_UpdateTTL_BeyondMaxTTL_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tick := 100 * time.Millisecond
	slots := int64(10)
	maxTTL := time.Duration(int64(tick) * slots)

	w, err := NewWheel(ctx, tick, slots, func(k string) {})
	require.NoError(t, err)

	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)

	err = w.UpdateTTL("key1", maxTTL+1)
	require.ErrorIs(t, err, ErrBeyondMaxTTL)
}

func TestWheel_Start_WhenAlreadyRunning_Ignored(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	w.Start()
	defer w.Stop()

	// Should not panic or create issues
	w.Start()
	w.Start()

	w.mu.Lock()
	running := w.running
	w.mu.Unlock()
	assert.True(t, running)
}

func TestWheel_Stop_WhenNotRunning_Ignored(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	// Should not panic
	w.Stop()
	w.Stop()
}

func TestWheel_MultipleStartStopCycles(t *testing.T) {
	ctx := context.Background()
	expiredCount := 0
	var mu sync.Mutex

	w, err := NewWheel(ctx, 50*time.Millisecond, 10, func(k string) {
		mu.Lock()
		expiredCount++
		mu.Unlock()
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		w.Start()
		time.Sleep(100 * time.Millisecond)
		w.Stop()
	}

	// No panics and wheel still functional
	w.mu.Lock()
	running := w.running
	w.mu.Unlock()
	assert.False(t, running)
}

func TestWheel_SlotWraparound(t *testing.T) {
	ctx := context.Background()
	var expired []string
	var mu sync.Mutex

	// Small wheel to test wraparound
	w, err := NewWheel(ctx, 10*time.Millisecond, 3, func(k string) {
		mu.Lock()
		expired = append(expired, k)
		mu.Unlock()
	})
	require.NoError(t, err)

	// Manually advance the wheel past the wraparound point
	for i := 0; i < 5; i++ {
		w.tickOnce()
	}

	// cur should have wrapped around
	w.mu.Lock()
	cur := w.cur
	w.mu.Unlock()

	// After 5 ticks with 3 slots: 5 % 3 = 2
	assert.Equal(t, int64(2), cur)
}

func TestWheel_DeleteFromMiddleOfSlot(t *testing.T) {
	ctx := context.Background()
	w, err := NewWheel(ctx, 100*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	// Add three entries to the same slot
	err = w.Add("key1", 500*time.Millisecond)
	require.NoError(t, err)
	err = w.Add("key2", 500*time.Millisecond)
	require.NoError(t, err)
	err = w.Add("key3", 500*time.Millisecond)
	require.NoError(t, err)

	// Delete middle entry
	w.Delete("key2")

	// Verify remaining entries are still accessible
	w.mu.Lock()
	_, exists1 := w.index["key1"]
	_, exists2 := w.index["key2"]
	_, exists3 := w.index["key3"]
	w.mu.Unlock()

	assert.True(t, exists1)
	assert.False(t, exists2)
	assert.True(t, exists3)
}

func TestWheel_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w, err := NewWheel(ctx, 50*time.Millisecond, 10, func(k string) {})
	require.NoError(t, err)

	w.ctx = ctx // Set the context for the running goroutine
	w.Start()

	// Cancel context
	cancel()

	// Give goroutine time to exit
	time.Sleep(100 * time.Millisecond)

	// Wheel may still report running=true since context cancellation
	// doesn't set running=false, but the goroutine should have exited
}
