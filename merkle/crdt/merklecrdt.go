package crdt

import (
	"github.com/sourcenetwork/defradb/core"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

type MerkleCRDT interface {
	core.ReplicatedData
	// core.MerkleClock
	// WithStore(ds.Datastore)
	// WithNS(ds.Key)
	// ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error)
	// NewObject() error
}

type MerkleCRDTInitFn func(ds.Key) MerkleCRDT
type MerkleCRDTFactory func(store ds.Datastore, namespace ds.Key) MerkleCRDTInitFn

type Type byte

const (
	LWW_REGISTER = Type(iota)
)

var (
	defaultMerkleCRDTs = make(map[Type]MerkleCRDTFactory)
)

// The baseMerkleCRDT handles the merkle crdt overhead functions
// that aren't CRDT specific like the mutations and state retrieval
// functions. It handles creating and publishing the crdt DAG with
// the help of the MerkleClock
type baseMerkleCRDT struct {
	clock core.MerkleClock
	crdt  core.ReplicatedData
}

func (base *baseMerkleCRDT) Merge(other core.Delta, id string) error {
	return base.crdt.Merge(other, id)
}

// func (base *baseMerkleCRDT) ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error) {
// 	current := node.Cid()
// 	err := base.Merge(delta, dshelp.CidToDsKey(current).String())
// 	if err != nil {
// 		return nil, errors.Wrapf(eff, "error merging delta from %s", current)
// 	}

// 	return base.clock.ProcessNode(ng, root, rootPrio, delta, node)
// }

// Publishes the delta to state
func (base *baseMerkleCRDT) Publish(delta core.Delta) (cid.Cid, error) {
	return base.clock.AddDAGNode(delta)
	// and broadcast
}
