// Copyright 2021 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package fetcher

import (
	"container/list"

	"github.com/sourcenetwork/defradb/core"

	"github.com/ipfs/go-cid"
	dsq "github.com/ipfs/go-datastore/query"
)

// HistoryFetcher is like the normal DocumentFetcher, except it is able to traverse
// to a specific version in the documents history graph, and return the fetched
// state at that point exactly.
//
// Given the following Document state graph
//
// {} --> V1 --> V2 --> V3 --> V4
//		  ^					   ^
//		  |					   |
// 	Target Version		 Current State
//
//
// A regular DocumentFetcher fetches and returns the state at V4, but the
// VersionsedFetcher would step backwards through the update graph, recompose
// the state at the "Target Version" V1, and return the state at that point.
//
// This is achieved by reconstructing the target state using the given MerkleCRDT
// DAG. Given the Target Version CID, we collect all the individual delta nodes
// in the MerkleDAG, until we reach the initial (genesis) state.
//
// Transient/Ephemeral datastores are intanciated for the lifetime of the
// traversal query, on a per object basis. This should be a basic map based
// ds.Datastore, abstracted into a DSReaderWriter.
//
// The goal of the VersionedFetcher is to implement the same external API/Interface as
// the DocumentFetcher, and to have it return the encoded/decoded document as
// defined in the version, so that it can be used as a drop in replacement within
// the scanNode query planner system.
//
// Current limitations:
// - We can only return a single record from an VersionedFetcher
// 	 instance.
// - We can't query into related sub objects (at the moment, as related objects
//   ids aren't in the state graphs.
// - Probably more...
//
// Future optimizations:
// - Incremental checkpoint/snapshotting
// - Reverse traversal (starting from the current state, and working backwards)
// - Create a effecient memory store for in-order traversal (BTree, etc)
type VersionedFetcher struct {
	txn   core.Txn
	spans core.Spans

	key     core.Key
	version cid.Cid

	queuedCids *list.List

	kv     *core.KeyValue
	kvIter dsq.Results
	kvEnd  bool
}

// Init

// Start

// Start a fetcher with the needed info (cid embedded in a span)

/*
1. Init with DocKey (VersionedFetched is scoped to a single doc)
2. - Create transient stores (head, data, block)
3. Start with a given Txn and CID span set (length 1 for now)
4. call traverse with the target cid
5.

err := VersionFetcher.Start(txn, spans) {
	vf.traverse(cid)
}
*/

// seekTo seeks to the given CID version by steping through the CRDT
// state graph from the beginning to the target state, creating the
// serialized state at the given version. It starts by seeking to the
// closest existing state snapshot in the transient Versioned stores,
// which on the first run is 0. It seeks by iteratively jumping through
// the state graph via the `_head` link.
func (vf *VersionedFetcher) seekTo(c cid.Cid) error {
	// recursive step through the graph
	// err := vf.seekNext(c)

	// after seekNext is completed, we have a populated
	// queuedCIDs list, and all the necessary
	// blocks in our local store

	// if we have a queuedCIDs length of 0, means we don't need
	// to do any more state serialization

	// for cid in CIDs {
	///
	/// vf.merge(cid)
	/// // Note: we need to determine what state we are "Merging"
	/// // into. This isn't necessary for the base case where we only
	/// // are concerned with generating the Versioned state for a single
	/// // CID, but for multiple CIDs, or if we reuse the transient store
	/// // as a cache, we need to swap out states to the parent of the current
	/// // CID.
	// }

	// we now have all the the required state stored
	// in our transient local Version_Index, we now need to
	// transfer it to the Primary_Index.

	// Once all values are transferred, exit with no errors
	// Any future operation can resume using the current PrimaryIndex
	// which is actually the serialized state of the CRDT graph at
	// the exact version

	return nil
}

// seekNext is the recursive iteration step of seekTo, its goal is
// to build the queuedCids list, and to transfer the required
// blocks from the global to the local store.
func (vf *VersionedFetcher) seekNext(c cid.Cid) error {
	// check if cid block exists in the global store, handle err

	// @todo: Find an effecient way to determine if a CID is a member of a
	// DocKey State graph
	// @body: We could possibly append the DocKey to the CID either as a
	// child key, or an instance on the CID key.

	/// blk, err := vf.txn.DAGstore().Get(c)

	// check if the block exists in the local (transient) store:

	// IF YES we've already processed this block, return with
	// no errors

	// IF NO:

	// add the CID to the queuedCIDs list
	// decode the block
	// nextCID = get the "_head" link target CID
	// err := vf.seekNext(nextCID)

	return nil
}
