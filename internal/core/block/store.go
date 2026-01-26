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
)

// blockMarshaler is implemented by blocks that can be marshaled to bytes.
type blockMarshaler interface {
	GenerateNode() ipld.Node
	Marshal() ([]byte, error)
}

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
	cidLink := link.(cidlink.Link) //nolint:forcetypeassert
	if marshaler, ok := block.(blockMarshaler); ok {
		if bytes, err := marshaler.Marshal(); err == nil {
			GetGlobalBlockCache().Put(cidLink.Cid, bytes)
		}
	}
	return cidLink, nil
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
	encBlock, encLink, err := determineBlockEncryption(ctx, string(block.Delta.GetDocID()), fieldName, heads)
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	dagBlock := block
	if encBlock != nil {
		dagBlock, err = encryptBlock(ctx, block, encBlock)
		if err != nil {
			return cidlink.Link{}, nil, err
		}
		dagBlock.Encryption = &encLink
	}

	// In batch mode store block without signature and collect CID for batch signing
	collector := BatchSigningCollectorFromContext(ctx)
	if collector != nil {
		link, err := putBlock(ctx, txn.Blockstore(), dagBlock)
		if err != nil {
			return cidlink.Link{}, nil, err
		}
		if !dagBlock.Delta.IsField() || dagBlock.Delta.GetPriority() == 1 {
			collector.Add(link.Cid)
		}
		err = ProcessBlock(ctx, crdtData, block, link)
		if err != nil {
			return cidlink.Link{}, nil, err
		}
		b, err := dagBlock.Marshal()
		if err != nil {
			return cidlink.Link{}, nil, err
		}
		return link, b, err
	}

	// Normal signing mode
	if EnabledSigningFromContext(ctx) {
		err = signBlock(ctx, txn.Blockstore(), dagBlock)
		if err != nil {
			return cidlink.Link{}, nil, err
		}
	}

	link, err := putBlock(ctx, txn.Blockstore(), dagBlock)
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	// merge the delta and update the state
	err = ProcessBlock(ctx, crdtData, block, link)
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	b, err := dagBlock.Marshal()
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	return link, b, err
}

func determineBlockEncryption(
	ctx context.Context,
	docID string,
	fieldName immutable.Option[string],
	heads []cid.Cid,
) (*Encryption, cidlink.Link, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	// if new encryption was requested by the user
	if encryption.ShouldEncryptDocField(ctx, fieldName) {
		encBlock := &Encryption{DocID: []byte(docID)}
		if encryption.ShouldEncryptIndividualField(ctx, fieldName) {
			f := fieldName.Value()
			encBlock.FieldName = &f
		}
		encryptor := encryption.GetEncryptorFromContext(ctx)
		if encryptor != nil {
			encKey, err := encryptor.GetOrGenerateEncryptionKey(docID, fieldName)
			if err != nil {
				return nil, cidlink.Link{}, err
			}
			if len(encKey) > 0 {
				encBlock.Key = encKey
			}

			link, err := putBlock(ctx, txn.Encstore(), encBlock)
			if err != nil {
				return nil, cidlink.Link{}, err
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
			return nil, cidlink.Link{}, err
		}
		if prevBlock.Encryption != nil {
			prevBlockEncBytes, err := blockstore.NewIPLDStore(txn.Encstore()).Get(ctx, prevBlock.Encryption.Cid.KeyString())
			if err != nil {
				return nil, cidlink.Link{}, NewErrCouldNotFindBlock(headCid, err)
			}
			prevEncBlock, err := GetEncryptionBlockFromBytes(prevBlockEncBytes)
			if err != nil {
				return nil, cidlink.Link{}, err
			}
			return &Encryption{
				DocID:     prevEncBlock.DocID,
				FieldName: prevEncBlock.FieldName,
				Key:       prevEncBlock.Key,
			}, *prevBlock.Encryption, nil
		}
	}

	return nil, cidlink.Link{}, nil
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
		return nil, err
	}
	clonedCRDT.SetData(bytes)
	return &Block{Delta: clonedCRDT, Heads: block.Heads, Links: block.Links}, nil
}

// ProcessBlock merges the delta CRDT and updates the state accordingly.
func ProcessBlock(
	ctx context.Context,
	crdtData crdt.ReplicatedData,
	block *Block,
	blockLink cidlink.Link,
) error {
	err := crdtData.Merge(ctx, block.Delta.GetDelta())
	if err != nil {
		return NewErrMergingDelta(blockLink.Cid, err)
	}

	return updateHeads(ctx, crdtData, block, blockLink)
}

func updateHeads(
	ctx context.Context,
	crdtData crdt.ReplicatedData,
	block *Block,
	blockLink cidlink.Link,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	headset := NewHeadSet(txn.Headstore(), crdtData.HeadstorePrefix())

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

	// For collection blocks and composites, we only process Heads, Links are processed in their own transactions.
	linksToProcess := block.AllLinks()
	if block.Delta.IsCollection() || block.Delta.IsComposite() {
		linksToProcess = block.Heads
	}

	for _, l := range linksToProcess {
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
				return NewErrAddingHead(blockLink.Cid, err)
			}
			continue
		}
	}

	return nil
}
