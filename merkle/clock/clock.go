// Copyright 2022 Democratized Data Foundation
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

	cid "github.com/ipfs/go-cid"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

var (
	log = logging.MustNewLogger("defra.merkleclock")
)

type MerkleClock struct {
	headstore datastore.DSReaderWriter
	dagstore  datastore.DAGStore
	// dagSyncer
	headset *heads
	crdt    core.ReplicatedData
}

// NewMerkleClock returns a new MerkleClock to read/write events (deltas) to the clock.
func NewMerkleClock(
	headstore datastore.DSReaderWriter,
	dagstore datastore.DAGStore,
	namespace core.HeadStoreKey,
	crdt core.ReplicatedData,
) core.MerkleClock {
	return &MerkleClock{
		headstore: headstore,
		dagstore:  dagstore,
		headset:   newHeadset(headstore, namespace),
		crdt:      crdt,
	}
}

func (mc *MerkleClock) putBlock(
	ctx context.Context,
	heads []cid.Cid,
	height uint64,
	delta core.Delta,
) (ipld.Node, error) {
	if delta != nil {
		delta.SetPriority(height)
	}

	node, err := makeNode(delta, heads)
	if err != nil {
		return nil, errors.Wrap("error creating block ", err)
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
		return nil, errors.Wrap(fmt.Sprintf("error writing new block %s ", node.Cid()), err)
	}

	return node, nil
}

// @todo Change AddDAGNode to AddDelta

// AddDAGNode adds a new delta to the existing DAG for this MerkleClock.
// It checks the current heads, sets the delta priority in the merkle dag
// adds it to the blockstore the runs ProcessNode.
func (mc *MerkleClock) AddDAGNode(
	ctx context.Context,
	delta core.Delta,
) (ipld.Node, error) {
	heads, height, err := mc.headset.List(ctx)
	if err != nil {
		return nil, errors.Wrap("error getting heads ", err)
	}
	height = height + 1

	delta.SetPriority(height)

	// write the delta and heads to a new block
	nd, err := mc.putBlock(ctx, heads, height, delta)
	if err != nil {
		return nil, errors.Wrap("could not add block ", err)
	}

	// apply the new node and merge the delta with state
	// @todo Remove NodeGetter as a parameter, and move it to a MerkleClock field
	_, err = mc.ProcessNode(
		ctx,
		&CrdtNodeGetter{DeltaExtractor: mc.crdt.DeltaDecode},
		nd.Cid(),
		height,
		delta,
		nd,
	)

	if err != nil {
		return nil, errors.Wrap("error processing new block ", err)
	}
	return nd, nil //@todo: Include raw block data in return
}

// ProcessNode processes an already merged delta into a crdt
// by
func (mc *MerkleClock) ProcessNode(
	ctx context.Context,
	ng core.NodeGetter,
	root cid.Cid,
	rootPrio uint64,
	delta core.Delta,
	node ipld.Node,
) ([]cid.Cid, error) {
	current := node.Cid()
	log.Debug(ctx, "Running ProcessNode", logging.NewKV("CID", current))
	err := mc.crdt.Merge(ctx, delta, dshelp.MultihashToDsKey(current.Hash()).String())
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("error merging delta from %s ", current), err)
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
		err := mc.headset.Write(ctx, root, rootPrio)
		if err != nil {
			return nil, errors.Wrap(fmt.Sprintf("error adding head (when reached the bottom) %s ", root), err)
		}
	}

	children := []cid.Cid{}

	for _, l := range links {
		child := l.Cid
		log.Debug(ctx, "Scanning for replacement heads", logging.NewKV("Child", child))
		isHead, _, err := mc.headset.IsHead(ctx, child)
		if err != nil {
			return nil, errors.Wrap(fmt.Sprintf("error checking if %s is head ", child), err)
		}

		if isHead {
			log.Debug(ctx, "Found head, replacing!")
			// reached one of the current heads, replace it with the tip
			// of current branch
			err = mc.headset.Replace(ctx, child, root, rootPrio)
			if err != nil {
				return nil, errors.Wrap(fmt.Sprintf("error replacing head: %s->%s ", child, root), err)
			}

			continue
		}

		known, err := mc.dagstore.Has(ctx, child)
		if err != nil {
			return nil, errors.Wrap(fmt.Sprintf("error checking for known block %s ", child), err)
		}
		if known {
			// we reached a non-head node in the known tree.
			// This means our root block is a new head
			log.Debug(ctx, "Adding head")
			err := mc.headset.Write(ctx, root, rootPrio)
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

		children = append(children, child)
	}

	return children, nil
}

func (mc *MerkleClock) Heads() *heads {
	return mc.headset
}
