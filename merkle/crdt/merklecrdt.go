package crdt

import (
	"github.com/sourcenetwork/defradb/core"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("defradb.merkle.crdt")
)

// MerkleCRDT is the implementation of a Merkle Clock along with a
// CRDT payload. It implements the ReplicatedData interface
// so it can be merged with any given semantics.
type MerkleCRDT interface {
	core.ReplicatedData
	// core.MerkleClock
	// WithStore(core.DSReaderWriter)
	// WithNS(ds.Key)
	// ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error)
	// NewObject() error
}

// type MerkleCRDTInitFn func(ds.Key) MerkleCRDT
// type MerkleCRDTFactory func(store core.DSReaderWriter, namespace ds.Key) MerkleCRDTInitFn

// Type indicates MerkleCRDT type
type Type byte

const (
	//no lint
	none = Type(iota) // reserved none type
	LWW_REGISTER
	OBJECT
)

var (
	// defaultMerkleCRDTs                     = make(map[Type]MerkleCRDTFactory)
	_ core.ReplicatedData = (*baseMerkleCRDT)(nil)
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
