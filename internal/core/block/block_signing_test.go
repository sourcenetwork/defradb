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
	"crypto/rand"
	"crypto/sha256"
	"testing"

	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeCIDs generates n deterministic CIDs for testing.
func makeCIDs(t *testing.T, n int) []cid.Cid {
	t.Helper()
	cids := make([]cid.Cid, n)
	for i := range n {
		buf := make([]byte, 32)
		_, err := rand.Read(buf)
		require.NoError(t, err)
		hash, err := mh.Sum(buf, mh.SHA2_256, -1)
		require.NoError(t, err)
		cids[i] = cid.NewCidV1(cid.DagCBOR, hash)
	}
	return cids
}

func TestGenerateMerkleProof_SingleCID(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 1))
	root := ComputeMerkleRoot(cids)

	proof := GenerateMerkleProof(cids, 0)
	require.NotNil(t, proof)
	assert.Empty(t, proof.Siblings)
	assert.True(t, VerifyMerkleProof(cids[0], proof, root))
}

func TestGenerateMerkleProof_TwoCIDs(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 2))
	root := ComputeMerkleRoot(cids)

	for i := range cids {
		proof := GenerateMerkleProof(cids, i)
		require.NotNil(t, proof)
		assert.Len(t, proof.Siblings, 1)
		assert.True(t, VerifyMerkleProof(cids[i], proof, root),
			"proof failed for index %d", i)
	}
}

func TestGenerateMerkleProof_EvenCount(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 4))
	root := ComputeMerkleRoot(cids)

	for i := range cids {
		proof := GenerateMerkleProof(cids, i)
		require.NotNil(t, proof)
		assert.True(t, VerifyMerkleProof(cids[i], proof, root),
			"proof failed for index %d", i)
	}
}

func TestGenerateMerkleProof_OddCount(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 5))
	root := ComputeMerkleRoot(cids)

	for i := range cids {
		proof := GenerateMerkleProof(cids, i)
		require.NotNil(t, proof)
		assert.True(t, VerifyMerkleProof(cids[i], proof, root),
			"proof failed for index %d", i)
	}
}

func TestGenerateMerkleProof_LargeSet(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 100))
	root := ComputeMerkleRoot(cids)

	// Verify a subset of indices.
	for _, i := range []int{0, 1, 49, 50, 98, 99} {
		proof := GenerateMerkleProof(cids, i)
		require.NotNil(t, proof)
		assert.True(t, VerifyMerkleProof(cids[i], proof, root),
			"proof failed for index %d", i)
	}
}

func TestVerifyMerkleProof_WrongRoot(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 4))

	proof := GenerateMerkleProof(cids, 0)
	require.NotNil(t, proof)

	fakeRoot := sha256.Sum256([]byte("wrong"))
	assert.False(t, VerifyMerkleProof(cids[0], proof, fakeRoot[:]))
}

func TestVerifyMerkleProof_WrongLeaf(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 4))
	root := ComputeMerkleRoot(cids)

	proof := GenerateMerkleProof(cids, 0)
	require.NotNil(t, proof)

	// Use a different CID as the leaf.
	wrongCID := makeCIDs(t, 1)[0]
	assert.False(t, VerifyMerkleProof(wrongCID, proof, root))
}

func TestGenerateMerkleProof_OutOfRange(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 3))
	assert.Nil(t, GenerateMerkleProof(cids, -1))
	assert.Nil(t, GenerateMerkleProof(cids, 3))
	assert.Nil(t, GenerateMerkleProof(nil, 0))
}

func TestVerifyMerkleProof_NilProof(t *testing.T) {
	cids := sortCIDs(makeCIDs(t, 1))
	root := ComputeMerkleRoot(cids)
	assert.False(t, VerifyMerkleProof(cids[0], nil, root))
}

func TestSortedCIDStrings_Deterministic(t *testing.T) {
	cids := makeCIDs(t, 10)

	result1 := SortedCIDStrings(cids)

	// Reverse the input order.
	reversed := make([]cid.Cid, len(cids))
	for i, c := range cids {
		reversed[len(cids)-1-i] = c
	}
	result2 := SortedCIDStrings(reversed)

	assert.Equal(t, result1, result2)
}

func TestGenerateMerkleProof_ThreeCIDs(t *testing.T) {
	// 3 CIDs: the third is the odd carry-forward at level 0.
	cids := sortCIDs(makeCIDs(t, 3))
	root := ComputeMerkleRoot(cids)

	for i := range cids {
		proof := GenerateMerkleProof(cids, i)
		require.NotNil(t, proof)
		assert.True(t, VerifyMerkleProof(cids[i], proof, root),
			"proof failed for index %d", i)
	}
}
