// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package clock provides a MerkleClock implementation, to track causal ordering of events.
*/
package clock

import (
	"context"

	dshelp "github.com/ipfs/boxo/datastore/dshelp"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/logging"
)

var (
	log = logging.MustNewLogger("merkleclock")
)

// MerkleClock is a MerkleCRDT clock that can be used to read/write events (deltas) to the clock.
type MerkleClock struct {
	headstore datastore.DSReaderWriter
	dagstore  datastore.DAGStore
	// dagSyncer
	headset *heads
	crdt    core.ReplicatedData
}

// NewMerkleClock returns a new MerkleClock.
func NewMerkleClock(
	headstore datastore.DSReaderWriter,
	dagstore datastore.DAGStore,
	namespace core.HeadStoreKey,
	crdt core.ReplicatedData,
) core.MerkleClock {
	return &MerkleClock{
		headstore: headstore,
		dagstore:  dagstore,
		headset:   NewHeadSet(headstore, namespace),
		crdt:      crdt,
	}
}

func (mc *MerkleClock) putBlock(
	ctx context.Context,
	heads []cid.Cid,
	delta core.Delta,
) (ipld.Node, error) {
	node, err := makeNode(delta, heads)
	if err != nil {
		return nil, NewErrCreatingBlock(err)
	}

	// @todo Add a DagSyncer instance to the MerkleCRDT structure
	// @body At the moment there is no configured DagSyncer so MerkleClock
	// blocks are not persisted into the database.
	// The following is an example implementation of how it might work:
	//
	// ctx := context.Background()
	// err = mc.store.dagSyncer.Add(ctx, node)
	// if err != nil {
	// 	return nil, errors.Wrap("error writing new block %s ", node.Cid(), err)
	// }
	err = mc.dagstore.Put(ctx, node)
	if err != nil {
		return nil, NewErrWritingBlock(node.Cid(), err)
	}

	return node, nil
}

// @todo Change AddDAGNode to AddDelta

// AddDAGNode adds a new delta to the existing DAG for this MerkleClock: checks the current heads,
// sets the delta priority in the Merkle DAG, and adds it to the blockstore the runs ProcessNode.
func (mc *MerkleClock) AddDAGNode(
	ctx context.Context,
	delta core.Delta,
) (ipld.Node, error) {
	heads, height, err := mc.headset.List(ctx)
	if err != nil {
		return nil, NewErrGettingHeads(err)
	}
	height = height + 1

	delta.SetPriority(height)

	// write the delta and heads to a new block
	nd, err := mc.putBlock(ctx, heads, delta)
	if err != nil {
		return nil, err
	}

	// apply the new node and merge the delta with state
	// @todo Remove NodeGetter as a parameter, and move it to a MerkleClock field
	_, err = mc.ProcessNode(
		ctx,
		&CrdtNodeGetter{DeltaExtractor: mc.crdt.DeltaDecode},
		nd.Cid(),
		delta,
		nd,
	)

	return nd, err //@todo: Include raw block data in return
}

// ProcessNode processes an already merged delta into a CRDT by adding it to the state.
func (mc *MerkleClock) ProcessNode(
	ctx context.Context,
	ng core.NodeGetter,
	root cid.Cid,
	delta core.Delta,
	node ipld.Node,
) ([]cid.Cid, error) {
	current := node.Cid()
	priority := delta.GetPriority()

	log.Debug(ctx, "Running ProcessNode", logging.NewKV("CID", current))
	err := mc.crdt.Merge(ctx, delta, dshelp.MultihashToDsKey(current.Hash()).String())
	if err != nil {
		return nil, NewErrMergingDelta(current, err)
	}

	links := node.Links()
	// check if we have any HEAD links
	hasHeads := false
	log.Debug(ctx, "Stepping through node links")
	for _, l := range links {
		log.Debug(ctx, "Checking link", logging.NewKV("Name", l.Name), logging.NewKV("CID", l.Cid))
		if l.Name == "_head" {
			hasHeads = true
			break
		}
	}
	if !hasHeads { // reached the bottom, at a leaf
		log.Debug(ctx, "No heads found")
		err := mc.headset.Write(ctx, root, priority)
		if err != nil {
			return nil, NewErrAddingHead(root, err)
		}
	}

	children := []cid.Cid{}

	for _, l := range links {
		linkCid := l.Cid
		log.Debug(ctx, "Scanning for replacement heads", logging.NewKV("Child", linkCid))
		isHead, err := mc.headset.IsHead(ctx, linkCid)
		if err != nil {
			return nil, NewErrCheckingHead(linkCid, err)
		}

		if isHead {
			log.Debug(ctx, "Found head, replacing!")
			// reached one of the current heads, replace it with the tip
			// of current branch
			err = mc.headset.Replace(ctx, linkCid, root, priority)
			if err != nil {
				return nil, NewErrReplacingHead(linkCid, root, err)
			}

			continue
		}

		known, err := mc.dagstore.Has(ctx, linkCid)
		if err != nil {
			return nil, NewErrCouldNotFindBlock(linkCid, err)
		}
		if known {
			// we reached a non-head node in the known tree.
			// This means our root block is a new head
			log.Debug(ctx, "Adding head")
			err := mc.headset.Write(ctx, root, priority)
			if err != nil {
				log.ErrorE(
					ctx,
					"Failure adding head (when root is a new head)",
					err,
					logging.NewKV("Root", root),
				)
				// OR should this also return like below comment??
				// return nil, errors.Wrap("error adding head (when root is new head): %s ", root, err)
			}
			continue
		}

		children = append(children, linkCid)
	}

	return children, nil
}

// Heads returns the current heads of the MerkleClock.
func (mc *MerkleClock) Heads() *heads {
	return mc.headset
}
