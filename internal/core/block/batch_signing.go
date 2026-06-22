// Copyright 2025 Democratized Data Foundation
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

	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type batchSigningContextKey struct{}

// BatchCIDCollector accumulates block CIDs produced during a batch operation.
// It is safe for concurrent use.
type BatchCIDCollector struct {
	mu   sync.Mutex
	cids []cid.Cid
}

// NewBatchCIDCollector returns a new, empty BatchCIDCollector.
func NewBatchCIDCollector() *BatchCIDCollector {
	return &BatchCIDCollector{}
}

// Add appends a CID to the collector. Safe to call from multiple goroutines.
func (c *BatchCIDCollector) Add(id cid.Cid) {
	c.mu.Lock()
	c.cids = append(c.cids, id)
	c.mu.Unlock()
}

// GetCIDs returns a copy of the collected CIDs.
func (c *BatchCIDCollector) GetCIDs() []cid.Cid {
	c.mu.Lock()
	out := make([]cid.Cid, len(c.cids))
	copy(out, c.cids)
	c.mu.Unlock()
	return out
}

// ContextWithBatchSigning embeds a BatchCIDCollector into the context so that
// code deep in the call stack can record CIDs without threading extra parameters.
func ContextWithBatchSigning(ctx context.Context, collector *BatchCIDCollector) context.Context {
	return context.WithValue(ctx, batchSigningContextKey{}, collector)
}

// BatchSigningCollectorFromContext extracts the BatchCIDCollector from ctx, returning
// nil when batch signing is not enabled.
func BatchSigningCollectorFromContext(ctx context.Context) *BatchCIDCollector {
	v := ctx.Value(batchSigningContextKey{})
	if v == nil {
		return nil
	}
	return v.(*BatchCIDCollector) //nolint:forcetypeassert
}

// IsBatchSigningEnabled returns true if a BatchCIDCollector is present in ctx.
func IsBatchSigningEnabled(ctx context.Context) bool {
	return ctx.Value(batchSigningContextKey{}) != nil
}

// BatchSignature holds the result of signing a batch of documents.
type BatchSignature struct {
	// Header identifies the signer.
	Header SignatureHeader
	// Value is the cryptographic signature over MerkleRoot.
	Value []byte
	// MerkleRoot is the root hash of the Merkle tree built from the batch CIDs.
	MerkleRoot []byte
	// CIDCount is the number of CIDs that were included in the Merkle tree.
	CIDCount int
}

// IPLDSchemaBytes returns a minimal IPLD schema for BatchSignature.
func (sig *BatchSignature) IPLDSchemaBytes() []byte {
	return []byte(`
		type BatchSignature struct {
			header     SignatureHeader
			value      Bytes
			merkleRoot Bytes
			cidCount   Int
		}
	`)
}

// ComputeMerkleRoot builds a deterministic binary Merkle tree from cids and returns
// the SHA-256 root hash.  The CIDs are sorted by their byte representation before
// tree construction so the root is independent of insertion order.
func ComputeMerkleRoot(cids []cid.Cid) []byte {
	if len(cids) == 0 {
		h := sha256.Sum256(nil)
		return h[:]
	}

	// Sort deterministically.
	sorted := make([]cid.Cid, len(cids))
	copy(sorted, cids)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i].Bytes(), sorted[j].Bytes()) < 0
	})

	// Leaf layer: hash each CID.
	layer := make([][]byte, len(sorted))
	for i, c := range sorted {
		h := sha256.Sum256(c.Bytes())
		layer[i] = h[:]
	}

	// Combine pairs until one root remains.
	for len(layer) > 1 {
		var next [][]byte
		for i := 0; i < len(layer); i += 2 {
			left := layer[i]
			right := left // duplicate last if odd count
			if i+1 < len(layer) {
				right = layer[i+1]
			}
			combined := append(left, right...) //nolint:gocritic
			h := sha256.Sum256(combined)
			next = append(next, h[:])
		}
		layer = next
	}

	return layer[0]
}

// SignBatch computes the Merkle root of the CIDs collected in collector and signs
// it using the identity stored in ctx.  The identity must be a FullIdentity with a
// non-nil private key (same requirement as per-block signing).
func SignBatch(ctx context.Context, collector *BatchCIDCollector) (*BatchSignature, error) {
	_, identOpt := EnabledSigningFromContext(ctx)
	if !identOpt.HasValue() {
		return nil, newErrBatchSigningNoIdentity()
	}
	ident := identOpt.Value()

	cids := collector.GetCIDs()
	root := ComputeMerkleRoot(cids)

	var sigType string
	switch ident.PrivateKey().Type() {
	case crypto.KeyTypeSecp256k1:
		sigType = SignatureTypeECDSA256K
	case crypto.KeyTypeEd25519:
		sigType = SignatureTypeEd25519
	default:
		return nil, NewErrUnsupportedKeyForSigning(ident.PrivateKey().Type())
	}

	sigBytes, err := ident.PrivateKey().Sign(root)
	if err != nil {
		return nil, err
	}

	return &BatchSignature{
		Header: SignatureHeader{
			Type:     sigType,
			Identity: []byte(ident.PublicKey().String()),
		},
		Value:      sigBytes,
		MerkleRoot: root,
		CIDCount:   len(cids),
	}, nil
}

// VerifyBatchSignature reconstructs the Merkle root from cids and verifies it
// against the stored root and cryptographic signature in batchSig.
func VerifyBatchSignature(batchSig *BatchSignature, cids []cid.Cid) (bool, error) {
	root := ComputeMerkleRoot(cids)
	if !bytes.Equal(root, batchSig.MerkleRoot) {
		return false, nil
	}

	pubKey, err := getPublicKeyFromSignature(&Signature{
		Header: batchSig.Header,
		Value:  batchSig.Value,
	})
	if err != nil {
		return false, err
	}

	return pubKey.Verify(batchSig.MerkleRoot, batchSig.Value)
}

// CollectDocumentCIDs returns the head block CIDs for the given document IDs by
// scanning the headstore.  Pass txn.Headstore() as headstore.
func CollectDocumentCIDs(
	ctx context.Context,
	headstore corekv.ReaderWriter,
	docIDs []string,
) ([]cid.Cid, error) {
	seen := make(map[cid.Cid]struct{})
	var out []cid.Cid

	for _, docID := range docIDs {
		prefix := keys.HeadstoreDocKey{DocID: docID}
		iter, err := headstore.Iterator(ctx, corekv.IterOptions{
			Prefix: prefix.Bytes(),
		})
		if err != nil {
			return nil, err
		}

		for {
			hasNext, err := iter.Next()
			if err != nil {
				return nil, err
			}
			if !hasNext {
				break
			}
			k := string(iter.Key())
			hk, err := keys.NewHeadstoreDocKey(k)
			if err != nil {
				// skip keys that don't parse (e.g. collection head keys)
				continue
			}
			if !hk.Cid.Defined() {
				continue
			}
			if _, ok := seen[hk.Cid]; !ok {
				seen[hk.Cid] = struct{}{}
				out = append(out, hk.Cid)
			}
		}
		if err := iter.Close(); err != nil {
			return nil, err
		}
	}

	return out, nil
}
