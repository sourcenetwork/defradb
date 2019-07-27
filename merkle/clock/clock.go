package clock

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sourcenetwork/defradb/core"

	cid "github.com/ipfs/go-cid"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	ipld "github.com/ipfs/go-ipld-format"
)

type merkleClock struct {
	store          ds.Datstore
	heads          *heads
	crdt           core.ReplicatedData
	extractDeltaFn func(ipld.Node) core.Delta
}

// NewMerkleClock returns a new merkle clock to read/write events (deltas) to
// the clock
func NewMerkleClock(store ds.Datastore, ns ds.Key, deltaFn func(ipld.Node) core.Delta) core.MerkleClock {
	return nil
}

func (mc *merkleClock) putBlock(heads []cid.Cid, height uint64, delta core.Delta) (ipld.Node, error) {
	if delta != nil {
		delta.SetPriority(height)
	}

	node, err := makeNode(delta, heads)
	if err != nil {
		return nil, errors.Wrap(err, "error creating block")
	}

	ctx := context.Background()
	err = mc.store.dagSyncer.Add(ctx, node)
	if err != nil {
		return nil, errors.Wrapf(err, "error writing new block %s", node.Cid())
	}

	return node, nil
}

func (mc *merkleClock) AddDAGNode(delta core.Delta) (cid.Cid, error) {
	heads, height, err := mc.Heads.List()
	if err != nil {
		return cid.Undef, errors.Wrap(err, "error getting heads")
	}
	height = height + 1

	delta.SetPriority(height)

	// write the delta and heads to a new block
	nd, err := mc.putBlock(heads, height, delta)
	if err != nil {
		return cid.Undef
	}

	// apply the new node and merge the delta with state
	_, err = mc.ProcessNode(
		&NodeGetter{mc.store.dagSyncer, mc.extractDeltaFn},
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
func (mc *merkleclock) ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error) {
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
		err := store.heads.Add(root, rootPrio)
		if err != nil {
			return nil, errors.Wrapf(err, "error adding head %s", root)
		}
		return nil, nil
	}

	children := []cid.Cid{}

	for _, l := range links {
		child := l.Cid
		isHead, _, err := mc.headsIsHead(child)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking if %s is head", child)
		}

		if isHead {
			// reached one of the current heads, replace it with the tip
			// of current branch
			err := mc.heads.Replace(child, root, rootPrio)
			if err != nil {
				return nil, errors.Wrapf(err, "error replacing head: %s->%s", child, root)
			}

			continue
		}

		known, err := mc.store.dagSyncer.HasBlock(child)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking for know block %s", child)
		}
		if known {
			// we reached a non-head node in the known tree.
			// This means our root block is a new head
			mc.heads.Add(root, rootPrio)
			continue
		}

		children = append(children, child)
	}

	return children, nil
}
