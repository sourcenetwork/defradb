// Copyright 2026 Democratized Data Foundation
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
	"errors"
	"io"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// collectedBlock holds a CID and its raw bytes to avoid double-reading from the store.
type collectedBlock struct {
	cid      cid.Cid
	rawBytes []byte
}

// rootBlockWithBytes holds a parsed block along with its raw bytes to avoid re-reading from the store.
type rootBlockWithBytes struct {
	block    *coreblock.Block
	rawBytes []byte
}

// generateCARForBlocksWithBytes creates a CAR file containing multiple root blocks and all their linked blocks.
func (p *P2P) generateCARForBlocksWithBytes(ctx context.Context, rootBlocks []rootBlockWithBytes) ([]byte, error) {
	if len(rootBlocks) == 0 {
		return nil, nil
	}

	var txnCtx context.Context
	if _, ok := corekv.TryGetCtxTxn(ctx); ok {
		txnCtx = ctx
	} else {
		txn := p.db.Rootstore().NewTxn(true)
		defer txn.Discard()
		txnCtx = corekv.SetCtxTxn(ctx, txn)
	}

	bstore := p.db.Multistore().Blockstore()
	encStore := datastore.EncstoreFrom(p.db.Rootstore())
	collectedBlocks := make(map[string]*collectedBlock)
	rootCIDs := make([]cid.Cid, 0, len(rootBlocks))

	for _, rb := range rootBlocks {
		var rootCID cid.Cid
		var err error
		if rb.rawBytes != nil {
			var rootLink cidlink.Link
			rootLink, err = rb.block.GenerateLink()
			rootCID = rootLink.Cid
		} else {
			var rootLink cidlink.Link
			rootLink, err = rb.block.GenerateLink()
			rootCID = rootLink.Cid
		}
		if err != nil {
			return nil, err
		}
		rootCIDs = append(rootCIDs, rootCID)
		if rb.rawBytes != nil {
			cidStr := rootCID.String()
			collectedBlocks[cidStr] = &collectedBlock{
				cid:      rootCID,
				rawBytes: rb.rawBytes,
			}
			if rb.block.Signature != nil {
				sigCID := rb.block.Signature.Cid
				if _, seen := collectedBlocks[sigCID.String()]; !seen {
					sigBlk, sigErr := bstore.Get(txnCtx, sigCID)
					if sigErr == nil {
						collectedBlocks[sigCID.String()] = &collectedBlock{
							cid:      sigCID,
							rawBytes: sigBlk.RawData(),
						}
					}
				}
			}
			if rb.block.Encryption != nil {
				encCID := rb.block.Encryption.Cid
				if _, seen := collectedBlocks[encCID.String()]; !seen {
					encBlk, encErr := encStore.Get(txnCtx, encCID)
					if encErr == nil {
						collectedBlocks[encCID.String()] = &collectedBlock{
							cid:      encCID,
							rawBytes: encBlk.RawData(),
						}
					}
				}
			}
			for _, dagLink := range rb.block.Links {
				if err := p.collectDAGBlocksWithBytes(txnCtx, bstore, encStore, dagLink.Cid, collectedBlocks); err != nil {
					return nil, err
				}
			}
		} else {
			if err := p.collectDAGBlocksWithBytes(txnCtx, bstore, encStore, rootCID, collectedBlocks); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	carWriter, err := storage.NewWritable(&buf, rootCIDs, car.WriteAsCarV1(true))
	if err != nil {
		return nil, err
	}

	for _, cb := range collectedBlocks {
		if err := carWriter.Put(txnCtx, cb.cid.KeyString(), cb.rawBytes); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// collectDAGBlocksWithBytes recursively collects all blocks in the DAG with their raw bytes.
func (p *P2P) collectDAGBlocksWithBytes(
	ctx context.Context,
	bstore datastore.Blockstore,
	encStore datastore.Blockstore,
	blockCID cid.Cid,
	collected map[string]*collectedBlock,
) error {
	cidStr := blockCID.String()
	if _, seen := collected[cidStr]; seen {
		return nil
	}

	var rawBytes []byte
	blk, err := bstore.Get(ctx, blockCID)
	if err != nil {
		encBlk, encErr := encStore.Get(ctx, blockCID)
		if encErr != nil {
			return err
		}
		rawBytes = encBlk.RawData()
	} else {
		rawBytes = blk.RawData()
	}

	collected[cidStr] = &collectedBlock{
		cid:      blockCID,
		rawBytes: rawBytes,
	}

	block, err := coreblock.GetFromBytes(rawBytes)
	if err != nil {
		return nil
	}

	if block.Signature != nil {
		sigCID := block.Signature.Cid
		if _, seen := collected[sigCID.String()]; !seen {
			sigBlk, sigErr := bstore.Get(ctx, sigCID)
			if sigErr == nil {
				collected[sigCID.String()] = &collectedBlock{
					cid:      sigCID,
					rawBytes: sigBlk.RawData(),
				}
			}
		}
	}

	if block.Encryption != nil {
		encCID := block.Encryption.Cid
		if _, seen := collected[encCID.String()]; !seen {
			encBlk, encErr := encStore.Get(ctx, encCID)
			if encErr == nil {
				collected[encCID.String()] = &collectedBlock{
					cid:      encCID,
					rawBytes: encBlk.RawData(),
				}
			}
		}
	}

	for _, dagLink := range block.Links {
		if err := p.collectDAGBlocksWithBytes(ctx, bstore, encStore, dagLink.Cid, collected); err != nil {
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
