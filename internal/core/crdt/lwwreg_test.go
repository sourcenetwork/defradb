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
	"reflect"
	"testing"

	ds "github.com/ipfs/go-datastore"

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

func TestLWWRegisterAddDelta(t *testing.T) {
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))

	if !reflect.DeepEqual(addDelta.Data, []byte("test")) {
		t.Errorf("Delta unexpected value, was %s want %s", addDelta.Data, []byte("test"))
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
