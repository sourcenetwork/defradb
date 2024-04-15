// Copyright 2022 Democratized Data Foundation
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
	"context"

	dag "github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	format "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/errors"
	merklecrdt "github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

var (
	// interface check
	_ Fetcher = (*VersionedFetcher)(nil)
)

// HistoryFetcher is like the normal DocumentFetcher, except it is able to traverse
// to a specific version in the documents history graph, and return the fetched
// state at that point exactly.
//
// Given the following Document state graph:
// {} --> V1 --> V2 --> V3 --> V4
//
//		  ^					   ^
//		  |					   |
//	Target Version		 Current State
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
// traversal query request, on a per object basis. This should be a basic map based
// ds.Datastore, abstracted into a DSReaderWriter.
//
// The goal of the VersionedFetcher is to implement the same external API/Interface as
// the DocumentFetcher, and to have it return the encoded/decoded document as
// defined in the version, so that it can be used as a drop in replacement within
// the scanNode request planner system.
//
// Current limitations:
// - We can only return a single record from an VersionedFetcher instance.
// - We can't request related sub objects (at the moment, as related objects
// ids aren't in the state graphs.
// - Probably more...
//
// Future optimizations:
// - Incremental checkpoint/snapshotting
// - Reverse traversal (starting from the current state, and working backwards)
// - Create an efficient memory store for in-order traversal (BTree, etc)
//
// Note: Should we transition this state traversal into the CRDT objects themselves, and not
// within a new fetcher?
type VersionedFetcher struct {
	// embed the regular doc fetcher
	*DocumentFetcher

	txn datastore.Txn
	ctx context.Context

	// Transient version store
	root  datastore.RootStore
	store datastore.Txn

	dsKey   core.DataStoreKey
	version cid.Cid

	queuedCids *list.List

	acp immutable.Option[acp.ACP]

	col client.Collection
	// @todo index  *client.IndexDescription
	mCRDTs map[uint32]merklecrdt.MerkleCRDT
}

// Init initializes the VersionedFetcher.
func (vf *VersionedFetcher) Init(
	ctx context.Context,
	identity acpIdentity.Identity,
	txn datastore.Txn,
	acp immutable.Option[acp.ACP],
	col client.Collection,
	fields []client.FieldDefinition,
	filter *mapper.Filter,
	docmapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	vf.acp = acp
	vf.col = col
	vf.queuedCids = list.New()
	vf.mCRDTs = make(map[uint32]merklecrdt.MerkleCRDT)
	vf.txn = txn

	// create store
	root := memory.NewDatastore(ctx)
	vf.root = root

	var err error
	vf.store, err = datastore.NewTxnFrom(
		ctx,
		vf.root,
		// We can take the parent txn id here
		txn.ID(),
		false,
	) // were going to discard and nuke this later
	if err != nil {
		return err
	}

	// run the DF init, VersionedFetchers only supports the Primary (0) index
	vf.DocumentFetcher = new(DocumentFetcher)
	return vf.DocumentFetcher.Init(
		ctx,
		identity,
		vf.store,
		acp,
		col,
		fields,
		filter,
		docmapper,
		reverse,
		showDeleted,
	)
}

// Start serializes the correct state according to the Key and CID.
func (vf *VersionedFetcher) Start(ctx context.Context, spans core.Spans) error {
	if vf.col == nil {
		return client.NewErrUninitializeProperty("VersionedFetcher", "CollectionDescription")
	}

	if len(spans.Value) != 1 {
		return ErrSingleSpanOnly
	}

	// For the VersionedFetcher, the spans needs to be in the format
	// Span{Start: DocID, End: CID}
	dk := spans.Value[0].Start()
	cidRaw := spans.Value[0].End()
	if dk.DocID == "" {
		return client.NewErrUninitializeProperty("Spans", "DocID")
	} else if cidRaw.DocID == "" { // todo: dont abuse DataStoreKey/Span like this!
		return client.NewErrUninitializeProperty("Spans", "CID")
	}

	// decode cidRaw from core.Key to cid.Cid
	// need to remove '/' prefix from the core.Key

	c, err := cid.Decode(cidRaw.DocID)
	if err != nil {
		return NewErrFailedToDecodeCIDForVFetcher(err)
	}

	vf.ctx = ctx
	vf.dsKey = dk
	vf.version = c

	if err := vf.seekTo(vf.version); err != nil {
		return NewErrFailedToSeek(c, err)
	}

	return vf.DocumentFetcher.Start(ctx, core.Spans{})
}

// Rootstore returns the rootstore of the VersionedFetcher.
func (vf *VersionedFetcher) Rootstore() ds.Datastore {
	return vf.root
}

// Start a fetcher with the needed info (cid embedded in a span)

/*
1. Init with DocID (VersionedFetched is scoped to a single doc)
2. - Create transient stores (head, data, block)
3. Start with a given Txn and CID span set (length 1 for now)
4. call traverse with the target cid
5.

err := VersionFetcher.Start(txn, spans) {
	vf.traverse(cid)
}
*/

// SeekTo exposes the private seekTo.
func (vf *VersionedFetcher) SeekTo(ctx context.Context, c cid.Cid) error {
	err := vf.seekTo(c)
	if err != nil {
		return err
	}

	return vf.DocumentFetcher.Start(ctx, core.Spans{})
}

// seekTo seeks to the given CID version by stepping through the CRDT state graph from the beginning
// to the target state, creating the serialized state at the given version. It starts by seeking
// to the closest existing state snapshot in the transient Versioned stores, which on the first
// run is 0. It seeks by iteratively jumping through the state graph via the `_head` link.
func (vf *VersionedFetcher) seekTo(c cid.Cid) error {
	// reinit the queued cids list
	vf.queuedCids = list.New()

	// recursive step through the graph
	err := vf.seekNext(c, true)
	if err != nil {
		return err
	}

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
	for ccv := vf.queuedCids.Front(); ccv != nil; ccv = ccv.Next() {
		cc, ok := ccv.Value.(cid.Cid)
		if !ok {
			return client.NewErrUnexpectedType[cid.Cid]("queueudCids", ccv.Value)
		}
		err := vf.merge(cc)
		if err != nil {
			return NewErrFailedToMergeState(err)
		}
	}

	// we now have all the required state stored
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
func (vf *VersionedFetcher) seekNext(c cid.Cid, topParent bool) error {
	// check if cid block exists in the global store, handle err

	// @todo: Find an efficient way to determine if a CID is a member of a
	// DocID State graph
	// @body: We could possibly append the DocID to the CID either as a
	// child key, or an instance on the CID key.

	hasLocalBlock, err := vf.store.DAGstore().Has(vf.ctx, c)
	if err != nil {
		return NewErrVFetcherFailedToFindBlock(err)
	}
	// skip if we already have it locally
	if hasLocalBlock {
		return nil
	}

	blk, err := vf.txn.DAGstore().Get(vf.ctx, c)
	if err != nil {
		return NewErrVFetcherFailedToGetBlock(err)
	}

	// store the block in the local (transient store)
	if err := vf.store.DAGstore().Put(vf.ctx, blk); err != nil {
		return NewErrVFetcherFailedToWriteBlock(err)
	}

	// add the CID to the queuedCIDs list
	if topParent {
		vf.queuedCids.PushFront(c)
	}

	// decode the block
	nd, err := dag.DecodeProtobuf(blk.RawData())
	if err != nil {
		return NewErrVFetcherFailedToDecodeNode(err)
	}

	// subDAGLinks := make([]cid.Cid, 0) // @todo: set slice size
	l, err := nd.GetNodeLink(core.HEAD)
	// ErrLinkNotFound is fine, it just means we have no more head links
	if err != nil && !errors.Is(err, dag.ErrLinkNotFound) {
		return NewErrVFetcherFailedToGetDagLink(err)
	}

	// only seekNext on parent if we have a HEAD link
	if !errors.Is(err, dag.ErrLinkNotFound) {
		err := vf.seekNext(l.Cid, true)
		if err != nil {
			return err
		}
	}

	// loop over links and ignore head links
	for _, l := range nd.Links() {
		if l.Name == core.HEAD {
			continue
		}

		err := vf.seekNext(l.Cid, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// merge in the state of the IPLD Block identified by CID c into the VersionedFetcher state.
// Requires the CID to already exist in the DAGStore.
// This function only works for merging Composite MerkleCRDT objects.
//
// First it checks for the existence of the block,
// then extracts the delta object and priority from the block
// gets the existing MerkleClock instance, or creates one.
//
// Currently we assume the CID is a CompositeDAG CRDT node.
func (vf *VersionedFetcher) merge(c cid.Cid) error {
	// get node
	nd, err := vf.getDAGNode(c)
	if err != nil {
		return err
	}

	// first arg 0 is the index for the composite DAG in the mCRDTs cache
	if err := vf.processNode(0, nd, client.COMPOSITE, client.FieldKind_None, ""); err != nil {
		return err
	}

	// handle subgraphs
	// loop over links and ignore head links
	for _, l := range nd.Links() {
		if l.Name == core.HEAD {
			continue
		}

		// get node
		subNd, err := vf.getDAGNode(l.Cid)
		if err != nil {
			return err
		}

		field, ok := vf.col.Definition().GetFieldByName(l.Name)
		if !ok {
			return client.NewErrFieldNotExist(l.Name)
		}
		if err := vf.processNode(uint32(field.ID), subNd, field.Typ, field.Kind, l.Name); err != nil {
			return err
		}
	}

	return nil
}

func (vf *VersionedFetcher) processNode(
	crdtIndex uint32,
	nd format.Node,
	ctype client.CType,
	kind client.FieldKind,
	fieldName string,
) (err error) {
	// handle CompositeDAG
	mcrdt, exists := vf.mCRDTs[crdtIndex]
	if !exists {
		dsKey, err := base.MakePrimaryIndexKeyForCRDT(vf.col.Definition(), ctype, vf.dsKey, fieldName)
		if err != nil {
			return err
		}
		mcrdt, err = merklecrdt.InstanceWithStore(
			vf.store,
			core.CollectionSchemaVersionKey{},
			ctype,
			kind,
			dsKey,
			fieldName,
		)
		if err != nil {
			return err
		}
		vf.mCRDTs[crdtIndex] = mcrdt
		// compositeClock = compMCRDT
	}

	delta, err := mcrdt.DeltaDecode(nd)
	if err != nil {
		return err
	}

	err = mcrdt.Clock().ProcessNode(vf.ctx, delta, nd)
	return err
}

func (vf *VersionedFetcher) getDAGNode(c cid.Cid) (*dag.ProtoNode, error) {
	// get Block
	blk, err := vf.store.DAGstore().Get(vf.ctx, c)
	if err != nil {
		return nil, NewErrFailedToGetDagNode(err)
	}

	// get node
	// decode the block
	return dag.DecodeProtobuf(blk.RawData())
}

// Close closes the VersionedFetcher.
func (vf *VersionedFetcher) Close() error {
	if err := vf.root.Close(); err != nil {
		return err
	}

	return vf.DocumentFetcher.Close()
}

// NewVersionedSpan creates a new VersionedSpan from a DataStoreKey and a version CID.
func NewVersionedSpan(dsKey core.DataStoreKey, version cid.Cid) core.Spans {
	// Todo: Dont abuse DataStoreKey for version cid!
	return core.NewSpans(core.NewSpan(dsKey, core.DataStoreKey{DocID: version.String()}))
}
