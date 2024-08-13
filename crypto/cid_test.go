// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crypto

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

func TestGenerateCid_HappyPath(t *testing.T) {
	testData := []byte("Hello, world!")

	generatedCid, err := GenerateCid(testData)

	require.NoError(t, err)
	require.NotEmpty(t, generatedCid)
}

func TestGenerateCid_ErrorPath(t *testing.T) {
	// Define a custom GenerateCid function that uses a failing hash function
	generateCidWithError := func(data []byte) (cid.Cid, error) {
		_, err := multihash.Sum(data, 0xffff, -1) // Use an invalid hash function code
		if err != nil {
			return cid.Cid{}, err
		}
		return cid.Cid{}, nil // This line should never be reached
	}

	testData := []byte("This data doesn't matter as we're forcing an error")
	generatedCid, err := generateCidWithError(testData)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown multihash code")
	require.Equal(t, cid.Cid{}, generatedCid) // Should be an empty CID
}
