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
	"testing"

	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
	"github.com/sourcenetwork/defradb/store"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/logging"
)

func newDS() ds.Datastore {
	return ds.NewMapDatastore()
}

func newTestBaseMerkleCRDT() (*baseMerkleCRDT, core.DSReaderWriter) {
	s := newDS()
	rw := store.AsDSReaderWriter(s)
	multistore := store.MultiStoreFrom(rw)

	id := "/1/0/MyKey"
	reg := corecrdt.NewLWWRegister(multistore.Datastore(), ds.NewKey(""), id)
	clk := clock.NewMerkleClock(multistore.Headstore(), multistore.DAGstore(), id, reg)
	return &baseMerkleCRDT{clock: clk, crdt: reg}, rw
}

func TestMerkleCRDTPublish(t *testing.T) {
	ctx := context.Background()
	bCRDT, store := newTestBaseMerkleCRDT()
	delta := &corecrdt.LWWRegDelta{
		Data: []byte("test"),
	}

	c, _, err := bCRDT.Publish(ctx, delta)
	if err != nil {
		t.Error("Failed to publish delta to MerkleCRDT:", err)
		return
	}

	if c == cid.Undef {
		t.Error("Published returned invalid CID Undef:", c)
		return
	}

	printStore(ctx, store)
}

func printStore(ctx context.Context, store core.DSReaderWriter) {
	q := query.Query{
		Prefix:   "",
		KeysOnly: false,
	}

	results, err := store.Query(ctx, q)

	if err != nil {
		panic(err)
	}

	defer results.Close()

	for r := range results.Next() {
		log.Info(ctx, "", logging.NewKV(r.Key, r.Value))
	}
}
