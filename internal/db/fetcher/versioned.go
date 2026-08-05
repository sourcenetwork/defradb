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
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/blockowner"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

const (
	// 1 MB, this matches the maximum badger-in-memory value size.
	//
	// Nearly at least, badger panics if this is set to it's max for reasons not yet
	// looked into.  Going one byte smaller does not have this issue.
	chunkSize = (1 << 20) - 1
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
// ds.Datastore, abstracted into a ReaderWriter.
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
	Fetcher

	txn datastore.Txn
	ctx context.Context

	// Transient version store
	root  corekv.TxnStore
	store datastore.Txn

	// Link system over the txn's encryption blockstore. Used to load encryption blocks
	// when replaying encrypted blocks during version traversal. Initialized lazily.
	encBlockLS *linking.LinkSystem

	queuedCids *list.List

	nodeACP     acpDB.NACInfo
	documentACP immutable.Option[dac.DocumentACP]

	col client.Collection
}

// Init initializes the VersionedFetcher.
func (vf *VersionedFetcher) Init(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	txn datastore.Txn,
	nodeACP acpDB.NACInfo,
	documentACP immutable.Option[dac.DocumentACP],
	index immutable.Option[client.IndexDescription],
	col client.Collection,
	fields []client.CollectionFieldDescription,
	filter *mapper.Filter,
	ordering []mapper.OrderCondition,
	docmapper *core.DocumentMapping,
	showDeleted bool,
) error {
	vf.nodeACP = nodeACP
	vf.documentACP = documentACP
	vf.col = col
	vf.queuedCids = list.New()
	vf.txn = txn

	// create store
	root := memory.NewDatastore(ctx)
	vf.root = root

	// Copy the entire system store into the temp store so that important stuff
	// such as collection definitions and short-ids are available.
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{})
	if err != nil {
		return NewErrCreateVersionIterator(err)
	}
	dst := datastore.SystemstoreFrom(root)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		if !hasValue {
			break
		}

		value, err := iter.Value()
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		err = dst.Set(ctx, iter.Key(), value)
		if err != nil {
			return errors.Join(NewErrCopyVersionedData(err), iter.Close())
		}
	}
	err = iter.Close()
	if err != nil {
		return err
	}

	vf.store = datastore.NewTxnFrom(
		vf.root,
		// Because we have created a new root, and are not operating on the actual 'main' Defra instance,
		// we should create a new lockset - the main lockset on `db` must not be used, as
		// we have zero reason to be locking that whilst operating on this temporary store.
		lock.NewLockSet(),
		// We can take the parent txn id here
		txn.ID(),
		false,
		// Chunk by default, it is a pain to figure out if it is necessary or not here, so
		// we chose to take the performance hit and chunk.
		immutable.Some(chunkSize),
	) // were going to discard and nuke this later

	// run the DF init, VersionedFetchers only supports the Primary (0) index
	vf.Fetcher = NewDocumentFetcher()
	return vf.Fetcher.Init(
		ctx,
		identity,
		vf.store,
		nodeACP,
		documentACP,
		index,
		col,
		fields,
		filter,
		ordering,
		docmapper,
		showDeleted,
	)
}

// Start serializes the correct state according to the Key and CID.
func (vf *VersionedFetcher) Start(ctx context.Context, prefixes ...keys.Walkable) error {
	if len(prefixes) == 0 {
		return ErrMissingVersionedPrefix
	}
	// The versioned fetcher is only ever given a headstore key by the planner.
	prefix, ok := prefixes[0].(keys.HeadstoreDocKey)
	if !ok {
		return client.NewErrUnexpectedType[keys.HeadstoreDocKey]("prefix", prefixes[0])
	}

	vf.ctx = ctx

	if err := vf.seekTo(prefix.Cid, prefix.DocShortID); err != nil {
		return NewErrFailedToSeek(prefix.Cid, err)
	}

	return vf.Fetcher.Start(ctx)
}

// Start a fetcher with the needed info (cid embedded in a prefix)

/*
1. Init with DocID (VersionedFetched is scoped to a single doc)
2. - Create transient stores (head, data, block)
3. Start with a given Txn and CID prefix set (length 1 for now)
4. call traverse with the target cid
5.

err := VersionFetcher.Start(txn, prefixes) {
	vf.traverse(cid)
}
*/

// seekTo seeks to the given CID version by stepping through the CRDT state graph from the beginning
// to the target state, creating the serialized state at the given version. It starts by seeking
// to the closest existing state snapshot in the transient Versioned stores, which on the first
// run is 0. It seeks by iteratively jumping through the state graph via the `_head` link.
func (vf *VersionedFetcher) seekTo(c cid.Cid, docShortID uint64) error {
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
	firstQueued := vf.queuedCids.Front()
	if docShortID == 0 && firstQueued != nil {
		cc, ok := firstQueued.Value.(cid.Cid)
		if !ok {
			return client.NewErrUnexpectedType[cid.Cid]("queueudCids", firstQueued.Value)
		}
		block, err := vf.getDAGBlock(cc)
		if err != nil {
			return err
		}
		if block.Delta.IsComposite() && len(block.Heads) == 0 {
			collectionShortID, err := id.GetCollectionShortID(vf.ctx, vf.col.Version().CollectionID)
			if err != nil {
				return err
			}
			docShortID, err = vf.docShortIDForBlock(collectionShortID, block, cc)
			if err != nil {
				return err
			}
		}
	}
	for ccv := firstQueued; ccv != nil; ccv = ccv.Next() {
		cc, ok := ccv.Value.(cid.Cid)
		if !ok {
			return client.NewErrUnexpectedType[cid.Cid]("queueudCids", ccv.Value)
		}
		err := vf.merge(cc, docShortID)
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

	hasLocalBlock, err := vf.store.Blockstore().Has(vf.ctx, c)
	if err != nil {
		return NewErrVFetcherFailedToFindBlock(err)
	}
	// skip if we already have it locally
	if hasLocalBlock {
		return nil
	}

	blk, err := vf.txn.Blockstore().Get(vf.ctx, c)
	if err != nil {
		return NewErrVFetcherFailedToGetBlock(err)
	}

	// store the block in the local (transient store)
	if err := vf.store.Blockstore().Put(vf.ctx, blk); err != nil {
		return NewErrVFetcherFailedToWriteBlock(err)
	}

	// add the CID to the queuedCIDs list
	if topParent {
		vf.queuedCids.PushFront(c)
	}

	// decode the block
	block, err := coreblock.GetFromBytes(blk.RawData())
	if err != nil {
		return NewErrVFetcherFailedToDecodeNode(err)
	}

	// only seekNext on parent if we have a HEAD link
	if len(block.Heads) != 0 {
		err := vf.seekNext(block.Heads[0].Cid, true)
		if err != nil {
			return err
		}
	}

	for _, l := range block.Links {
		err := vf.seekNext(l.Link.Cid, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// merge in the state of the IPLD Block identified by CID c into the VersionedFetcher state.
// Requires the CID to already exist in the Blockstore.
// This function only works for merging Composite MerkleCRDT objects.
//
// First it checks for the existence of the block,
// then extracts the delta object and priority from the block
// gets the existing MerkleClock instance, or creates one.
//
// Currently we assume the CID is a CompositeDAG CRDT node.
func (vf *VersionedFetcher) merge(c cid.Cid, docShortID uint64) error {
	collectionShortID, err := id.GetCollectionShortID(vf.ctx, vf.col.Version().CollectionID)
	if err != nil {
		return err
	}

	type mergeItem struct {
		cid cid.Cid
	}

	stack := make([]mergeItem, 0, 64)
	stack = append(stack, mergeItem{cid: c})

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		block, err := vf.getDAGBlock(current.cid)
		if err != nil {
			return err
		}

		block, canRead, err := coreblock.ProcessEncryptedBlock(vf.ctx, vf.getEncBlockLS(), block)
		if err != nil {
			return NewErrDecryptVersionedBlock(err, current.cid.String())
		}
		if !canRead {
			return NewErrEncryptionKeyMissing(current.cid.String())
		}

		var blockDocShortID uint64
		if !block.Delta.IsCollection() {
			blockDocShortID = docShortID
			if blockDocShortID == 0 {
				var err error
				blockDocShortID, err = vf.docShortIDForBlock(collectionShortID, block, current.cid)
				if err != nil {
					return err
				}
			}
		}

		var mcrdt crdt.ReplicatedData
		switch {
		case block.Delta.IsCollection():
			mcrdt = crdt.NewCollection(
				vf.col.Version().VersionID,
				keys.NewHeadstoreColKey(collectionShortID),
			)

		case block.Delta.IsComposite():
			mcrdt = crdt.NewDocComposite(
				vf.store.Datastore(),
				block.Delta.GetCollectionVersionID(),
				keys.DataStoreKey{
					CollectionShortID: collectionShortID,
					DocShortID:        blockDocShortID,
					FieldID:           fmt.Sprint(core.COMPOSITE_NAMESPACE),
				},
			)

		default:
			field, ok := vf.col.Version().GetFieldByName(block.Delta.GetFieldName())
			if !ok {
				return client.NewErrFieldNotExist(block.Delta.GetFieldName())
			}

			fieldShortID, err := id.GetShortFieldID(vf.ctx, collectionShortID, field.FieldID)
			if err != nil {
				return err
			}

			mcrdt, err = crdt.FieldLevelCRDTWithStore(
				vf.store.Datastore(),
				block.Delta.GetCollectionVersionID(),
				field.Typ,
				field.Kind,
				keys.DataStoreKey{
					CollectionShortID: collectionShortID,
					DocShortID:        blockDocShortID,
					FieldID:           fmt.Sprint(fieldShortID),
				},
				field.Name,
			)
			if err != nil {
				return err
			}
		}

		err = coreblock.ProcessBlock(
			vf.ctx,
			mcrdt,
			block,
			cidlink.Link{
				Cid: current.cid,
			},
		)
		if err != nil {
			return err
		}

		for i := len(block.Links) - 1; i >= 0; i-- {
			stack = append(stack, mergeItem{
				cid: block.Links[i].Cid,
			})
		}
	}

	return nil
}

func (vf *VersionedFetcher) docShortIDForBlock(
	collectionShortID uint32,
	block *coreblock.Block,
	blockCID cid.Cid,
) (uint64, error) {
	if block.Delta.IsCollection() {
		return 0, nil
	}

	owners, err := blockowner.DocIDs(
		vf.ctx,
		vf.txn.Systemstore(),
		blockCID,
	)
	if err != nil {
		return 0, err
	}

	var docID string
	switch {
	case len(owners) == 1:
		// A single-owner block (composite, or an unshared field block) belongs to that
		// document. A composite block's CID is always document-unique.
		docID = owners[0]
	case block.Delta.IsComposite() && len(block.Heads) == 0:
		docID = client.NewDocIDV0(blockCID).String()
	case block.Delta.IsComposite():
		return vf.docShortIDForCompositeHead(collectionShortID, block.Heads)
	default:
		// A field block with no single owner (shared across documents) cannot be attributed
		// to one document without a composite context; time-travel enters via composite CIDs.
		return 0, client.ErrMalformedDocID
	}

	docShortID, found, err := id.GetDocShortID(vf.ctx, collectionShortID, docID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, client.ErrMalformedDocID
	}
	return docShortID, nil
}

func (vf *VersionedFetcher) docShortIDForCompositeHead(
	collectionShortID uint32,
	heads []cidlink.Link,
) (uint64, error) {
	for _, head := range heads {
		headBlock, err := vf.getDAGBlock(head.Cid)
		if err != nil {
			return 0, err
		}
		docShortID, err := vf.docShortIDForBlock(collectionShortID, headBlock, head.Cid)
		if err != nil {
			return 0, err
		}
		if docShortID != 0 {
			return docShortID, nil
		}
	}
	return 0, client.ErrMalformedDocID
}

func (vf *VersionedFetcher) getDAGBlock(c cid.Cid) (*coreblock.Block, error) {
	// get Block
	blk, err := vf.store.Blockstore().Get(vf.ctx, c)
	if err != nil {
		return nil, NewErrFailedToGetDagNode(err)
	}

	return coreblock.GetFromBytes(blk.RawData())
}

// getEncBlockLS lazily builds (and caches) a link system over the txn's encryption
// blockstore. Used for loading encryption blocks when replaying encrypted blocks.
func (vf *VersionedFetcher) getEncBlockLS() linking.LinkSystem {
	if vf.encBlockLS == nil {
		ls := cidlink.DefaultLinkSystem()
		ls.SetReadStorage(blockstore.NewIPLDStore(vf.txn.Encstore()))
		vf.encBlockLS = &ls
	}
	return *vf.encBlockLS
}

// Close closes the VersionedFetcher.
func (vf *VersionedFetcher) Close() error {
	// vf.root may be nil if Init failed (or was never called) before
	// allocating it. Close is reachable in that state through the
	// MultiVersioned cleanup path that tracks children eagerly.
	if vf.root != nil {
		if err := vf.root.Close(); err != nil {
			return err
		}
	}

	if vf.Fetcher != nil {
		return vf.Fetcher.Close()
	}

	return nil
}
