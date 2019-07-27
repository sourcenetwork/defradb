package crdt

import (
	"github.com/pkg/errors"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/merkle/clock"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
)

type MerkleCRDT interface {
	core.ReplicatedData
	core.MerkleClock
	WithStore(ds.Datastore)
	WithNS(ds.Key)
	ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error)
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

type baseMerkleCRDT struct {
	clock clock.MerkleClock
	crdt  core.ReplicatedData
}

func (base *baseMerkleCRDT) Merge(other core.Delta, id string) error {
	return base.crdt.Merge(other, id)
}

func (base *baseMerkleCRDT) ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error) {
	current := node.Cid()
	err := base.Merge(delta, dshelp.CidToDsKey(current).String())
	if err != nil {
		return nil, errors.Wrapf(eff, "error merging delta from %s", current)
	}

	return base.clock.ProcessNode(ng, root, rootPrio, delta, node)
}

// Publishes the delta to state
func (base *baseMerkleCRDT) Publish(delta core.Delta) (cid.Cid, error) {
	return base.addDAGNode(delta)
}

func (base *baseMerkleCRDT) addDAGNode(delta core.Delta) (cid.Cid, error) {
	heads, height, err := base.clock.Heads.List()
	if err != nil {
		return cid.Undef, errors.Wrap(err, "error getting heads")
	}
	height = height + 1

	delta.SetPriority(height)

	// write the delta and heads to a new block
	nd, err := base.putBlock(heads, height, delta)
	if err != nil {
		return cid.Undef
	}

	// apply the new node and merge the delta with state
	_, err = base.ProcessNode(
		&merkleclock.NodeGetter{base.store.dagSynce},
		nd.Cid(),
		height,
		delta,
		nd,
	)

	if err != nil {
		return cid.Undef, errors.Wrap(err, "error processing new block")
	}
	return nd.Cid(), nil
}
