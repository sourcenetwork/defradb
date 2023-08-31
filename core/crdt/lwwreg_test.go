// Copyright 2022 Democratized Data Foundation
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

	dag "github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"
	"github.com/ugorji/go/codec"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

func newMockStore() datastore.DSReaderWriter {
	return datastore.AsDSReaderWriter(ds.NewMapDatastore())
}

func setupLWWRegister() LWWRegister {
	store := newMockStore()
	key := core.DataStoreKey{DocKey: "AAAA-BBBB"}
	return NewLWWRegister(store, core.CollectionSchemaVersionKey{}, key, "")
}

func setupLoadedLWWRegster(ctx context.Context) LWWRegister {
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))
	addDelta.SetPriority(1)
	lww.Merge(ctx, addDelta)
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

func TestLWWReisterFollowupMerge(t *testing.T) {
	ctx := context.Background()
	lww := setupLoadedLWWRegster(ctx)
	addDelta := lww.Set([]byte("test2"))
	addDelta.SetPriority(2)
	lww.Merge(ctx, addDelta)

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
	lww := setupLoadedLWWRegster(ctx)
	addDelta := lww.Set([]byte("test-1"))
	addDelta.SetPriority(0)
	lww.Merge(ctx, addDelta)

	val, err := lww.Value(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(val) != string([]byte("test")) {
		t.Errorf("Incorrect merge state, want %s, have %s", []byte("test"), val)
	}
}

func TestLWWRegisterDeltaInit(t *testing.T) {
	delta := &LWWRegDelta{
		Data: []byte("test"),
	}

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

func TestLWWRegisterDeltaMarshal(t *testing.T) {
	delta := &LWWRegDelta{
		Data:     []byte("test"),
		Priority: uint64(10),
	}
	bytes, err := delta.Marshal()
	if err != nil {
		t.Errorf("Marshal returned an error: %v", err)
		return
	}
	if len(bytes) == 0 {
		t.Error("Expected Marshal to return serialized bytes, but output is empty")
		return
	}

	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(bytes, h)
	unmarshaledDelta := &LWWRegDelta{}
	err = dec.Decode(unmarshaledDelta)
	if err != nil {
		t.Errorf("Decode returned an error: %v", err)
		return
	}

	if !reflect.DeepEqual(delta.Data, unmarshaledDelta.Data) {
		t.Errorf(
			"Unmarshalled data value doesn't match expected. Want %v, have %v",
			[]byte("test"),
			unmarshaledDelta.Data,
		)
		return
	}

	if delta.Priority != unmarshaledDelta.Priority {
		t.Errorf(
			"Unmarshalled priority value doesn't match. Want %v, have %v",
			uint64(10),
			unmarshaledDelta.Priority,
		)
	}
}

func makeNode(delta core.Delta, heads []cid.Cid) (ipld.Node, error) {
	var data []byte
	var err error
	if delta != nil {
		data, err = delta.Marshal()
		if err != nil {
			return nil, err
		}
	}

	nd := dag.NodeWithData(data)
	// The cid builder defaults to v0, we want to be using v1 CIDs
	nd.SetCidBuilder(
		cid.V1Builder{
			Codec:    cid.Raw,
			MhType:   mh.SHA2_256,
			MhLength: -1,
		})

	for _, h := range heads {
		err = nd.AddRawLink("", &ipld.Link{Cid: h})
		if err != nil {
			return nil, err
		}
	}

	return nd, nil
}

func TestLWWRegisterDeltaDecode(t *testing.T) {
	delta := &LWWRegDelta{
		Data:     []byte("test"),
		Priority: uint64(10),
	}

	node, err := makeNode(delta, []cid.Cid{})
	if err != nil {
		t.Errorf("Received errors while creating node: %v", err)
		return
	}

	reg := LWWRegister{}
	extractedDelta, err := reg.DeltaDecode(node)
	if err != nil {
		t.Errorf("Received error while extracing node: %v", err)
		return
	}

	typedExtractedDelta, ok := extractedDelta.(*LWWRegDelta)
	if !ok {
		t.Error("Extracted delta from node is NOT a LWWRegDelta type")
		return
	}

	if !reflect.DeepEqual(typedExtractedDelta, delta) {
		t.Errorf(
			"Extracted delta is not the same value as the original. Expected %v, have %v",
			delta,
			typedExtractedDelta,
		)
		return
	}
}
