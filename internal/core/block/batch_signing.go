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
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type batchSigningContextKey struct{}

// BatchCIDCollector collects CIDs during batch operations for later signing.
type BatchCIDCollector struct {
	mu   sync.Mutex
	cids []cid.Cid
}

// NewBatchCIDCollector creates a new CID collector for batch signing.
func NewBatchCIDCollector() *BatchCIDCollector {
	return &BatchCIDCollector{
		cids: make([]cid.Cid, 0),
	}
}

// Add adds a CID to the collector.
func (c *BatchCIDCollector) Add(cid cid.Cid) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cids = append(c.cids, cid)
}

// GetCIDs returns all collected CIDs.
func (c *BatchCIDCollector) GetCIDs() []cid.Cid {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]cid.Cid, len(c.cids))
	copy(result, c.cids)
	return result
}

// ContextWithBatchSigning returns a context with batch signing mode enabled.
// In batch signing mode, individual block signing is skipped and CIDs are
// collected for later batch signing.
func ContextWithBatchSigning(ctx context.Context, collector *BatchCIDCollector) context.Context {
	return context.WithValue(ctx, batchSigningContextKey{}, collector)
}

// BatchSigningCollectorFromContext returns the batch CID collector if batch signing is enabled.
func BatchSigningCollectorFromContext(ctx context.Context) *BatchCIDCollector {
	val := ctx.Value(batchSigningContextKey{})
	if val == nil {
		return nil
	}
	return val.(*BatchCIDCollector) //nolint:forcetypeassert
}

// IsBatchSigningEnabled returns true if batch signing mode is enabled.
func IsBatchSigningEnabled(ctx context.Context) bool {
	return BatchSigningCollectorFromContext(ctx) != nil
}

// ComputeMerkleRoot computes a merkle root hash from a list of CIDs.
// The CIDs are sorted for deterministic ordering before hashing.
func ComputeMerkleRoot(cids []cid.Cid) []byte {
	if len(cids) == 0 {
		return nil
	}

	sortedCids := make([]cid.Cid, len(cids))
	copy(sortedCids, cids)
	sort.Slice(sortedCids, func(i, j int) bool {
		return bytes.Compare(sortedCids[i].Bytes(), sortedCids[j].Bytes()) < 0
	})

	n := len(sortedCids)
	hashes := make([][]byte, n)
	for i, c := range sortedCids {
		hash := sha256.Sum256(c.Bytes())
		hashes[i] = hash[:]
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

// BatchSignature contains a signature over a batch of block CIDs.
type BatchSignature struct {
	Header     SignatureHeader
	Value      []byte
	MerkleRoot []byte
	CIDCount   int
}

// IPLDSchemaBytes returns the IPLD schema for BatchSignature.
func (sig *BatchSignature) IPLDSchemaBytes() []byte {
	return []byte(`
		type BatchSignature struct {
			header SignatureHeader
			value Bytes
			merkleRoot Bytes
			cidCount Int
		}
	`)
}

// SignBatch creates a batch signature for the collected CIDs.
// This function should be called after all documents in a batch have been created.
func SignBatch(
	ctx context.Context,
	collector *BatchCIDCollector,
) (*BatchSignature, error) {
	cids := collector.GetCIDs()
	if len(cids) == 0 {
		return nil, nil
	}

	merkleRoot := ComputeMerkleRoot(cids)

	ident := identity.FromContext(ctx)
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

	batchSig := &BatchSignature{
		Header: SignatureHeader{
			Type:     sigType,
			Identity: []byte(fullIdent.PublicKey().String()),
		},
		Value:      sigBytes,
		MerkleRoot: merkleRoot,
		CIDCount:   len(cids),
	}

	return batchSig, nil
}

// VerifyBatchSignature verifies a batch signature against a list of CIDs.
func VerifyBatchSignature(batchSig *BatchSignature, cids []cid.Cid) (bool, error) {
	if batchSig == nil {
		return false, nil
	}
	computedRoot := ComputeMerkleRoot(cids)
	if len(computedRoot) != len(batchSig.MerkleRoot) {
		return false, nil
	}
	for i := range computedRoot {
		if computedRoot[i] != batchSig.MerkleRoot[i] {
			return false, nil
		}
	}
	pubKey, err := getPublicKeyFromSignature(&Signature{Header: batchSig.Header, Value: batchSig.Value})
	if err != nil {
		return false, err
	}
	return verifySignature(pubKey, batchSig.MerkleRoot, batchSig.Value) == nil, nil
}

// CollectDocumentCIDs retrieves all head block CIDs for the given documents from the headstore.
func CollectDocumentCIDs(ctx context.Context, docIDs []string) ([]cid.Cid, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	var allCIDs []cid.Cid

	for _, docID := range docIDs {
		prefix := keys.HeadstoreDocKey{DocID: docID}
		iter, err := txn.Headstore().Iterator(ctx, corekv.IterOptions{
			Prefix: prefix.Bytes(),
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
