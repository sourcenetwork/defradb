// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package clock

import (
	"context"
	"fmt"

	"github.com/sourcenetwork/defradb/core"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	ipld "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("defradb.merkle.clock")
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

func (mc *MerkleClock) putBlock(ctx context.Context, heads []cid.Cid, height uint64, delta core.Delta) (ipld.Node, error) {
	if delta != nil {
		delta.SetPriority(height)
	}

	node, err := makeNode(delta, heads)
	if err != nil {
		return nil, fmt.Errorf("error creating block : %w", err)
	}

	// @todo Add a DagSyncer instance to the MerkleCRDT structure
	// @body At the moment there is no configured DagSyncer so MerkleClock
	// blocks are not persisted into the database.
	// The following is an example implementation of how it might work:
	//
	// ctx := context.Background()
	// err = mc.store.dagSyncer.Add(ctx, node)
	// if err != nil {
	// 	return nil, fmt.Errorf("error writing new block %s : %w", node.Cid(), err)
	// }
	err = mc.dagstore.Put(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("error writing new block %s : %w", node.Cid(), err)
	}

	return node, nil
}

// @todo Change AddDAGNode to AddDelta

// AddDAGNode adds a new delta to the existing DAG for this MerkleClock
// It checks the current heads, sets the delta priority in the merkle dag
// adds it to the blockstore the runs ProcessNode
func (mc *MerkleClock) AddDAGNode(ctx context.Context, delta core.Delta) (cid.Cid, error) {
	heads, height, err := mc.headset.List(ctx)
	if err != nil {
		return cid.Undef, fmt.Errorf("error getting heads : %w", err)
	}
	height = height + 1

	delta.SetPriority(height)

	// write the delta and heads to a new block
	nd, err := mc.putBlock(ctx, heads, height, delta)
	if err != nil {
		return cid.Undef, fmt.Errorf("Error adding block : %w", err)
	}

	// apply the new node and merge the delta with state
	// @todo Remove NodeGetter as a parameter, and move it to a MerkleClock field
	_, err = mc.ProcessNode(
		ctx,
		&crdtNodeGetter{deltaExtractor: mc.crdt.DeltaDecode},
		nd.Cid(),
		height,
		delta,
		nd,
	)

	if err != nil {
		return cid.Undef, fmt.Errorf("error processing new block : %w", err)
	}
	return nd.Cid(), nil
}

// ProcessNode processes an already merged delta into a crdt
// by
func (mc *MerkleClock) ProcessNode(ctx context.Context, ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error) {
	current := node.Cid()
	err := mc.crdt.Merge(ctx, delta, dshelp.MultihashToDsKey(current.Hash()).String())
	if err != nil {
		return nil, fmt.Errorf("error merging delta from %s : %w", current, err)
	}

	// if prio := delta.GetPriority(); prio%10 == 0 {
	// 	store.logger.Infof("merged delta from %s (priority: %d)", current, prio)
	// } else {
	// 	store.logger.Debugf("merged delta from %s (priority: %d)", current, prio)
	// }

	links := node.Links()
	// check if we have any HEAD links
	hasHeads := false
	for _, l := range links {
		if l.Name == "_head" {
			hasHeads = true
			break
		}
	}
	if !hasHeads { // reached the bottom, at a leaf
		err := mc.headset.Add(ctx, root, rootPrio)
		if err != nil {
			return nil, fmt.Errorf("error adding head (when reached the bottom) %s : %w", root, err)
		}
		return nil, nil
	}

	children := []cid.Cid{}

	for _, l := range links {
		child := l.Cid
		isHead, _, err := mc.headset.IsHead(ctx, child)
		if err != nil {
			return nil, fmt.Errorf("error checking if %s is head : %w", child, err)
		}

		if isHead {
			// reached one of the current heads, replace it with the tip
			// of current branch
			err = mc.headset.Replace(ctx, child, root, rootPrio)
			if err != nil {
				return nil, fmt.Errorf("error replacing head: %s->%s : %w", child, root, err)
			}

			continue
		}

		known, err := mc.dagstore.Has(ctx, child)
		if err != nil {
			return nil, fmt.Errorf("error checking for know block %s : %w", child, err)
		}
		if known {
			// we reached a non-head node in the known tree.
			// This means our root block is a new head
			err := mc.headset.Add(ctx, root, rootPrio)
			if err != nil {
				log.Errorf("error adding head (when root is new head): %s : %w", root, err)
				// OR should this also return like below comment??
				// return nil, fmt.Errorf("error adding head (when root is new head): %s : %w", root, err)
			}

			continue
		}

		children = append(children, child)
	}

	return children, nil
}

func (mc *MerkleClock) Heads() *heads {
	return mc.headset
}
