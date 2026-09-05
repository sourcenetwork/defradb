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
	car "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/immutable"
)

// generateCAR creates a CAR file containing the root block and all its linked blocks.
func (p *P2P) generateCAR(ctx context.Context, rootBlock *coreblock.Block) ([]byte, error) {
	data, err := p.buildCAR(ctx, rootBlock)
	if err != nil {
		p.statCARFailed.Add(1)
		return nil, err
	}
	p.statCARBuilt.Add(1)
	return data, nil
}

// carFailure records the reason a CAR could not be built and logs the first occurrence of
// that reason, so a persistent failure shows up as a count rather than a line per call.
func (p *P2P) carFailure(reason string, err error) {
	if p.carFailureReason.recordFirst(reason) {
		log.ErrorE("Failed to generate CAR", err, corelog.String("reason", reason))
	}
}

// carImportFailure records an abandoned import: how it failed and how many blocks it had
// already written. Those blocks sit in the store owned by no document until a later merge
// claims them.
func (p *P2P) carImportFailure(reason string, written int64, err error) {
	p.statCARImportFailed.Add(1)
	p.statCARImportOrphanBlocks.Add(written)
	if p.carImportFailureReason.recordFirst(reason) {
		log.ErrorE("Failed to import CAR", err,
			corelog.String("reason", reason),
			corelog.Int64("blocksWritten", written))
	}
}

// buildCAR serialises the root block and the blocks it links, following Links only. Heads are
// not followed, so a CAR carries one commit and none of the document's history.
//
// A link that will not load is not followed, and if its block cannot be read either the whole
// CAR is abandoned: a receiver given a CAR imports it instead of walking the DAG, so a block
// absent from the CAR is one it never fetches. missingLinks is counted only for a CAR the
// build returns.
func (p *P2P) buildCAR(ctx context.Context, rootBlock *coreblock.Block) ([]byte, error) {
	txn := p.db.Rootstore().NewTxn(true)
	defer txn.Discard()
	txnCtx := corekv.SetCtxTxn(ctx, txn)

	rootLink, err := rootBlock.GenerateLink()
	if err != nil {
		p.carFailure(reasonRootLink, err)
		return nil, err
	}

	bstore := datastore.BlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	linkSystem := makeLinkSystem(blockstore.NewIPLDStore(bstore))

	blockCIDs := make(map[string]struct{})
	var missingLinks int64
	if err := p.collectDAGBlocks(txnCtx, &linkSystem, rootLink.Cid, blockCIDs, &missingLinks); err != nil {
		p.carFailure(reasonWalk, err)
		return nil, err
	}

	var buf bytes.Buffer
	carWriter, err := storage.NewWritable(&buf, []cid.Cid{rootLink.Cid}, car.WriteAsCarV1(true))
	if err != nil {
		p.carFailure(reasonCARWriter, err)
		return nil, err
	}

	encStore := datastore.EncstoreFrom(p.db.Rootstore())

	for cidStr := range blockCIDs {
		c, err := cid.Decode(cidStr)
		if err != nil {
			p.carFailure(reasonBlockRead, err)
			return nil, err
		}

		var blockBytes []byte
		block, err := bstore.Get(txnCtx, c)
		if err != nil {
			encBlock, encErr := encStore.Get(txnCtx, c)
			if encErr != nil {
				p.carFailure(reasonBlockRead, err)
				return nil, err
			}
			blockBytes = encBlock.RawData()
		} else {
			blockBytes = block.RawData()
		}

		if err := carWriter.Put(txnCtx, c.KeyString(), blockBytes); err != nil {
			p.carFailure(reasonCARPut, err)
			return nil, err
		}
	}

	p.statCARMissing.Add(missingLinks)
	return buf.Bytes(), nil
}

// collectDAGBlocks recursively collects block CIDs by following Links.
// missing is incremented for every link whose block could not be loaded. The CID stays in the
// write set either way, so the block is still read when the CAR is written.
func (p *P2P) collectDAGBlocks(
	ctx context.Context,
	linkSystem *linking.LinkSystem,
	blockCID cid.Cid,
	visited map[string]struct{},
	missing *int64,
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
		// A block that will not load leaves its links unknown, but its CID is already in the
		// write set, so the write loop still has to read it: if it cannot, the whole CAR is
		// abandoned.
		*missing++
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
		if err := p.collectDAGBlocks(ctx, linkSystem, dagLink.Link.Cid, visited, missing); err != nil {
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
	// An arrival carrying a CAR takes this path, one without it takes syncDAG. The generateCAR
	// counters describe what this node sends, not what it receives.
	p.statCARImports.Add(1)

	reader, err := car.NewBlockReader(bytes.NewReader(carData))
	if err != nil {
		p.carImportFailure(reasonReader, 0, err)
		return nil, err
	}

	roots := reader.Roots
	if len(roots) == 0 {
		p.carImportFailure(reasonNoRoots, 0, ErrEmptyCARRoots)
		return nil, ErrEmptyCARRoots
	}

	// Content-addressed field blocks recur across documents that share a field value, so a
	// CAR routinely contains blocks already stored. The guarded store skips those instead of
	// rewriting the block and re-stamping a to-merge marker on one that already merged.
	bstore := datastore.P2PBlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	encStore := datastore.EncstoreFrom(p.db.Rootstore())
	var rootBlock *coreblock.Block

	var written int64

	for {
		carBlock, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			p.carImportFailure(reasonNext, written, err)
			return nil, err
		}

		decodedBlock, err := coreblock.GetFromBytes(carBlock.RawData())
		if err != nil {
			store := bstore
			if _, encErr := coreblock.GetEncryptionBlockFromBytes(carBlock.RawData()); encErr == nil {
				store = encStore
			}
			added, putErr := addBlock(ctx, store, carBlock)
			if putErr != nil {
				p.carImportFailure(reasonCARPut, written, putErr)
				return nil, putErr
			}
			if added {
				written++
			}
			continue
		}

		added, putErr := addBlock(ctx, bstore, carBlock)
		if putErr != nil {
			p.carImportFailure(reasonCARPut, written, putErr)
			return nil, putErr
		}
		if added {
			written++
		}

		if carBlock.Cid().Equals(roots[0]) {
			rootBlock = decodedBlock
		}
	}

	if rootBlock == nil {
		p.carImportFailure(reasonNoRootBlock, written, ErrCARRootBlockNotFound)
		return nil, ErrCARRootBlockNotFound
	}

	return rootBlock, nil
}

// addBlock stores the block unless the store already holds it, reporting whether it was added.
// The blockstores return nil for a block they already have, so Put alone cannot report that.
//
// A Has that fails is treated as not held, as Put does.
func addBlock(ctx context.Context, store datastore.Blockstore, block blocks.Block) (bool, error) {
	if held, err := store.Has(ctx, block.Cid()); err == nil && held {
		return false, nil
	}
	if err := store.Put(ctx, block); err != nil {
		return false, err
	}
	return true, nil
}
