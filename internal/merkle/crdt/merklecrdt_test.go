// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package merklecrdt

import (
	"context"
	"testing"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	crdt "github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

func newDS() ds.Datastore {
	return ds.NewMapDatastore()
}

func newTestBaseMerkleCRDT() (*baseMerkleCRDT, datastore.DSReaderWriter) {
	s := newDS()
	multistore := datastore.MultiStoreFrom(s)

	reg := crdt.NewLWWRegister(multistore.Datastore(), core.CollectionSchemaVersionKey{}, core.DataStoreKey{}, "")
	clk := clock.NewMerkleClock(multistore.Headstore(), multistore.Blockstore(), core.HeadStoreKey{}, reg)
	return &baseMerkleCRDT{clock: clk, crdt: reg}, multistore.Rootstore()
}

func TestMerkleCRDTPublish(t *testing.T) {
	ctx := context.Background()
	bCRDT, _ := newTestBaseMerkleCRDT()
	reg := crdt.LWWRegister{}
	delta := reg.Set([]byte("test"))

	link, _, err := bCRDT.clock.AddDelta(ctx, delta)
	if err != nil {
		t.Error("Failed to publish delta to MerkleCRDT:", err)
		return
	}

	if link.Cid == cid.Undef {
		t.Error("Published returned invalid CID Undef:", link.Cid)
		return
	}
}
