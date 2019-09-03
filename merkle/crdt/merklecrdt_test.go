package crdt

import (
	"testing"

	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
	dagstore "github.com/sourcenetwork/defradb/store"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log"
	"github.com/sourcenetwork/defradb/core/crdt"
)

var (
	log = logging.Logger("defrabd.tests.merklecrdt")
)

func newDS() ds.Datastore {
	return ds.NewMapDatastore()
}

func newTestBaseMerkleCRDT() *baseMerkleCRDT {
	store := newDS()
	batchStore := namespace.Wrap(store, ds.NewKey("blockstore"))
	dagstore := dagstore.NewDAGStore(batchStore)
	ns := ds.NewKey("/test/db")
	id := "mydockey"
	reg := corecrdt.NewLWWRegister(store, ns, id)
	clk := clock.NewMerkleClock(store, dagstore, ns, id, reg, crdt.LWWRegDeltaExtractorFn, log)
	return &baseMerkleCRDT{clk, reg}
}

func TestMerkleCRDTPublish(t *testing.T) {
	bCRDT := newTestBaseMerkleCRDT()
	delta := &corecrdt.LWWRegDelta{
		Data: []byte("test"),
	}

	c, err := bCRDT.Publish(delta)
	if err != nil {
		t.Error("Failed to publish delta to MerkleCRDT:", err)
		return
	}

	if c == cid.Undef {
		t.Error("Published returned invalid CID Undef:", c)
		return
	}
}
