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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

func newDS() datastore.DSReaderWriter {
	return memory.NewDatastore(context.TODO())
}

func newTestBaseMerkleCRDT() (*baseMerkleCRDT, datastore.DSReaderWriter) {
	s := newDS()
	multistore := datastore.MultiStoreFrom(s)

	reg := corecrdt.NewLWWRegister(multistore.Datastore(), core.CollectionSchemaVersionKey{}, core.DataStoreKey{}, "")
	clk := clock.NewMerkleClock(multistore.Headstore(), multistore.DAGstore(), core.HeadStoreKey{}, reg)
	return &baseMerkleCRDT{clock: clk, crdt: reg}, multistore.Rootstore()
}

func TestMerkleCRDTPublish(t *testing.T) {
	ctx := context.Background()
	bCRDT, store := newTestBaseMerkleCRDT()
	delta := &corecrdt.LWWRegDelta{
		Data: []byte("test"),
	}

	nd, err := bCRDT.clock.AddDAGNode(ctx, delta)
	if err != nil {
		t.Error("Failed to publish delta to MerkleCRDT:", err)
		return
	}

	if nd.Cid() == cid.Undef {
		t.Error("Published returned invalid CID Undef:", nd.Cid())
		return
	}

	printStore(ctx, store)
}

func printStore(ctx context.Context, store datastore.DSReaderWriter) {
	iter := store.Iterator(ctx, corekv.DefaultIterOptions)
	defer iter.Close(ctx)

	for ; iter.Valid(); iter.Next() {
		log.Info(ctx, "", logging.NewKV(string(iter.Key()), iter.Value()))
	}
}
