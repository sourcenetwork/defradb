// Copyright 2025 Democratized Data Foundation
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
package coreblock

import (
	"context"

	cid "github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func putBlock(
	ctx context.Context,
	store datastore.Blockstore,
	block interface{ GenerateNode() ipld.Node },
) (cidlink.Link, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(store))
	link, err := lsys.Store(linking.LinkContext{Ctx: ctx}, GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return cidlink.Link{}, NewErrWritingBlock(err)
	}

	return link.(cidlink.Link), nil //nolint:forcetypeassert
}

// AddDeltaOptions controls storage behavior around a CRDT delta.
type AddDeltaOptions struct {
	EncryptionDocKey []byte
}

// AddDelta adds a new delta to the existing DAG.
//
// It checks the current heads, sets the delta priority, adds it to the blockstore, then runs ProcessBlock.
func AddDelta(
	ctx context.Context,
	crdtData crdt.ReplicatedData,
	delta crdt.Delta,
	links ...DAGLink,
) (cidlink.Link, []byte, error) {
	return AddDeltaWithOptions(ctx, crdtData, delta, AddDeltaOptions{}, links...)
}

// AddDeltaWithOptions adds a delta with explicit storage behavior options.
func AddDeltaWithOptions(
	ctx context.Context,
	crdtData crdt.ReplicatedData,
	delta crdt.Delta,
	options AddDeltaOptions,
	links ...DAGLink,
) (cidlink.Link, []byte, error) {
	return addDelta(ctx, crdtData, delta, options, links...)
}

func addDelta(
	ctx context.Context,
	crdtData crdt.ReplicatedData,
	delta crdt.Delta,
	options AddDeltaOptions,
	links ...DAGLink,
) (cidlink.Link, []byte, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	headset := NewHeadSet(txn.Headstore(), crdtData.HeadstorePrefix())

	heads, height, err := headset.List(ctx)
	if err != nil {
		return cidlink.Link{}, nil, NewErrGettingHeads(err)
	}
	height = height + 1

	delta.SetPriority(height)
	block := New(crdt.NewCRDT(delta), links, heads...)

	fieldName := immutable.None[string]()
	if block.Delta.GetFieldName() != "" {
		fieldName = immutable.Some(block.Delta.GetFieldName())
	}
	var encBlock *Encryption
	var encLink cidlink.Link
	if !block.Delta.IsCollection() {
		encBlock, encLink, err = determineBlockEncryption(ctx, options.EncryptionDocKey, fieldName, heads)
		if err != nil {
			return cidlink.Link{}, nil, NewErrDetermineBlockEncryption(err)
		}
	}

	dagBlock := block
	if encBlock != nil {
		dagBlock, err = encryptBlock(ctx, block, encBlock)
		if err != nil {
			return cidlink.Link{}, nil, NewErrEncryptBlock(err)
		}
		dagBlock.Encryption = &encLink
	}

	if ok, ident := EnabledSigningFromContext(ctx); ok && ident.HasValue() {
		err = signBlock(ctx, txn.Blockstore(), dagBlock, ident.Value())
		if err != nil {
			return cidlink.Link{}, nil, NewErrSignBlock(err)
		}
	}

	link, err := putBlock(ctx, txn.Blockstore(), dagBlock)
	if err != nil {
		return cidlink.Link{}, nil, NewErrStoreBlock(err)
	}

	err = crdtData.Merge(ctx, txn.Datastore(), block.Delta.GetDelta())
	if err != nil {
		return cidlink.Link{}, nil, NewErrProcessBlock(NewErrMergingDelta(link.Cid, err))
	}

	err = UpdateHeads(ctx, crdtData.HeadstorePrefix(), block, link)
	if err != nil {
		return cidlink.Link{}, nil, NewErrProcessBlock(err)
	}

	b, err := dagBlock.Marshal()
	if err != nil {
		return cidlink.Link{}, nil, NewErrMarshalBlock(err)
	}

	return link, b, err
}

func determineBlockEncryption(
	ctx context.Context,
	docKey []byte,
	fieldName immutable.Option[string],
	heads []cid.Cid,
) (*Encryption, cidlink.Link, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	// if new encryption was requested by the user
	if encryption.ShouldEncryptDocField(ctx, fieldName) {
		encryptor := encryption.GetEncryptorFromContext(ctx)
		if encryptor != nil {
			encKey, err := encryptor.GetOrGenerateEncryptionKey(string(docKey), fieldName)
			if err != nil {
				return nil, cidlink.Link{}, NewErrGetEncryptionKey(err)
			}
			encBlock := newEncryptionBlock(encKey)
			if encBlock == nil {
				return nil, cidlink.Link{}, nil
			}

			link, err := putBlock(ctx, txn.Encstore(), encBlock)
			if err != nil {
				return nil, cidlink.Link{}, NewErrStoreEncryptionBlock(err)
			}
			return encBlock, link, nil
		}
	}

	// otherwise we use the same encryption as the previous block
	for _, headCid := range heads {
		prevBlockBytes, err := blockstore.NewIPLDStore(txn.Blockstore()).Get(ctx, headCid.KeyString())
		if err != nil {
			return nil, cidlink.Link{}, NewErrCouldNotFindBlock(headCid, err)
		}
		prevBlock, err := GetFromBytes(prevBlockBytes)
		if err != nil {
			return nil, cidlink.Link{}, NewErrDecodePreviousBlock(err)
		}
		if prevBlock.Encryption != nil {
			prevBlockEncBytes, err := blockstore.NewIPLDStore(txn.Encstore()).Get(ctx, prevBlock.Encryption.Cid.KeyString())
			if err != nil {
				return nil, cidlink.Link{}, NewErrCouldNotFindBlock(headCid, err)
			}
			prevEncBlock, err := GetEncryptionBlockFromBytes(prevBlockEncBytes)
			if err != nil {
				return nil, cidlink.Link{}, NewErrDecodeEncryptionBlock(err)
			}
			return &Encryption{
				Key: prevEncBlock.Key,
			}, *prevBlock.Encryption, nil
		}
	}

	return nil, cidlink.Link{}, nil
}

func newEncryptionBlock(encKey []byte) *Encryption {
	if len(encKey) == 0 {
		return nil
	}
	return &Encryption{Key: encKey}
}

func encryptBlock(
	ctx context.Context,
	block *Block,
	encBlock *Encryption,
) (*Block, error) {
	if block.Delta.IsComposite() || block.Delta.IsCollection() {
		return block, nil
	}

	clonedCRDT := block.Delta.Clone()
	_, encryptor := encryption.EnsureContextWithEncryptor(ctx)
	bytes, err := encryptor.Encrypt(clonedCRDT.GetData(), encBlock.Key)
	if err != nil {
		return nil, NewErrEncryptBlockData(err)
	}
	clonedCRDT.SetData(bytes)
	return &Block{Delta: clonedCRDT, Heads: block.Heads, Links: block.Links}, nil
}

func UpdateHeads(
	ctx context.Context,
	prefix keys.HeadstoreKey,
	block *Block,
	blockLink cidlink.Link,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	headset := NewHeadSet(txn.Headstore(), prefix)

	priority := block.Delta.GetPriority()

	if len(block.Heads) == 0 { // reached the bottom, at a leaf
		err := headset.Write(ctx, blockLink.Cid, priority)
		if err != nil {
			return NewErrAddingHead(blockLink.Cid, err)
		}
	}

	// Marking the block as merged removes the to-merge index. It signals that nothing
	// else needs to be done for that block.
	err := txn.Blockstore().MarkAsMerged(ctx, blockLink.Cid)
	if err != nil {
		return NewErrMarkingAsMerged(blockLink.Cid, err)
	}

	for _, l := range block.AllLinks() {
		linkCid := l.Cid
		isHead, err := headset.IsHead(ctx, linkCid)
		if err != nil {
			return NewErrCheckingHead(linkCid, err)
		}

		// Marking the block as merged removes the to-merge index. It signals that nothing
		// else needs to be done for that block.
		err = txn.Blockstore().MarkAsMerged(ctx, linkCid)
		if err != nil {
			return NewErrMarkingAsMerged(blockLink.Cid, err)
		}

		if isHead {
			// reached one of the current heads, replace it with the tip
			// of current branch
			err = headset.Replace(ctx, linkCid, blockLink.Cid, priority)
			if err != nil {
				return NewErrReplacingHead(linkCid, blockLink.Cid, err)
			}

			continue
		}

		known, err := txn.Blockstore().Has(ctx, linkCid)
		if err != nil {
			return NewErrCouldNotFindBlock(linkCid, err)
		}
		if known {
			// we reached a non-head node in the known tree.
			// This means our root block is a new head
			err := headset.Write(ctx, blockLink.Cid, priority)
			if err != nil {
				log.ErrorContextE(
					ctx,
					"Failure adding head (when root is a new head)",
					err,
					corelog.Any("Root", blockLink.Cid),
				)
				// OR should this also return like below comment??
				// return nil, errors.Wrap("error adding head (when root is new head): %s ", root, err)
			}
			continue
		}
	}

	return nil
}
