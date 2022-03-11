// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clock

import (
	"bytes"
	"context"
	"encoding/binary"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	mh "github.com/multiformats/go-multihash"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

func newRandomCID() cid.Cid {
	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}

	// And then feed it some data
	bs := make([]byte, 4)
	i := rand.Uint32()
	binary.LittleEndian.PutUint32(bs, i)

	c, err := pref.Sum(bs)
	if err != nil {
		return cid.Undef
	}

	return c
}

func newHeadSet() *heads {
	s := newDS()

	return newHeadset(datastore.AsDSReaderWriter(s), core.HeadStoreKey{}.WithDocKey("mydockey").WithFieldId("1"))
}

func TestHeadsWrite(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c := newRandomCID()
	err := heads.write(ctx, heads.store, c, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
		return
	}
}

func TestHeadsLoad(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c := newRandomCID()
	err := heads.write(ctx, heads.store, c, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
		return
	}

	h, err := heads.load(ctx, c)
	if err != nil {
		t.Error("failed to load from head set:", err)
		return
	}

	if h != uint64(1) {
		t.Errorf("Incorrect value from head set load(), have %v, want %v", h, uint64(1))
		return
	}
}

func TestHeadsDelete(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c := newRandomCID()
	err := heads.write(ctx, heads.store, c, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
		return
	}

	err = heads.delete(ctx, heads.store, c)
	if err != nil {
		t.Error("Failed to delete from head set:", err)
		return
	}

	_, err = heads.load(ctx, c)
	if err != ds.ErrNotFound {
		t.Error("failed to delete from head set, value still set")
		return
	}
}

func TestHeadsIsHead(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c := newRandomCID()
	err := heads.write(ctx, heads.store, c, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
		return
	}

	ishead, h, err := heads.IsHead(ctx, c)
	if err != nil {
		t.Error("Failedd to check isHead:", err)
		return
	}

	if ishead == false {
		t.Error("Expected isHead to return true, instead false")
		return
	}

	if h != uint64(1) {
		t.Errorf("Incorrect height value from isHead, have %v, want %v", h, uint64(1))
		return
	}
}

func TestHeadsLen(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c := newRandomCID()
	err := heads.write(ctx, heads.store, c, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
		return
	}

	l, err := heads.Len(ctx)
	if err != nil {
		t.Error("Failed to get head set length:", err)
		return
	}

	if l != 1 {
		t.Errorf("Incorrect length for head set, have %v, want %v", l, 1)
		return
	}

	c = newRandomCID()
	err = heads.write(ctx, heads.store, c, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
		return
	}

	l, err = heads.Len(ctx)
	if err != nil {
		t.Error("Failed to get head set length (second call):", err)
		return
	}

	if l != 2 {
		t.Errorf("Incorrect length for head set, have %v, want %v", l, 2)
		return
	}
}

func TestHeadsReplaceEmpty(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c1 := newRandomCID()
	c2 := newRandomCID()
	err := heads.Replace(ctx, c1, c2, uint64(3))
	if err != nil {
		t.Error("Failed to Replace items in head set:", err)
		return
	}

	h, err := heads.load(ctx, c2)
	if err != nil {
		t.Error("Failed to load items in head set:", err)
		return
	}

	if h != uint64(3) {
		t.Errorf("Invalid value for replaced head element, have %v, want %v", h, uint64(3))
		return
	}
}

func TestHeadsReplaceNonEmpty(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c1 := newRandomCID()
	err := heads.write(ctx, heads.store, c1, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
		return
	}

	c2 := newRandomCID()
	err = heads.Replace(ctx, c1, c2, uint64(3))
	if err != nil {
		t.Error("Failed to Replace items in head set:", err)
		return
	}

	h, err := heads.load(ctx, c2)
	if err != nil {
		t.Error("Failed to load items in head set:", err)
		return
	}

	if h != uint64(3) {
		t.Errorf("Invalid value for replaced head element, have %v, want %v", h, uint64(3))
		return
	}
}

// this test is largely unneeded from a functional point of view
// since its just a wrapper around the heads.load function, however,
// Add() is an exposed function so to ensure the public API doesnt
// break we'll include it
func TestHeadsAdd(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c1 := newRandomCID()
	err := heads.Add(ctx, c1, uint64(1))
	if err != nil {
		t.Error("Failed to Add element to head set:", err)
		return
	}
}

func TestHeaddsList(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c1 := newRandomCID()
	c2 := newRandomCID()
	heads.Add(ctx, c1, uint64(1))
	heads.Add(ctx, c2, uint64(2))

	list, h, err := heads.List(ctx)
	if err != nil {
		t.Error("Failed to List head set:", err)
		return
	}

	cids := []cid.Cid{c1, c2}
	sort.Slice(cids, func(i, j int) bool {
		ci := cids[i].Bytes()
		cj := cids[j].Bytes()
		return bytes.Compare(ci, cj) < 0
	})

	if !reflect.DeepEqual(cids, list) {
		t.Errorf("Invalid set returned from List, have %v, want %v", list, cids)
		return
	}

	if h != uint64(2) {
		t.Errorf("Invalid max height from List, have %v, want %v", h, uint64(2))
		return
	}

}
