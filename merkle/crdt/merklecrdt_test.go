package crdt

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
	"github.com/sourcenetwork/defradb/store"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log"
	"github.com/sourcenetwork/defradb/core/crdt"
)

var (
	merklecrdtlog = logging.Logger("defradb.tests.merklecrdt")
	// store core.DSReaderWriter
)

func newDS() ds.Datastore {
	return ds.NewMapDatastore()
}

func newTestBaseMerkleCRDT() (*baseMerkleCRDT, core.DSReaderWriter) {
	ns := ds.NewKey("/test/db")
	s := newDS()
	datastore := namespace.Wrap(s, ns.ChildString("data"))
	headstore := namespace.Wrap(s, ns.ChildString("heads"))
	batchStore := namespace.Wrap(s, ds.NewKey("blockstore"))
	dagstore := store.NewDAGStore(batchStore)

	id := "MyKey"
	reg := corecrdt.NewLWWRegister(datastore, ds.NewKey(""), id)
	clk := clock.NewMerkleClock(headstore, dagstore, id, reg, crdt.LWWRegDeltaExtractorFn)
	return &baseMerkleCRDT{clk, reg}, s
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

func printStore(store core.DSReaderWriter) {
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
