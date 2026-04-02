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

// BlockCIDCollector collects CIDs during batch operations for later signing.
type BlockCIDCollector = coreblock.BlockCIDCollector

// BlockSignature contains a signature over a block's document CIDs.
type BlockSignature = coreblock.BlockSignature

// NewBlockCIDCollector creates a new CID collector for block signing.
func NewBlockCIDCollector() *BlockCIDCollector {
	return coreblock.NewBlockCIDCollector()
}

// ContextWithBlockSigning returns a context with block signing mode enabled.
func ContextWithBlockSigning(ctx context.Context, collector *BlockCIDCollector) context.Context {
	return coreblock.ContextWithBlockSigning(ctx, collector)
}

// IsBlockSigningEnabled returns true if block signing mode is enabled.
func IsBlockSigningEnabled(ctx context.Context) bool {
	return coreblock.IsBlockSigningEnabled(ctx)
}

// SignBlock creates a block signature for the collected CIDs.
// This function should be called after all documents in a batch have been created.
func SignBlock(ctx context.Context, collector *BlockCIDCollector) (*BlockSignature, error) {
	return coreblock.SignBlock(ctx, collector)
}

// ComputeMerkleRoot computes a merkle root hash from a list of CIDs.
// The CIDs are sorted for deterministic ordering before hashing.
func ComputeMerkleRoot(cids []cid.Cid) []byte {
	return coreblock.ComputeMerkleRoot(cids)
}

// VerifyBlockSignatureCIDs verifies a block signature against a list of CIDs.
func VerifyBlockSignatureCIDs(blockSig *BlockSignature, cids []cid.Cid) (bool, error) {
	return coreblock.VerifyBlockSignatureCIDs(blockSig, cids)
}

// CollectDocumentCIDs retrieves all head block CIDs for the given documents from the headstore.
func CollectDocumentCIDs(ctx context.Context, docIDs []string) ([]cid.Cid, error) {
	return coreblock.CollectDocumentCIDs(ctx, docIDs)
}

// MerkleProof contains sibling hashes and path flags for Merkle inclusion proofs.
type MerkleProof = coreblock.MerkleProof

// SortedCIDStrings returns CID string representations in canonical sorted order.
func SortedCIDStrings(cids []cid.Cid) []string {
	return coreblock.SortedCIDStrings(cids)
}

// GenerateMerkleProof builds an inclusion proof for the CID at targetIndex
// within a pre-sorted CID list.
func GenerateMerkleProof(sortedCids []cid.Cid, targetIndex int) *MerkleProof {
	return coreblock.GenerateMerkleProof(sortedCids, targetIndex)
}

// VerifyMerkleProof checks that a leaf CID's proof produces the expected Merkle root.
func VerifyMerkleProof(leafCID cid.Cid, proof *MerkleProof, expectedRoot []byte) bool {
	return coreblock.VerifyMerkleProof(leafCID, proof, expectedRoot)
}
