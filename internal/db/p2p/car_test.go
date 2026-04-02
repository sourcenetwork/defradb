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
	"testing"

	"github.com/ipfs/go-cid"
	gocar "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

// buildTestBlock constructs a minimal coreblock.Block, marshals it to bytes,
// and derives its CID using the standard DAG-CBOR link prototype.
func buildTestBlock(t *testing.T) (*coreblock.Block, cid.Cid, []byte) {
	t.Helper()

	block := &coreblock.Block{
		Delta: crdt.CRDT{
			DocCompositeDelta: &crdt.DocCompositeDelta{
				DocID:               []byte("testDoc"),
				Priority:            1,
				CollectionVersionID: "v1",
				Status:              1,
			},
		},
	}

	rawBytes, err := block.Marshal()
	require.NoError(t, err)

	link, err := block.GenerateLink()
	require.NoError(t, err)

	return block, link.Cid, rawBytes
}

// buildExtraBlock constructs a non-root coreblock.Block (a field-level LWW block).
func buildExtraBlock(t *testing.T) (cid.Cid, []byte) {
	t.Helper()

	block := &coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:               []byte("testDoc"),
				FieldName:           "name",
				Priority:            1,
				CollectionVersionID: "v1",
				Data:                []byte("Alice"),
			},
		},
	}

	rawBytes, err := block.Marshal()
	require.NoError(t, err)

	link, err := block.GenerateLink()
	require.NoError(t, err)

	return link.Cid, rawBytes
}

// buildCARFromBlocks creates a CARv1 byte slice with the given root and blocks
// without relying on any db package.
func buildCARFromBlocks(t *testing.T, rootCID cid.Cid, blocks map[cid.Cid][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer, err := storage.NewWritable(&buf, []cid.Cid{rootCID}, gocar.WriteAsCarV1(true))
	require.NoError(t, err)

	for c, data := range blocks {
		link := cidlink.Link{Cid: c}
		err := writer.Put(t.Context(), link.Binary(), data)
		require.NoError(t, err)
	}

	return buf.Bytes()
}

func TestParseCAR_RoundTrip(t *testing.T) {
	_, rootCID, rootBytes := buildTestBlock(t)
	extraCID, extraBytes := buildExtraBlock(t)

	allBlocks := map[cid.Cid][]byte{
		rootCID:  rootBytes,
		extraCID: extraBytes,
	}

	carData := buildCARFromBlocks(t, rootCID, allBlocks)

	result, err := parseCAR(carData)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.rootBlock)

	// Verify the root block CID matches what we put in.
	rootLink, err := result.rootBlock.GenerateLink()
	require.NoError(t, err)
	require.Equal(t, rootCID, rootLink.Cid)

	// All regular blocks (root + extra) should be present.
	require.Len(t, result.regularBlocks, 2)

	cidSet := make(map[cid.Cid]struct{})
	for _, blk := range result.regularBlocks {
		cidSet[blk.Cid()] = struct{}{}
	}
	require.Contains(t, cidSet, rootCID)
	require.Contains(t, cidSet, extraCID)
}

func TestParseCAR_EmptyData(t *testing.T) {
	_, err := parseCAR([]byte{})
	require.Error(t, err)
}

func TestParseCAR_NilData(t *testing.T) {
	_, err := parseCAR(nil)
	require.Error(t, err)
}
