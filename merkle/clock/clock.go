package clock

import (
	"github.com/pkg/errors"

	"github.com/sourcenetwork/defradb/core"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	ipld "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log"
)

var (
	log     = logging.Logger("defradb.merkle.clock")
	headsNS = "h"
)

type MerkleClock struct {
	headstore core.DSReaderWriter
	dagstore  core.DAGStore
	// daySyncer
	headset *heads
	crdt    core.ReplicatedData
}

// NewMerkleClock returns a new merkle clock to read/write events (deltas) to
// the clock
func NewMerkleClock(headstore core.DSReaderWriter, dagstore core.DAGStore, id string, crdt core.ReplicatedData) core.MerkleClock {
	return &MerkleClock{
		headstore: headstore,
		dagstore:  dagstore,
		headset:   newHeadset(headstore, ds.NewKey(id)), //TODO: Config logger param package wide
		crdt:      crdt,
	}
}

func (mc *MerkleClock) putBlock(heads []cid.Cid, height uint64, delta core.Delta) (ipld.Node, error) {
	if delta != nil {
		delta.SetPriority(height)
	}

	node, err := makeNode(delta, heads)
	if err != nil {
		return nil, errors.Wrap(err, "error creating block")
	}

	// @todo Add a DagSyncer instance to the MerkleCRDT structure
	// @body At the moment there is no configured DagSyncer so MerkleClock
	// blocks are not persisted into the database.
	// The following is an example implementation of how it might work:
	//
	// ctx := context.Background()
	// err = mc.store.dagSyncer.Add(ctx, node)
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "error writing new block %s", node.Cid())
	// }
	err = mc.dagstore.Put(node)
	if err != nil {
		return nil, errors.Wrapf(err, "error writing new block %s", node.Cid())
	}

	return node, nil
}

// @todo Change AddDAGNode to AddDelta

// AddDAGNode adds a new delta to the existing DAG for this MerkleClock
// It checks the current heads, sets the delta priority in the merkle dag
// adds it to the blockstore the runs ProcessNode
func (mc *MerkleClock) AddDAGNode(delta core.Delta) (cid.Cid, error) {
	heads, height, err := mc.headset.List()
	if err != nil {
		return cid.Undef, errors.Wrap(err, "error getting heads")
	}
	height = height + 1

	delta.SetPriority(height)

	// write the delta and heads to a new block
	nd, err := mc.putBlock(heads, height, delta)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "Error adding block")
	}

	// apply the new node and merge the delta with state
	// @todo Remove NodeGetter as a paramter, and move it to a MerkleClock field
	_, err = mc.ProcessNode(
		&crdtNodeGetter{deltaExtractor: mc.crdt.DeltaDecode},
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

// ProcessNode processes an already merged delta into a crdt
// by
func (mc *MerkleClock) ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error) {
	current := node.Cid()
	err := mc.crdt.Merge(delta, dshelp.CidToDsKey(current).String())
	if err != nil {
		return nil, errors.Wrapf(err, "error merging delta from %s", current)
	}

	// if prio := delta.GetPriority(); prio%10 == 0 {
	// 	store.logger.Infof("merged delta from %s (priority: %d)", current, prio)
	// } else {
	// 	store.logger.Debugf("merged delta from %s (priority: %d)", current, prio)
	// }

	links := node.Links()
	if len(links) == 0 { // reached the bottom, at a leaf
		err := mc.headset.Add(root, rootPrio)
		if err != nil {
			return nil, errors.Wrapf(err, "error adding head %s", root)
		}
		return nil, nil
	}

	children := []cid.Cid{}

	for _, l := range links {
		child := l.Cid
		isHead, _, err := mc.headset.IsHead(child)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking if %s is head", child)
		}

		if isHead {
			// reached one of the current heads, replace it with the tip
			// of current branch
			err := mc.headset.Replace(child, root, rootPrio)
			if err != nil {
				return nil, errors.Wrapf(err, "error replacing head: %s->%s", child, root)
			}

			continue
		}

		known, err := mc.dagstore.Has(child)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking for know block %s", child)
		}
		if known {
			// we reached a non-head node in the known tree.
			// This means our root block is a new head
			mc.headset.Add(root, rootPrio)
			continue
		}

		children = append(children, child)
	}

	return children, nil
}

func (mc *MerkleClock) Heads() *heads {
	return mc.headset
}
