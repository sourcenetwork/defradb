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

	"github.com/sourcenetwork/defradb/core"
	ccid "github.com/sourcenetwork/defradb/core/cid"
)

func newRandomCID() cid.Cid {
	// And then feed it some data
	bs := make([]byte, 4)
	i := rand.Uint32()
	binary.LittleEndian.PutUint32(bs, i)

	c, err := ccid.NewSHA256CidV1(bs)
	if err != nil {
		return cid.Undef
	}

	return c
}

func newHeadSet() *heads {
	ds := newDS()

	return NewHeadSet(
		ds,
		core.HeadStoreKey{}.WithDocKey("mydockey").WithFieldId("1"),
	)
}

func TestHeadsWrite(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c := newRandomCID()
	err := heads.Write(ctx, c, uint64(1))
	if err != nil {
		t.Error("Failed to write to head set:", err)
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
}

func TestHeadsReplaceNonEmpty(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c1 := newRandomCID()
	err := heads.Write(ctx, c1, uint64(1))
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
}

// this test is largely unneeded from a functional point of view
// since its just a wrapper around the heads.load function, however,
// Add() is an exposed function so to ensure the public API doesnt
// break we'll include it
func TestHeadsAdd(t *testing.T) {
	ctx := context.Background()
	heads := newHeadSet()
	c1 := newRandomCID()
	err := heads.Write(ctx, c1, uint64(1))
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
	heads.Write(ctx, c1, uint64(1))
	heads.Write(ctx, c2, uint64(2))

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
