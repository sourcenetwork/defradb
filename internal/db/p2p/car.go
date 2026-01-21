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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/immutable"
)

// generateCAR creates a CAR file containing the root block and all its linked blocks.
func (p *P2P) generateCAR(ctx context.Context, rootBlock *coreblock.Block) ([]byte, error) {
	return p.generateCARForBlocks(ctx, []*coreblock.Block{rootBlock})
}

// generateCARForBlocks creates a CAR file containing multiple root blocks and all their linked blocks.
func (p *P2P) generateCARForBlocks(ctx context.Context, rootBlocks []*coreblock.Block) ([]byte, error) {
	if len(rootBlocks) == 0 {
		return nil, nil
	}

	txn := p.db.Rootstore().NewTxn(true)
	defer txn.Discard()
	txnCtx := corekv.SetCtxTxn(ctx, txn)

	bstore := p.db.Multistore().Blockstore()
	linkSystem := makeLinkSystem(blockstore.NewIPLDStore(bstore))

	blockCIDs := make(map[string]struct{})
	rootCIDs := make([]cid.Cid, 0, len(rootBlocks))

	for _, rootBlock := range rootBlocks {
		rootLink, err := rootBlock.GenerateLink()
		if err != nil {
			return nil, err
		}
		rootCIDs = append(rootCIDs, rootLink.Cid)

		if err := p.collectDAGBlocks(txnCtx, &linkSystem, rootLink.Cid, blockCIDs); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	carWriter, err := storage.NewWritable(&buf, rootCIDs, car.WriteAsCarV1(true))
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

	node, err := linkSystem.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCID}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return err
	}

	block, err := coreblock.GetFromNode(node)
	if err != nil {
		return err
	}

	// Include Signature block if present
	if block.Signature != nil {
		visited[block.Signature.Cid.String()] = struct{}{}
	}

	// Include Encryption block if present
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

// parsedCAR holds the parsed blocks from a CAR file, ready for batched import.
type parsedCAR struct {
	rootBlock     *coreblock.Block
	regularBlocks []blocks.Block
	encBlocks     []blocks.Block
}

// parseCAR extracts all blocks from a CAR file without creating a transaction.
func parseCAR(carData []byte) (*parsedCAR, error) {
	reader, err := car.NewBlockReader(bytes.NewReader(carData))
	if err != nil {
		return nil, err
	}

	roots := reader.Roots
	if len(roots) == 0 {
		return nil, ErrEmptyCARRoots
	}

	var regularBlocks []blocks.Block
	var encBlocks []blocks.Block
	var rootBlock *coreblock.Block

	for {
		carBlock, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		decodedBlock, err := coreblock.GetFromBytes(carBlock.RawData())
		if err != nil {
			_, encErr := coreblock.GetEncryptionBlockFromBytes(carBlock.RawData())
			if encErr == nil {
				encBlocks = append(encBlocks, carBlock)
			} else {
				regularBlocks = append(regularBlocks, carBlock)
			}
			continue
		}

		regularBlocks = append(regularBlocks, carBlock)

		if carBlock.Cid().Equals(roots[0]) {
			rootBlock = decodedBlock
		}
	}

	if rootBlock == nil {
		return nil, ErrCARRootBlockNotFound
	}

	return &parsedCAR{
		rootBlock:     rootBlock,
		regularBlocks: regularBlocks,
		encBlocks:     encBlocks,
	}, nil
}

// importCARBatch imports multiple parsed CAR files in a single transaction.
func (p *P2P) importCARBatch(ctx context.Context, parsedCARs []*parsedCAR) error {
	if len(parsedCARs) == 0 {
		return nil
	}

	var allRegularBlocks []blocks.Block
	var allEncBlocks []blocks.Block

	for _, pc := range parsedCARs {
		allRegularBlocks = append(allRegularBlocks, pc.regularBlocks...)
		allEncBlocks = append(allEncBlocks, pc.encBlocks...)
	}

	txn := p.db.Rootstore().NewTxn(false)
	defer txn.Discard()

	txnCtx := corekv.SetCtxTxn(ctx, txn)
	bstore := datastore.P2PBlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	encStore := datastore.EncstoreFrom(p.db.Rootstore())

	if len(allRegularBlocks) > 0 {
		if err := bstore.PutMany(txnCtx, allRegularBlocks); err != nil {
			return err
		}
	}

	if len(allEncBlocks) > 0 {
		if err := encStore.PutMany(txnCtx, allEncBlocks); err != nil {
			return err
		}
	}

	return txn.Commit()
}

// importCAR extracts all blocks from a CAR file and stores them in the blockstore.
func (p *P2P) importCAR(ctx context.Context, carData []byte) (*coreblock.Block, error) {
	parsed, err := parseCAR(carData)
	if err != nil {
		return nil, err
	}

	err = p.importCARBatch(ctx, []*parsedCAR{parsed})
	if err != nil {
		return nil, err
	}

	return parsed.rootBlock, nil
}
