// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"bytes"
	"context"
	"crypto/sha256"
	"sort"
	"sync"

	cid "github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	iidentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type blockSigningContextKey struct{}

// BlockCIDCollector collects CIDs during block operations for later signing.
type BlockCIDCollector struct {
	mu   sync.Mutex
	cids []cid.Cid
}

// NewBlockCIDCollector creates a new CID collector for block signing.
func NewBlockCIDCollector() *BlockCIDCollector {
	return &BlockCIDCollector{
		cids: make([]cid.Cid, 0),
	}
}

// Add adds a CID to the collector.
func (c *BlockCIDCollector) Add(cid cid.Cid) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cids = append(c.cids, cid)
}

// GetCIDs returns all collected CIDs.
func (c *BlockCIDCollector) GetCIDs() []cid.Cid {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]cid.Cid, len(c.cids))
	copy(result, c.cids)
	return result
}

// ContextWithBlockSigning returns a context with block signing mode enabled.
// In block signing mode, individual block signing is skipped and CIDs are
// collected for later block signing.
func ContextWithBlockSigning(ctx context.Context, collector *BlockCIDCollector) context.Context {
	return context.WithValue(ctx, blockSigningContextKey{}, collector)
}

// BlockSigningCollectorFromContext returns the batch CID collector if block signing is enabled.
func BlockSigningCollectorFromContext(ctx context.Context) *BlockCIDCollector {
	val := ctx.Value(blockSigningContextKey{})
	if val == nil {
		return nil
	}
	return val.(*BlockCIDCollector) //nolint:forcetypeassert
}

// IsBlockSigningEnabled returns true if block signing mode is enabled.
func IsBlockSigningEnabled(ctx context.Context) bool {
	return BlockSigningCollectorFromContext(ctx) != nil
}

// sortCIDs returns a new slice of CIDs sorted in canonical byte order.
func sortCIDs(cids []cid.Cid) []cid.Cid {
	sorted := make([]cid.Cid, len(cids))
	copy(sorted, cids)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i].Bytes(), sorted[j].Bytes()) < 0
	})
	return sorted
}

// SortedCIDStrings returns CID string representations in canonical sorted order.
// This is the format stored on BlockSignature documents.
func SortedCIDStrings(cids []cid.Cid) []string {
	sorted := sortCIDs(cids)
	result := make([]string, len(sorted))
	for i, c := range sorted {
		result[i] = c.String()
	}
	return result
}

// computeLeafHashes computes SHA256 leaf hashes from pre-sorted CIDs.
func computeLeafHashes(sortedCids []cid.Cid) [][]byte {
	hashes := make([][]byte, len(sortedCids))
	for i, c := range sortedCids {
		hash := sha256.Sum256(c.Bytes())
		hashes[i] = hash[:]
	}
	return hashes
}

// buildMerkleTree computes the root from leaf hashes using pair-wise SHA256.
func buildMerkleTree(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return nil
	}
	combined := make([]byte, 64)
	for len(hashes) > 1 {
		newLen := (len(hashes) + 1) / 2
		newHashes := make([][]byte, 0, newLen)
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				copy(combined[:32], hashes[i])
				copy(combined[32:], hashes[i+1])
				hash := sha256.Sum256(combined)
				newHashes = append(newHashes, hash[:])
			} else {
				newHashes = append(newHashes, hashes[i])
			}
		}
		hashes = newHashes
	}
	return hashes[0]
}

// ComputeMerkleRoot computes a merkle root hash from a list of CIDs.
// The CIDs are sorted for deterministic ordering before hashing.
func ComputeMerkleRoot(cids []cid.Cid) []byte {
	if len(cids) == 0 {
		return nil
	}
	return buildMerkleTree(computeLeafHashes(sortCIDs(cids)))
}

// MerkleProof contains the sibling hashes and path flags needed to verify
// that a single leaf is included in a Merkle tree with a known root.
type MerkleProof struct {
	// Siblings contains the sibling hash at each tree level from leaf to root.
	Siblings [][]byte
	// Path indicates the position at each level. false = sibling is on the right
	// (proven node is left child), true = sibling is on the left (proven node is right child).
	Path []bool
}

// GenerateMerkleProof builds an inclusion proof for the CID at targetIndex
// within the sorted CID list. The caller must pass CIDs already sorted in
// canonical byte order (use sortCIDs). Returns nil if the list is empty or
// the index is out of range.
func GenerateMerkleProof(sortedCids []cid.Cid, targetIndex int) *MerkleProof {
	n := len(sortedCids)
	if n == 0 || targetIndex < 0 || targetIndex >= n {
		return nil
	}

	hashes := computeLeafHashes(sortedCids)

	var siblings [][]byte
	var path []bool
	idx := targetIndex

	combined := make([]byte, 64)
	for len(hashes) > 1 {
		newLen := (len(hashes) + 1) / 2
		newHashes := make([][]byte, 0, newLen)

		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				if i == idx || i+1 == idx {
					if idx%2 == 0 {
						sibling := make([]byte, 32)
						copy(sibling, hashes[i+1])
						siblings = append(siblings, sibling)
						path = append(path, false)
					} else {
						sibling := make([]byte, 32)
						copy(sibling, hashes[i])
						siblings = append(siblings, sibling)
						path = append(path, true)
					}
				}
				copy(combined[:32], hashes[i])
				copy(combined[32:], hashes[i+1])
				hash := sha256.Sum256(combined)
				newHashes = append(newHashes, hash[:])
			} else {
				// Odd node — carried forward without a sibling.
				// No proof entry needed at this level.
				newHashes = append(newHashes, hashes[i])
			}
		}

		idx /= 2
		hashes = newHashes
	}

	return &MerkleProof{Siblings: siblings, Path: path}
}

// VerifyMerkleProof checks that a leaf CID combined with the proof siblings
// produces the expected Merkle root. Returns true if valid.
func VerifyMerkleProof(leafCID cid.Cid, proof *MerkleProof, expectedRoot []byte) bool {
	if proof == nil || len(expectedRoot) == 0 {
		return false
	}

	current := sha256.Sum256(leafCID.Bytes())
	hash := current[:]

	combined := make([]byte, 64)
	for i, sibling := range proof.Siblings {
		if proof.Path[i] {
			// Sibling is on the left.
			copy(combined[:32], sibling)
			copy(combined[32:], hash)
		} else {
			// Sibling is on the right.
			copy(combined[:32], hash)
			copy(combined[32:], sibling)
		}
		h := sha256.Sum256(combined)
		hash = h[:]
	}

	return bytes.Equal(hash, expectedRoot)
}

// BlockSignature contains a signature over a block's document CIDs.
type BlockSignature struct {
	Header     SignatureHeader
	Value      []byte
	MerkleRoot []byte
	CIDCount   int
}

// IPLDSchemaBytes returns the IPLD schema for BlockSignature.
func (sig *BlockSignature) IPLDSchemaBytes() []byte {
	return []byte(`
		type BlockSignature struct {
			header SignatureHeader
			value Bytes
			merkleRoot Bytes
			cidCount Int
		}
	`)
}

// SignBlock creates a block signature for the collected CIDs.
// This function should be called after all documents in a block have been created.
func SignBlock(
	ctx context.Context,
	collector *BlockCIDCollector,
) (*BlockSignature, error) {
	cids := collector.GetCIDs()
	if len(cids) == 0 {
		return nil, nil
	}

	merkleRoot := ComputeMerkleRoot(cids)

	ident := iidentity.FromContext(ctx)
	if !ident.HasValue() {
		return nil, nil
	}

	fullIdent, ok := ident.Value().(identity.FullIdentity)
	if !ok {
		return nil, nil
	}

	var sigType string
	switch fullIdent.PrivateKey().Type() {
	case crypto.KeyTypeSecp256k1:
		sigType = SignatureTypeECDSA256K
	case crypto.KeyTypeEd25519:
		sigType = SignatureTypeEd25519
	default:
		return nil, NewErrUnsupportedKeyForSigning(fullIdent.PrivateKey().Type())
	}

	sigBytes, err := fullIdent.PrivateKey().Sign(merkleRoot)
	if err != nil {
		return nil, err
	}

	blockSig := &BlockSignature{
		Header: SignatureHeader{
			Type:     sigType,
			Identity: []byte(fullIdent.PublicKey().String()),
		},
		Value:      sigBytes,
		MerkleRoot: merkleRoot,
		CIDCount:   len(cids),
	}

	return blockSig, nil
}

// VerifyBlockSignatureCIDs verifies a block signature against a list of CIDs.
func VerifyBlockSignatureCIDs(blockSig *BlockSignature, cids []cid.Cid) (bool, error) {
	if blockSig == nil {
		return false, nil
	}
	computedRoot := ComputeMerkleRoot(cids)
	if len(computedRoot) != len(blockSig.MerkleRoot) {
		return false, nil
	}
	for i := range computedRoot {
		if computedRoot[i] != blockSig.MerkleRoot[i] {
			return false, nil
		}
	}
	pubKey, err := getPublicKeyFromSignature(&Signature{Header: blockSig.Header, Value: blockSig.Value})
	if err != nil {
		return false, err
	}
	return verifySignature(pubKey, blockSig.MerkleRoot, blockSig.Value) == nil, nil
}

// CollectDocumentCIDs retrieves the composite head CID for each given document from the headstore.
// Only composite CIDs (FieldID == "C") are returned — one per document.
func CollectDocumentCIDs(ctx context.Context, docIDs []string) ([]cid.Cid, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	var allCIDs []cid.Cid

	for _, docID := range docIDs {
		compositeKey := keys.HeadstoreDocKey{DocID: docID, FieldID: core.COMPOSITE_NAMESPACE}
		iter, err := txn.Headstore().Iterator(ctx, corekv.IterOptions{
			Prefix: compositeKey.Bytes(),
		})
		if err != nil {
			return nil, err
		}

		for {
			hasNext, err := iter.Next()
			if err != nil {
				_ = iter.Close()
				return nil, err
			}
			if !hasNext {
				break
			}

			headKey, err := keys.NewHeadstoreDocKey(string(iter.Key()))
			if err != nil {
				_ = iter.Close()
				return nil, err
			}

			allCIDs = append(allCIDs, headKey.GetCid())
		}

		if err := iter.Close(); err != nil {
			return nil, err
		}
	}

	return allCIDs, nil
}
