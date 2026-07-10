// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package node re-exports batch-signing helpers from the internal coreblock package
// so that callers of the node package do not need to import internal paths.
package node

import (
	"context"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// BatchCIDCollector is re-exported from coreblock for use by node consumers.
type BatchCIDCollector = coreblock.BatchCIDCollector

// BatchSignature is re-exported from coreblock for use by node consumers.
type BatchSignature = coreblock.BatchSignature

// NewBatchCIDCollector returns a new, empty BatchCIDCollector.
func NewBatchCIDCollector() *BatchCIDCollector {
	return coreblock.NewBatchCIDCollector()
}

// ContextWithBatchSigning embeds a BatchCIDCollector into the context.
func ContextWithBatchSigning(ctx context.Context, collector *BatchCIDCollector) context.Context {
	return coreblock.ContextWithBatchSigning(ctx, collector)
}

// IsBatchSigningEnabled returns true if a BatchCIDCollector is present in ctx.
func IsBatchSigningEnabled(ctx context.Context) bool {
	return coreblock.IsBatchSigningEnabled(ctx)
}

// SignBatch computes the Merkle root of the collected CIDs and signs it.
func SignBatch(ctx context.Context, collector *BatchCIDCollector) (*BatchSignature, error) {
	return coreblock.SignBatch(ctx, collector)
}

// ComputeMerkleRoot builds a deterministic binary Merkle tree and returns the root hash.
func ComputeMerkleRoot(cids []cid.Cid) []byte {
	return coreblock.ComputeMerkleRoot(cids)
}

// VerifyBatchSignature verifies a batch signature against a set of CIDs.
func VerifyBatchSignature(batchSig *BatchSignature, cids []cid.Cid) (bool, error) {
	return coreblock.VerifyBatchSignature(batchSig, cids)
}

// CollectDocumentCIDs returns the head block CIDs for the given document IDs by
// scanning the headstore. systemstore is used to resolve each DocID to its short ID.
func CollectDocumentCIDs(
	ctx context.Context,
	systemstore corekv.Reader,
	headstore corekv.ReaderWriter,
	collectionShortID uint32,
	docIDs []string,
) ([]cid.Cid, error) {
	return coreblock.CollectDocumentCIDs(ctx, systemstore, headstore, collectionShortID, docIDs)
}
