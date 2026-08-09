// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"bytes"
	"context"
	"io"

	"github.com/ipfs/go-cid"
	car "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/immutable"
)

// generateCAR creates a CAR file containing the root block and all its linked blocks.
func (p *P2P) generateCAR(ctx context.Context, rootBlock *coreblock.Block) ([]byte, error) {
	txn := p.db.Rootstore().NewTxn(true)
	defer txn.Discard()
	txnCtx := corekv.SetCtxTxn(ctx, txn)

	rootLink, err := rootBlock.GenerateLink()
	if err != nil {
		return nil, err
	}

	bstore := datastore.BlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	linkSystem := makeLinkSystem(blockstore.NewIPLDStore(bstore))

	blockCIDs := make(map[string]struct{})
	if err := p.collectDAGBlocks(txnCtx, &linkSystem, rootLink.Cid, blockCIDs); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	carWriter, err := storage.NewWritable(&buf, []cid.Cid{rootLink.Cid}, car.WriteAsCarV1(true))
	if err != nil {
		return nil, err
	}

	encStore := datastore.EncstoreFrom(p.db.Rootstore())

	for cidStr := range blockCIDs {
		c, err := cid.Decode(cidStr)
		if err != nil {
			return nil, err
		}

		var blockBytes []byte
		block, err := bstore.Get(txnCtx, c)
		if err != nil {
			encBlock, encErr := encStore.Get(txnCtx, c)
			if encErr != nil {
				return nil, err
			}
			blockBytes = encBlock.RawData()
		} else {
			blockBytes = block.RawData()
		}

		if err := carWriter.Put(txnCtx, c.KeyString(), blockBytes); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// collectDAGBlocks recursively collects all block CIDs in the DAG by following Links.
func (p *P2P) collectDAGBlocks(
	ctx context.Context,
	linkSystem *linking.LinkSystem,
	blockCID cid.Cid,
	visited map[string]struct{},
) error {
	cidStr := blockCID.String()
	if _, seen := visited[cidStr]; seen {
		return nil
	}
	visited[cidStr] = struct{}{}

	node, err := linkSystem.Load(
		linking.LinkContext{Ctx: ctx},
		cidlink.Link{Cid: blockCID},
		coreblock.BlockSchemaPrototype,
	)
	if err != nil {
		// Block may have been pruned locally; it was already pushed to replicators before
		// pruning so they already hold it. Skip rather than aborting the CAR.
		return nil
	}

	block, err := coreblock.GetFromNode(node)
	if err != nil {
		return err
	}

	// Include Signature block CID if present.
	if block.Signature != nil {
		visited[block.Signature.Cid.String()] = struct{}{}
	}

	// Include Encryption block CID if present.
	if block.Encryption != nil {
		visited[block.Encryption.Cid.String()] = struct{}{}
	}

	for _, dagLink := range block.Links {
		if err := p.collectDAGBlocks(ctx, linkSystem, dagLink.Link.Cid, visited); err != nil {
			return err
		}
	}

	return nil
}

// peekCARRootBlock reads only the root block from a CAR byte slice and decodes it
// without writing anything to the blockstore. Used to run pre-storage checks
// (CID mismatch, access, replication filter) before committing any blocks.
func peekCARRootBlock(carData []byte) (*coreblock.Block, error) {
	reader, err := car.NewBlockReader(bytes.NewReader(carData))
	if err != nil {
		return nil, err
	}
	if len(reader.Roots) == 0 {
		return nil, ErrEmptyCARRoots
	}
	rootCID := reader.Roots[0]
	for {
		carBlock, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if carBlock.Cid().Equals(rootCID) {
			return coreblock.GetFromBytes(carBlock.RawData())
		}
	}
	return nil, ErrCARRootBlockNotFound
}

// importCAR extracts all blocks from a CAR byte slice and stores them in the blockstore.
// Returns the root block for further processing.
func (p *P2P) importCAR(ctx context.Context, carData []byte) (*coreblock.Block, error) {
	reader, err := car.NewBlockReader(bytes.NewReader(carData))
	if err != nil {
		return nil, err
	}

	roots := reader.Roots
	if len(roots) == 0 {
		return nil, ErrEmptyCARRoots
	}

	// A CAR carries whole document DAGs, and content-addressed field blocks recur across
	// documents sharing a field value, so a CAR routinely contains blocks already stored.
	// The guarded store skips those instead of rewriting the block and re-stamping a
	// to-merge marker on one that already merged.
	bstore := datastore.P2PBlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	encStore := datastore.EncstoreFrom(p.db.Rootstore())
	var rootBlock *coreblock.Block

	for {
		carBlock, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		decodedBlock, err := coreblock.GetFromBytes(carBlock.RawData())
		if err != nil {
			_, encErr := coreblock.GetEncryptionBlockFromBytes(carBlock.RawData())
			if encErr == nil {
				if putErr := encStore.Put(ctx, carBlock); putErr != nil {
					return nil, putErr
				}
			} else {
				if putErr := bstore.Put(ctx, carBlock); putErr != nil {
					return nil, putErr
				}
			}
			continue
		}

		if putErr := bstore.Put(ctx, carBlock); putErr != nil {
			return nil, putErr
		}

		if carBlock.Cid().Equals(roots[0]) {
			rootBlock = decodedBlock
		}
	}

	if rootBlock == nil {
		return nil, ErrCARRootBlockNotFound
	}

	return rootBlock, nil
}
