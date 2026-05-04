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
	"testing"

	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

func buildFilterBlock(t *testing.T, block *coreblock.Block) (cid.Cid, blocks.Block) {
	t.Helper()

	rawBytes, err := block.Marshal()
	require.NoError(t, err)

	link, err := block.GenerateLink()
	require.NoError(t, err)

	rawBlock, err := blocks.NewBlockWithCid(rawBytes, link.Cid)
	require.NoError(t, err)

	return link.Cid, rawBlock
}

func buildFilterFieldBlock(t *testing.T, fieldName string, value any, priority uint64) (cid.Cid, blocks.Block) {
	t.Helper()

	data, err := cbor.Marshal(value)
	require.NoError(t, err)

	block := &coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:               []byte("testDoc"),
				FieldName:           fieldName,
				Priority:            priority,
				CollectionVersionID: "v1",
				Data:                data,
			},
		},
	}

	return buildFilterBlock(t, block)
}

func buildFilterCompositeBlock(
	t *testing.T,
	priority uint64,
	heads []cidlink.Link,
	links []coreblock.DAGLink,
) (*coreblock.Block, cid.Cid, blocks.Block) {
	t.Helper()

	block := &coreblock.Block{
		Delta: crdt.CRDT{
			DocCompositeDelta: &crdt.DocCompositeDelta{
				DocID:               []byte("testDoc"),
				Priority:            priority,
				CollectionVersionID: "v1",
				Status:              1,
			},
		},
		Heads: heads,
		Links: links,
	}

	blockCID, rawBlock := buildFilterBlock(t, block)
	return block, blockCID, rawBlock
}

func TestExtractFieldsFromCAR_UsesRootLinksNotHistoricalHeads(t *testing.T) {
	oldFieldCID, oldFieldBlock := buildFilterFieldBlock(t, "name", "Alice", 1)
	_, oldCompositeCID, oldCompositeBlock := buildFilterCompositeBlock(
		t,
		1,
		nil,
		[]coreblock.DAGLink{coreblock.NewDAGLink("name", cidlink.Link{Cid: oldFieldCID})},
	)

	newFieldCID, newFieldBlock := buildFilterFieldBlock(t, "name", "Bob", 2)
	rootBlock, _, rootRawBlock := buildFilterCompositeBlock(
		t,
		2,
		[]cidlink.Link{{Cid: oldCompositeCID}},
		[]coreblock.DAGLink{coreblock.NewDAGLink("name", cidlink.Link{Cid: newFieldCID})},
	)

	parsed := &parsedCAR{
		rootBlock: rootBlock,
		regularBlocks: []blocks.Block{
			newFieldBlock,
			rootRawBlock,
			oldFieldBlock,
			oldCompositeBlock,
		},
	}

	fields := extractFieldsFromCAR(parsed)

	require.Equal(t, map[string]any{"name": "Bob"}, fields)
}
