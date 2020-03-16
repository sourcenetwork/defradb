package crdt

import (
	"fmt"
	"testing"

	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
	dagstore "github.com/sourcenetwork/defradb/store"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log"
	"github.com/sourcenetwork/defradb/core/crdt"
)

var (
	log = logging.Logger("defrabd.tests.merklecrdt")
	// store ds.Datastore
)

func newDS() ds.Datastore {
	return ds.NewMapDatastore()
}

func newTestBaseMerkleCRDT() (*baseMerkleCRDT, ds.Datastore) {
	ns := ds.NewKey("/test/db")
	store := newDS()
	datastore := namespace.Wrap(store, ns.ChildString("data"))
	headstore := namespace.Wrap(store, ns.ChildString("heads"))
	batchStore := namespace.Wrap(store, ds.NewKey("blockstore"))
	dagstore := dagstore.NewDAGStore(batchStore)

	id := "MyKey"
	reg := corecrdt.NewLWWRegister(datastore, ns, id)
	clk := clock.NewMerkleClock(headstore, dagstore, id, reg, crdt.LWWRegDeltaExtractorFn, log)
	return &baseMerkleCRDT{clk, reg}, store
}

func TestMerkleCRDTPublish(t *testing.T) {
	bCRDT, store := newTestBaseMerkleCRDT()
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

	printStore(store)
}

func printStore(store ds.Datastore) {
	q := query.Query{
		Prefix:   "",
		KeysOnly: false,
	}

	results, err := store.Query(q)
	defer results.Close()
	if err != nil {
		panic(err)
	}

	for r := range results.Next() {
		fmt.Println(r.Key, ": ", r.Value)
	}
}
