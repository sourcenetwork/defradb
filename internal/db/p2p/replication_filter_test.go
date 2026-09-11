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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

// fieldBlock returns a field block holding value under name, in the decoded form the filter
// walks and the encoded form a CAR carries. The delta is written with the same encoder the
// write path uses, so the test exercises a real round trip rather than its own encoding.
func fieldBlock(t *testing.T, name string, value client.NormalValue) (*coreblock.Block, blocks.Block) {
	t.Helper()
	data, err := client.NewFieldValue(client.LWW_REGISTER, value).Bytes()
	require.NoError(t, err)

	core := &coreblock.Block{
		Delta: crdt.CRDT{LWWDelta: &crdt.LWWDelta{FieldName: name, Priority: 1, Data: data}},
	}
	return core, encodeBlock(t, core)
}

// encodeBlock returns the encoded form of a block, as a CAR carries it.
func encodeBlock(t *testing.T, block *coreblock.Block) blocks.Block {
	t.Helper()
	raw, err := block.Marshal()
	require.NoError(t, err)
	link, err := block.GenerateLink()
	require.NoError(t, err)
	encoded, err := blocks.NewBlockWithCid(raw, link.Cid)
	require.NoError(t, err)
	return encoded
}

// compositeOver returns a composite block linking the given blocks under the given names.
func compositeOver(t *testing.T, names []string, children ...blocks.Block) (*coreblock.Block, cid.Cid) {
	t.Helper()
	core := &coreblock.Block{Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Priority: 1}}}
	for i, child := range children {
		core.Links = append(core.Links, coreblock.NewDAGLink(names[i], cidlink.Link{Cid: child.Cid()}))
	}
	link, err := core.GenerateLink()
	require.NoError(t, err)
	return core, link.Cid
}

// A composite block holds no values of its own, so without reading the blocks it links the
// filter sees a document it cannot judge.
func TestExtractFields_ReadsValuesFromTheLinkedFieldBlocks(t *testing.T) {
	_, number := fieldBlock(t, "blockNumber", client.NewNormalInt(25951718))
	_, address := fieldBlock(t, "address", client.NewNormalString("0xabc"))
	composite, root := compositeOver(t, []string{"blockNumber", "address"}, number, address)

	fields := extractFieldsFromBlock(composite, carWith(t, root, number, address))

	require.Equal(t, map[string]any{"blockNumber": uint64(25951718), "address": "0xabc"}, fields)
}

// Without a CAR the field blocks have not arrived, and a name carrying no value would read as
// a field the document does not have.
func TestExtractFields_IsEmptyWithoutACAR(t *testing.T) {
	_, number := fieldBlock(t, "blockNumber", client.NewNormalInt(25951718))
	composite, _ := compositeOver(t, []string{"blockNumber"}, number)

	require.Empty(t, extractFieldsFromBlock(composite, nil))
}

// An encrypted delta holds ciphertext. Decoding it would hand the filter a value the document
// does not have.
func TestExtractFields_SkipsAnEncryptedFieldBlock(t *testing.T) {
	_, plain := fieldBlock(t, "address", client.NewNormalString("0xabc"))
	secret, _ := fieldBlock(t, "blockNumber", client.NewNormalInt(25951718))
	secret.Encryption = &cidlink.Link{Cid: plain.Cid()}
	encrypted := encodeBlock(t, secret)

	composite, root := compositeOver(t, []string{"blockNumber", "address"}, encrypted, plain)

	fields := extractFieldsFromBlock(composite, carWith(t, root, encrypted, plain))

	require.Equal(t, map[string]any{"address": "0xabc"}, fields)
}
