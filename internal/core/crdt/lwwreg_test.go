// Copyright 2024 Democratized Data Foundation
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
	"reflect"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
)

func newMockStore() datastore.DSReaderWriter {
	return datastore.AsDSReaderWriter(ds.NewMapDatastore())
}

func setupLWWRegister() LWWRegister {
	store := newMockStore()
	key := core.DataStoreKey{DocID: "AAAA-BBBB"}
	return NewLWWRegister(store, core.CollectionSchemaVersionKey{}, key, "")
}

func setupLoadedLWWRegister(t *testing.T, ctx context.Context) LWWRegister {
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))
	addDelta.SetPriority(1)
	err := lww.Merge(ctx, addDelta)
	require.NoError(t, err)
	return lww
}

func TestLWWRegisterAddDelta(t *testing.T) {
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))

	if !reflect.DeepEqual(addDelta.Data, []byte("test")) {
		t.Errorf("Delta unexpected value, was %s want %s", addDelta.Data, []byte("test"))
	}
}

func TestLWWRegisterInitialMerge(t *testing.T) {
	ctx := context.Background()
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))
	addDelta.SetPriority(1)
	err := lww.Merge(ctx, addDelta)
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
		return
	}

	val, err := lww.Value(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	expectedVal := []byte("test")
	if string(val) != string(expectedVal) {
		t.Errorf("Mismatch value for LWWRegister, was %s want %s", val, expectedVal)
	}
}

func TestLWWRegisterFollowupMerge(t *testing.T) {
	ctx := context.Background()
	lww := setupLoadedLWWRegister(t, ctx)
	addDelta := lww.Set([]byte("test2"))
	addDelta.SetPriority(2)
	err := lww.Merge(ctx, addDelta)
	require.NoError(t, err)

	val, err := lww.Value(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(val) != string([]byte("test2")) {
		t.Errorf("Incorrect merge state, want %s, have %s", []byte("test2"), val)
	}
}

func TestLWWRegisterOldMerge(t *testing.T) {
	ctx := context.Background()
	lww := setupLoadedLWWRegister(t, ctx)
	addDelta := lww.Set([]byte("test-1"))
	addDelta.SetPriority(0)
	err := lww.Merge(ctx, addDelta)
	require.NoError(t, err)

	val, err := lww.Value(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(val) != string([]byte("test")) {
		t.Errorf("Incorrect merge state, want %s, have %s", []byte("test"), val)
	}
}

func TestLWWRegisterDeltaInit(t *testing.T) {
	delta := &LWWRegDelta{}

	var _ core.Delta = delta // checks if LWWRegDelta implements core.Delta (also checked in the implementation code, but w.e)
}

func TestLWWRegisterDeltaGetPriority(t *testing.T) {
	delta := &LWWRegDelta{
		Data:     []byte("test"),
		Priority: uint64(10),
	}

	if delta.GetPriority() != uint64(10) {
		t.Errorf(
			"LWWRegDelta: GetPriority returned incorrect value, want %v, have %v",
			uint64(10),
			delta.GetPriority(),
		)
	}
}

func TestLWWRegisterDeltaSetPriority(t *testing.T) {
	delta := &LWWRegDelta{
		Data: []byte("test"),
	}

	delta.SetPriority(10)

	if delta.GetPriority() != uint64(10) {
		t.Errorf(
			"LWWRegDelta: SetPriority incorrect value, want %v, have %v",
			uint64(10),
			delta.GetPriority(),
		)
	}
}
