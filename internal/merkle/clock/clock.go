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

	cid "github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

var (
	log = corelog.NewLogger("merkleclock")
)

// MerkleClock is a MerkleCRDT clock that can be used to read/write events (deltas) to the clock.
type MerkleClock struct {
	headstore  datastore.DSReaderWriter
	blockstore datastore.Blockstore
	headset    *heads
	crdt       core.ReplicatedData
}

// NewMerkleClock returns a new MerkleClock.
func NewMerkleClock(
	headstore datastore.DSReaderWriter,
	blockstore datastore.Blockstore,
	namespace core.HeadStoreKey,
	crdt core.ReplicatedData,
) *MerkleClock {
	return &MerkleClock{
		headstore:  headstore,
		blockstore: blockstore,
		headset:    NewHeadSet(headstore, namespace),
		crdt:       crdt,
	}
}

func (mc *MerkleClock) putBlock(
	ctx context.Context,
	block *coreblock.Block,
) (cidlink.Link, error) {
	nd := block.GenerateNode()
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(mc.blockstore.AsIPLDStorage())
	link, err := lsys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), nd)
	if err != nil {
		return cidlink.Link{}, NewErrWritingBlock(err)
	}

	return link.(cidlink.Link), nil
}

// AddDelta adds a new delta to the existing DAG for this MerkleClock: checks the current heads,
// sets the delta priority in the Merkle DAG, and adds it to the blockstore then runs ProcessBlock.
func (mc *MerkleClock) AddDelta(
	ctx context.Context,
	delta core.Delta,
	links ...coreblock.DAGLink,
) (cidlink.Link, []byte, error) {
	heads, height, err := mc.headset.List(ctx)
	if err != nil {
		return cidlink.Link{}, nil, NewErrGettingHeads(err)
	}
	height = height + 1

	delta.SetPriority(height)
	block := coreblock.New(delta, links, heads...)

	fieldName := immutable.None[string]()
	if block.Delta.GetFieldName() != "" {
		fieldName = immutable.Some(block.Delta.GetFieldName())
	}
	blockEnc, err := mc.determineBlockEncryptionData(ctx, fieldName, height, heads)
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	dagBlock := block
	if blockEnc != nil && blockEnc.Type != coreblock.NotEncrypted {
		if !block.Delta.IsComposite() {
			dagBlock, err = encryptBlock(ctx, block, blockEnc)
			if err != nil {
				return cidlink.Link{}, nil, err
			}
		} else {
			dagBlock.Encryption = blockEnc
		}
	}

	link, err := mc.putBlock(ctx, dagBlock)
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	// merge the delta and update the state
	err = mc.ProcessBlock(ctx, block, link, false, false)
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	b, err := dagBlock.Marshal()
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	return link, b, err
}

func (mc *MerkleClock) determineBlockEncryptionData(
	ctx context.Context,
	fieldName immutable.Option[string],
	height uint64,
	heads []cid.Cid,
) (*coreblock.Encryption, error) {
	// if new encryption was requested by the user
	if encryption.ShouldEncryptDocField(ctx, fieldName) {
		blockEnc := &coreblock.Encryption{From: height}
		if encryption.ShouldEncryptIndividualField(ctx, fieldName) {
			blockEnc.Type = coreblock.FieldEncrypted
		} else {
			blockEnc.Type = coreblock.DocumentEncrypted
		}
		return blockEnc, nil
	}

	// otherwise we use the same encryption as the previous block
	for _, headCid := range heads {
		bytes, err := mc.blockstore.AsIPLDStorage().Get(ctx, headCid.KeyString())
		if err != nil {
			return nil, NewErrCouldNotFindBlock(headCid, err)
		}
		prevBlock, err := coreblock.GetFromBytes(bytes)
		if err != nil {
			return nil, err
		}
		if prevBlock.Encryption != nil {
			blockEnc := &coreblock.Encryption{Type: (*prevBlock.Encryption).Type}
			blockEnc.From = prevBlock.Encryption.From
			return blockEnc, nil
		}
	}

	return nil, nil
}

func encryptBlock(
	ctx context.Context,
	block *coreblock.Block,
	blockEnc *coreblock.Encryption,
) (*coreblock.Block, error) {
	clonedCRDT := block.Delta.Clone()
	fieldName := immutable.None[string]()
	if blockEnc.Type == coreblock.FieldEncrypted {
		fieldName = immutable.Some(clonedCRDT.GetFieldName())
	}
	bytes, err := encryption.EncryptDoc(ctx, string(clonedCRDT.GetDocID()),
		fieldName, blockEnc.From, clonedCRDT.GetData())
	if err != nil {
		return nil, err
	}
	clonedCRDT.SetData(bytes)
	return &coreblock.Block{Delta: clonedCRDT, Links: block.Links, Encryption: blockEnc}, nil
}

// ProcessBlock merges the delta CRDT and updates the state accordingly.
// If skipMerge is true, it will skip merging and update only the heads.
// If skipHeads is true, it will skip updating the heads.
func (mc *MerkleClock) ProcessBlock(
	ctx context.Context,
	block *coreblock.Block,
	blockLink cidlink.Link,
	skipMerge bool,
	skipHeads bool,
) error {
	if !skipMerge {
		err := mc.crdt.Merge(ctx, block.Delta.GetDelta())
		if err != nil {
			return NewErrMergingDelta(blockLink.Cid, err)
		}
	}

	if !skipHeads {
		return mc.updateHeads(ctx, block, blockLink)
	}
	return nil
}

func (mc *MerkleClock) updateHeads(
	ctx context.Context,
	block *coreblock.Block,
	blockLink cidlink.Link,
) error {
	priority := block.Delta.GetPriority()

	// check if we have any HEAD links
	hasHeads := false
	for _, l := range block.Links {
		if l.Name == core.HEAD {
			hasHeads = true
			break
		}
	}
	if !hasHeads { // reached the bottom, at a leaf
		err := mc.headset.Write(ctx, blockLink.Cid, priority)
		if err != nil {
			return NewErrAddingHead(blockLink.Cid, err)
		}
	}

	for _, l := range block.Links {
		linkCid := l.Cid
		isHead, err := mc.headset.IsHead(ctx, linkCid)
		if err != nil {
			return NewErrCheckingHead(linkCid, err)
		}

		if isHead {
			// reached one of the current heads, replace it with the tip
			// of current branch
			err = mc.headset.Replace(ctx, linkCid, blockLink.Cid, priority)
			if err != nil {
				return NewErrReplacingHead(linkCid, blockLink.Cid, err)
			}

			continue
		}

		known, err := mc.blockstore.Has(ctx, linkCid)
		if err != nil {
			return NewErrCouldNotFindBlock(linkCid, err)
		}
		if known {
			// we reached a non-head node in the known tree.
			// This means our root block is a new head
			err := mc.headset.Write(ctx, blockLink.Cid, priority)
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

// Heads returns the current heads of the MerkleClock.
func (mc *MerkleClock) Heads() *heads {
	return mc.headset
}
