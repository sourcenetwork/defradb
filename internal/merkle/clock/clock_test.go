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
	"context"
	"testing"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	ccid "github.com/sourcenetwork/defradb/internal/core/cid"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

func newDS() ds.Datastore {
	return ds.NewMapDatastore()
}

func newTestMerkleClock() *MerkleClock {
	s := newDS()

	multistore := datastore.MultiStoreFrom(s)
	reg := crdt.NewLWWRegister(multistore.Rootstore(), core.CollectionSchemaVersionKey{}, core.DataStoreKey{}, "")
	return NewMerkleClock(
		multistore.Headstore(),
		multistore.DAGstore(),
		core.HeadStoreKey{DocID: request.DocIDArgName, FieldId: "1"},
		reg,
	)
}

func TestNewMerkleClock(t *testing.T) {
	s := newDS()
	multistore := datastore.MultiStoreFrom(s)
	reg := crdt.NewLWWRegister(multistore.Rootstore(), core.CollectionSchemaVersionKey{}, core.DataStoreKey{}, "")
	clk := NewMerkleClock(multistore.Headstore(), multistore.DAGstore(), core.HeadStoreKey{}, reg)

	if clk.headstore != multistore.Headstore() {
		t.Error("MerkleClock store not correctly set")
	} else if clk.headset.store == nil {
		t.Error("MerkleClock head set not correctly set")
	} else if clk.crdt == nil {
		t.Error("MerkleClock CRDT not correctly set")
	}
}

func TestMerkleClockPutBlock(t *testing.T) {
	ctx := context.Background()
	clk := newTestMerkleClock()
	reg := crdt.LWWRegister{}
	delta := reg.Set([]byte("test"))
	block := coreblock.New(delta, nil)
	_, err := clk.putBlock(ctx, block)
	if err != nil {
		t.Errorf("Failed to putBlock, err: %v", err)
	}

	if len(block.Links) != 0 {
		t.Errorf("Node links should be empty. Have %v, want %v", len(block.Links), 0)
		return
	}

	// @todo Add DagSyncer check to tests
	// @body Once we add the DagSyncer to the MerkleClock implementation it needs to be
	// tested as well here.
}

func TestMerkleClockPutBlockWithHeads(t *testing.T) {
	ctx := context.Background()
	clk := newTestMerkleClock()
	reg := crdt.LWWRegister{}
	delta := reg.Set([]byte("test"))
	c, err := ccid.NewSHA256CidV1([]byte("Hello World!"))
	if err != nil {
		t.Error("Failed to create new head CID:", err)
		return
	}
	heads := []cid.Cid{c}
	block := coreblock.New(delta, nil, heads...)
	_, err = clk.putBlock(ctx, block)
	if err != nil {
		t.Error("Failed to putBlock with heads:", err)
		return
	}

	if len(block.Links) != 1 {
		t.Errorf("putBlock has incorrect number of heads. Have %v, want %v", len(block.Links), 1)
	}
}

func TestMerkleClockAddDelta(t *testing.T) {
	ctx := context.Background()
	clk := newTestMerkleClock()
	reg := crdt.LWWRegister{}
	delta := reg.Set([]byte("test"))

	_, _, err := clk.AddDelta(ctx, delta)
	if err != nil {
		t.Error("Failed to add dag node:", err)
		return
	}
}

func TestMerkleClockAddDeltaWithHeads(t *testing.T) {
	ctx := context.Background()
	clk := newTestMerkleClock()
	reg := crdt.LWWRegister{}
	delta := reg.Set([]byte("test"))

	_, _, err := clk.AddDelta(ctx, delta)
	if err != nil {
		t.Error("Failed to add dag node:", err)
		return
	}

	reg2 := crdt.LWWRegister{}
	delta2 := reg2.Set([]byte("test2"))

	_, _, err = clk.AddDelta(ctx, delta2)
	if err != nil {
		t.Error("Failed to add second dag node with err:", err)
		return
	}

	if delta.GetPriority() != 1 && delta2.GetPriority() != 2 {
		t.Errorf(
			"AddDelta failed with incorrect delta priority vals, want (%v) (%v), have (%v) (%v)",
			1,
			2,
			delta.GetPriority(),
			delta2.GetPriority(),
		)
	}

	numBlocks := 0
	cids, err := clk.dagstore.AllKeysChan(ctx)
	if err != nil {
		t.Error("Failed to get blockstore content for merkle clock:", err)
		return
	}
	for range cids {
		numBlocks++
	}
	if numBlocks != 2 {
		t.Errorf("Incorrect number of blocks in clock state, have %v, want %v", numBlocks, 2)
		return
	}
}

// func TestMerkleClockProcessNode(t *testing.T) {
// 	t.Error("Test not implemented")
// }

// func TestMerkleClockProcessNodeWithHeads(t *testing.T) {
// 	t.Error("Test not implemented")
// }
