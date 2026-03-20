// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"context"
	"testing"

	"github.com/sourcenetwork/corekv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// memKeyedstore is a minimal in-memory implementation of datastore.Keyedstore
// suitable for unit tests. It does not require a transaction in the context.
type memKeyedstore struct {
	data map[string][]byte
}

func newMemKeyedstore() *memKeyedstore {
	return &memKeyedstore{data: make(map[string][]byte)}
}

func (m *memKeyedstore) Get(_ context.Context, key datastore.Key) ([]byte, error) {
	v, ok := m.data[string(key.Bytes())]
	if !ok {
		return nil, corekv.ErrNotFound
	}
	// Return a copy to avoid mutation issues.
	cp := make([]byte, len(v))
	copy(cp, v)
	return cp, nil
}

func (m *memKeyedstore) Has(_ context.Context, key datastore.Key) (bool, error) {
	_, ok := m.data[string(key.Bytes())]
	return ok, nil
}

func (m *memKeyedstore) Iterator(_ context.Context, _ datastore.IterOptions) (corekv.Iterator, error) {
	return nil, nil
}

func (m *memKeyedstore) Set(_ context.Context, key datastore.Key, value []byte) error {
	cp := make([]byte, len(value))
	copy(cp, value)
	m.data[string(key.Bytes())] = cp
	return nil
}

func (m *memKeyedstore) Delete(_ context.Context, key datastore.Key) error {
	delete(m.data, string(key.Bytes()))
	return nil
}

// newTestLWW creates an LWW register backed by the given store for testing.
func newTestLWW(store datastore.Keyedstore) *LWW {
	key := keys.DataStoreKey{}.WithDocID("testdoc")
	return NewLWW(store, "collectionVersionID", key, "fieldName")
}

// TestLWW_NewDocCreateMode_SkipsPriorityCheck verifies that when a context is
// set with NewDocCreateMode the LWW register skips the priority check and
// unconditionally stores the provided value.
//
// Scenario:
//  1. Set a value with priority 5 in normal mode → stored, priority becomes 5.
//  2. Set a value with priority 1 in normal mode → rejected (1 < 5), old value kept.
//  3. Set a value with priority 1 in NewDocCreateMode → accepted, value is stored
//     regardless of the existing priority 5.
func TestLWW_NewDocCreateMode_SkipsPriorityCheck(t *testing.T) {
	ctx := context.Background()
	store := newMemKeyedstore()
	lww := newTestLWW(store)

	highPrioValue := []byte("high-priority-value")
	lowPrioValue := []byte("low-priority-value")
	newDocValue := []byte("new-doc-value")

	// Step 1: store a value at priority 5 (normal mode).
	err := lww.setValue(ctx, highPrioValue, 5)
	require.NoError(t, err)

	// Confirm the value was stored.
	stored, err := store.Get(ctx, lww.key.WithValueFlag())
	require.NoError(t, err)
	assert.Equal(t, highPrioValue, stored, "value at priority 5 should be stored")

	// Step 2: attempt to set a value at priority 1 in normal mode — should be rejected.
	err = lww.setValue(ctx, lowPrioValue, 1)
	require.NoError(t, err)

	stored, err = store.Get(ctx, lww.key.WithValueFlag())
	require.NoError(t, err)
	assert.Equal(t, highPrioValue, stored, "lower priority value should be rejected in normal mode")

	// Step 3: set a value at priority 1 in NewDocCreateMode — priority check is skipped.
	newDocCtx := ContextWithNewDocCreateMode(ctx)
	err = lww.setValue(newDocCtx, newDocValue, 1)
	require.NoError(t, err)

	stored, err = store.Get(ctx, lww.key.WithValueFlag())
	require.NoError(t, err)
	assert.Equal(
		t,
		newDocValue,
		stored,
		"in NewDocCreateMode the value should be stored even if priority is lower",
	)
}
