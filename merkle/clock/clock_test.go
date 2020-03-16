package clock

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore/namespace"

	"github.com/sourcenetwork/defradb/core/crdt"
	dagstore "github.com/sourcenetwork/defradb/store"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	mh "github.com/multiformats/go-multihash"
)

// var ()

func newDS() ds.Datastore {
	return ds.NewMapDatastore()
}

func newTestMerkleClock() *merkleClock {
	log := logging.Logger("defrabd.tests.clock")
	ns := ds.NewKey("/test/db")
	store := newDS()
	// datastore := namespace.Wrap(store, ns.ChildString("data"))
	headstore := namespace.Wrap(store, ns.ChildString("heads"))
	batchStore := namespace.Wrap(store, ds.NewKey("blockstore"))
	dagstore := dagstore.NewDAGStore(batchStore)
	id := "mydockey"
	reg := crdt.NewLWWRegister(store, ns, id)
	return NewMerkleClock(headstore, dagstore, id, reg, crdt.LWWRegDeltaExtractorFn, log).(*merkleClock)
}

func TestNewMerkleClock(t *testing.T) {
	log := logging.Logger("defrabd.tests.clock")
	ns := ds.NewKey("/test/db")
	store := newDS()
	// datastore := namespace.Wrap(store, ns.ChildString("data"))
	headstore := namespace.Wrap(store, ns.ChildString("heads"))
	batchStore := namespace.Wrap(store, ds.NewKey("blockstore"))
	dagstore := dagstore.NewDAGStore(batchStore)
	id := "mydockey"
	reg := crdt.NewLWWRegister(store, ns, id)
	clk := NewMerkleClock(headstore, dagstore, id, reg, crdt.LWWRegDeltaExtractorFn, log).(*merkleClock)

	if clk.store != headstore {
		t.Error("MerkleClock store not correctly set")
	} else if clk.heads.store == nil {
		t.Error("MerkleClock head set not correctly set")
	} else if clk.crdt == nil {
		t.Error("MerkleClock CRDT not correctly set")
	} else if clk.extractDeltaFn == nil {
		t.Error("MerkleClock DeltaFn not correctly set")
	}
}

func TestMerkleClockPutBlock(t *testing.T) {
	clk := newTestMerkleClock()
	delta := &crdt.LWWRegDelta{
		Data: []byte("test"),
	}
	node, err := clk.putBlock(nil, 0, delta)
	if err != nil {
		t.Errorf("Failed to putBlock, err: %v", err)
	}

	if len(node.Links()) != 0 {
		t.Errorf("Node links should be empty. Have %v, want %v", len(node.Links()), 0)
		return
	}

	// @todo Add DagSyncer check to tests
	// @body Once we add the DagSyncer to the MerkleClock implementation it needs to be
	// tested as well here.
}

func TetMerkleClockPutBlockWithHeads(t *testing.T) {
	clk := newTestMerkleClock()
	delta := &crdt.LWWRegDelta{
		Data: []byte("test"),
	}
	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}

	// And then feed it some data
	c, err := pref.Sum([]byte("Hello World!"))
	if err != nil {
		t.Error("Failed to create new head CID:", err)
		return
	}
	heads := []cid.Cid{c}
	node, err := clk.putBlock(heads, 0, delta)
	if err != nil {
		t.Error("Failed to putBlock with heads:", err)
		return
	}

	if len(node.Links()) != 1 {
		t.Errorf("putBlock has incorrect number of heads. Have %v, want %v", len(node.Links()), 1)
	}
}

func TestMerkleClockAddDAGNode(t *testing.T) {
	clk := newTestMerkleClock()
	delta := &crdt.LWWRegDelta{
		Data: []byte("test"),
	}

	c, err := clk.AddDAGNode(delta)
	if err != nil {
		t.Error("Failed to add dag node:", err)
		return
	}

	t.Log("Added Delta CID:", c)
}

func TestMerkleClockAddDAGNodeWithHeads(t *testing.T) {
	clk := newTestMerkleClock()
	delta := &crdt.LWWRegDelta{
		Data: []byte("test1"),
	}

	_, err := clk.AddDAGNode(delta)
	if err != nil {
		t.Error("Failed to add dag node:", err)
		return
	}

	delta2 := &crdt.LWWRegDelta{
		Data: []byte("test2"),
	}

	_, err = clk.AddDAGNode(delta2)
	if err != nil {
		t.Error("Failed to add second dag node with err:", err)
		return
	}

	// fmt.Println(delta.GetPriority())
	// fmt.Println(delta2.GetPriority())
	if delta.GetPriority() != 1 && delta2.GetPriority() != 2 {
		t.Errorf("AddDAGNOde failed with incorrect delta priority vals, want (%v) (%v), have (%v) (%v)", 1, 2, delta.GetPriority(), delta2.GetPriority())
	}

	// check if lww state is correct (val is test2)
	// check if head/blockstore state is correct (one head, two blocks)
	nHeads, err := clk.heads.Len()
	if err != nil {
		t.Error("Error getting MerkleClock heads size:", err)
		return
	}
	if nHeads != 1 {
		t.Errorf("Incorrect number of heads of current clock state, have %v, want %v", nHeads, 1)
		return
	}

	numBlocks := 0
	cids, err := clk.dagstore.Blockstore().AllKeysChan(context.Background())
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
