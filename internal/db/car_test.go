// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"bytes"
	"testing"

	gocar "github.com/ipld/go-car/v2"
	"github.com/stretchr/testify/require"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

func buildSavedBlock(t *testing.T, block *coreblock.Block) savedBlock {
	t.Helper()
	raw, err := block.Marshal()
	require.NoError(t, err)
	link, err := block.GenerateLink()
	require.NoError(t, err)
	return savedBlock{cid: link.Cid, data: raw}
}

func makeCompositeBlock() *coreblock.Block {
	return &coreblock.Block{
		Delta: crdt.CRDT{
			DocCompositeDelta: &crdt.DocCompositeDelta{
				DocID:               []byte("doc1"),
				Priority:            1,
				CollectionVersionID: "v1",
				Status:              1,
			},
		},
	}
}

func makeFieldBlock(fieldName string, data []byte) *coreblock.Block {
	return &coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:               []byte("doc1"),
				FieldName:           fieldName,
				Priority:            1,
				CollectionVersionID: "v1",
				Data:                data,
			},
		},
	}
}

func TestGenerateCAR_SingleBlock(t *testing.T) {
	rootBlk := makeCompositeBlock()
	root := buildSavedBlock(t, rootBlk)

	carData, err := generateCAR(t.Context(), root.cid, []savedBlock{root})
	require.NoError(t, err)
	require.NotEmpty(t, carData)

	reader, err := gocar.NewBlockReader(bytes.NewReader(carData))
	require.NoError(t, err)
	require.Len(t, reader.Roots, 1)
	require.Equal(t, root.cid, reader.Roots[0])

	block, err := reader.Next()
	require.NoError(t, err)
	require.Equal(t, root.cid, block.Cid())
	require.Equal(t, root.data, block.RawData())
}

func TestGenerateCAR_MultipleBlocks(t *testing.T) {
	rootBlk := makeCompositeBlock()
	root := buildSavedBlock(t, rootBlk)

	field1 := buildSavedBlock(t, makeFieldBlock("name", []byte("Alice")))
	field2 := buildSavedBlock(t, makeFieldBlock("age", []byte{42}))

	blocks := []savedBlock{root, field1, field2}
	carData, err := generateCAR(t.Context(), root.cid, blocks)
	require.NoError(t, err)
	require.NotEmpty(t, carData)

	reader, err := gocar.NewBlockReader(bytes.NewReader(carData))
	require.NoError(t, err)
	require.Equal(t, root.cid, reader.Roots[0])

	found := make(map[string][]byte)
	for {
		blk, err := reader.Next()
		if err != nil {
			break
		}
		found[blk.Cid().String()] = blk.RawData()
	}

	require.Len(t, found, 3)
	require.Equal(t, root.data, found[root.cid.String()])
	require.Equal(t, field1.data, found[field1.cid.String()])
	require.Equal(t, field2.data, found[field2.cid.String()])
}

func TestGenerateCAR_EmptyBlocks(t *testing.T) {
	rootBlk := makeCompositeBlock()
	root := buildSavedBlock(t, rootBlk)

	carData, err := generateCAR(t.Context(), root.cid, []savedBlock{})
	require.NoError(t, err)
	require.NotEmpty(t, carData)

	reader, err := gocar.NewBlockReader(bytes.NewReader(carData))
	require.NoError(t, err)
	require.Len(t, reader.Roots, 1)
	require.Equal(t, root.cid, reader.Roots[0])
}
