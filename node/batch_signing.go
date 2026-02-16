// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"

	cid "github.com/ipfs/go-cid"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// BatchCIDCollector collects CIDs during batch operations for later signing.
type BatchCIDCollector = coreblock.BatchCIDCollector

// BatchSignature contains a signature over a batch of block CIDs.
type BatchSignature = coreblock.BatchSignature

// NewBatchCIDCollector creates a new CID collector for batch signing.
func NewBatchCIDCollector() *BatchCIDCollector {
	return coreblock.NewBatchCIDCollector()
}

// ContextWithBatchSigning returns a context with batch signing mode enabled.
func ContextWithBatchSigning(ctx context.Context, collector *BatchCIDCollector) context.Context {
	return coreblock.ContextWithBatchSigning(ctx, collector)
}

// IsBatchSigningEnabled returns true if batch signing mode is enabled.
func IsBatchSigningEnabled(ctx context.Context) bool {
	return coreblock.IsBatchSigningEnabled(ctx)
}

// SignBatch creates a batch signature for the collected CIDs.
// This function should be called after all documents in a batch have been created.
func SignBatch(ctx context.Context, collector *BatchCIDCollector) (*BatchSignature, error) {
	return coreblock.SignBatch(ctx, collector)
}

// ComputeMerkleRoot computes a merkle root hash from a list of CIDs.
// The CIDs are sorted for deterministic ordering before hashing.
func ComputeMerkleRoot(cids []cid.Cid) []byte {
	return coreblock.ComputeMerkleRoot(cids)
}

// VerifyBatchSignature verifies a batch signature against a list of CIDs.
func VerifyBatchSignature(batchSig *BatchSignature, cids []cid.Cid) (bool, error) {
	return coreblock.VerifyBatchSignature(batchSig, cids)
}
