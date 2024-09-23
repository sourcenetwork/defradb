// Copyright 2024 Democratized Data Foundation
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
	"testing"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

func generateBlocks(lsys *linking.LinkSystem) (cidlink.Link, error) {
	// Generate new Block and save to lsys
	fieldBlock := Block{
		Delta: crdt.CRDT{
			LWWRegDelta: &crdt.LWWRegDelta{
				DocID:           []byte("docID"),
				FieldName:       "name",
				Priority:        1,
				SchemaVersionID: "schemaVersionID",
				Data:            []byte("John"),
			},
		},
	}
	fieldBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), fieldBlock.GenerateNode())
	if err != nil {
		return cidlink.Link{}, err
	}

	compositeBlock := Block{
		Delta: crdt.CRDT{
			CompositeDAGDelta: &crdt.CompositeDAGDelta{
				DocID:           []byte("docID"),
				FieldName:       "C",
				Priority:        1,
				SchemaVersionID: "schemaVersionID",
				Status:          1,
			},
		},
		Links: []DAGLink{
			{
				Name: "name",
				Link: fieldBlockLink.(cidlink.Link),
			},
		},
	}
	compositeBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), compositeBlock.GenerateNode())
	if err != nil {
		return cidlink.Link{}, err
	}

	fieldUpdateBlock := Block{
		Delta: crdt.CRDT{
			LWWRegDelta: &crdt.LWWRegDelta{
				DocID:           []byte("docID"),
				FieldName:       "name",
				Priority:        2,
				SchemaVersionID: "schemaVersionID",
				Data:            []byte("Johny"),
			},
		},
		Links: []DAGLink{
			{
				Name: core.HEAD,
				Link: fieldBlockLink.(cidlink.Link),
			},
		},
	}
	fieldUpdateBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), fieldUpdateBlock.GenerateNode())
	if err != nil {
		return cidlink.Link{}, err
	}

	compositeUpdateBlock := Block{
		Delta: crdt.CRDT{
			CompositeDAGDelta: &crdt.CompositeDAGDelta{
				DocID:           []byte("docID"),
				FieldName:       "C",
				Priority:        2,
				SchemaVersionID: "schemaVersionID",
				Status:          1,
			},
		},
		Links: []DAGLink{
			{
				Name: core.HEAD,
				Link: compositeBlockLink.(cidlink.Link),
			},
			{
				Name: "name",
				Link: fieldUpdateBlockLink.(cidlink.Link),
			},
		},
	}
	compositeUpdateBlockLink, err := lsys.Store(
		ipld.LinkContext{},
		GetLinkPrototype(),
		compositeUpdateBlock.GenerateNode(),
	)
	if err != nil {
		return cidlink.Link{}, err
	}

	return compositeUpdateBlockLink.(cidlink.Link), nil
}

func TestBlock(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	rootLink, err := generateBlocks(&lsys)
	require.NoError(t, err)

	nd, err := lsys.Load(ipld.LinkContext{}, rootLink, SchemaPrototype)
	require.NoError(t, err)

	block, err := GetFromNode(nd)
	require.NoError(t, err)

	b, err := block.Marshal()
	require.NoError(t, err)

	newBlock, err := GetFromBytes(b)
	require.NoError(t, err)

	require.Equal(t, block, newBlock)

	newNode := bindnode.Wrap(block, Schema)
	require.Equal(t, nd, newNode)

	link, err := block.GenerateLink()
	require.NoError(t, err)
	require.Equal(t, rootLink, link)

	newLink, err := GetLinkFromNode(newNode)
	require.NoError(t, err)
	require.Equal(t, rootLink, newLink)
}

func TestGetFromNode_WithInvalidType_ShouldFail(t *testing.T) {
	_, err := GetFromNode(basicnode.NewString("test"))
	require.ErrorIs(t, err, ErrNodeToBlock)
}

func TestBlockDeltaPriority(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	rootLink, err := generateBlocks(&lsys)
	require.NoError(t, err)

	nd, err := lsys.Load(ipld.LinkContext{}, rootLink, SchemaPrototype)
	require.NoError(t, err)

	block, err := GetFromNode(nd)
	require.NoError(t, err)

	// The generateBlocks function creates a block with one update
	// which results in a priority of 2.
	require.Equal(t, uint64(2), block.Delta.GetPriority())
}

func TestBlockMarshal_IfEncryptedNotSet_ShouldNotContainIsEncryptedField(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	encBlock := Encryption{
		DocID: []byte("docID"),
		Key:   []byte("keyID"),
	}

	encBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), encBlock.GenerateNode())
	require.NoError(t, err)

	link := encBlockLink.(cidlink.Link)

	block := Block{
		Delta: crdt.CRDT{
			LWWRegDelta: &crdt.LWWRegDelta{
				DocID:           []byte("docID"),
				FieldName:       "name",
				Priority:        1,
				SchemaVersionID: "schemaVersionID",
				Data:            []byte("John"),
			},
		},
		Encryption: &link,
	}

	blockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), block.GenerateNode())
	require.NoError(t, err)

	nd, err := lsys.Load(ipld.LinkContext{}, blockLink, SchemaPrototype)
	require.NoError(t, err)

	loadedBlock, err := GetFromNode(nd)
	require.NoError(t, err)

	require.NotNil(t, loadedBlock.Encryption)

	nd, err = lsys.Load(ipld.LinkContext{}, loadedBlock.Encryption, EncryptionSchemaPrototype)
	require.NoError(t, err)

	loadedEncBlock, err := GetEncryptionBlockFromNode(nd)
	require.NoError(t, err)

	require.Equal(t, encBlock, *loadedEncBlock)
}

func TestBlockMarshal_IsEncryptedNotSetWithLinkSystem_ShouldLoadWithNoError(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	fieldBlock := Block{
		Delta: crdt.CRDT{
			LWWRegDelta: &crdt.LWWRegDelta{
				DocID:           []byte("docID"),
				FieldName:       "name",
				Priority:        1,
				SchemaVersionID: "schemaVersionID",
				Data:            []byte("John"),
			},
		},
	}
	fieldBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), fieldBlock.GenerateNode())
	require.NoError(t, err)

	nd, err := lsys.Load(ipld.LinkContext{}, fieldBlockLink, SchemaPrototype)
	require.NoError(t, err)
	_, err = GetFromNode(nd)
	require.NoError(t, err)
}

func TestBlockUnmarshal_ValidInput_Succeed(t *testing.T) {
	validBlock := Block{
		Delta: crdt.CRDT{
			LWWRegDelta: &crdt.LWWRegDelta{
				DocID:           []byte("docID"),
				FieldName:       "name",
				Priority:        1,
				SchemaVersionID: "schemaVersionID",
				Data:            []byte("John"),
			},
		},
	}

	marshaledData, err := validBlock.Marshal()
	require.NoError(t, err)

	var unmarshaledBlock Block
	err = unmarshaledBlock.Unmarshal(marshaledData)
	require.NoError(t, err)

	require.Equal(t, validBlock, unmarshaledBlock)
}

func TestBlockUnmarshal_InvalidCBOR_Error(t *testing.T) {
	invalidData := []byte("invalid CBOR data")
	var block Block
	err := block.Unmarshal(invalidData)
	require.Error(t, err)
}

func TestEncryptionBlockUnmarshal_InvalidCBOR_Error(t *testing.T) {
	invalidData := []byte("invalid CBOR data")
	var encBlock Encryption
	err := encBlock.Unmarshal(invalidData)
	require.Error(t, err)
}

func TestEncryptionBlockUnmarshal_ValidInput_Succeed(t *testing.T) {
	fieldName := "fieldName"
	encBlock := Encryption{
		DocID:     []byte("docID"),
		Key:       []byte("keyID"),
		FieldName: &fieldName,
	}

	marshaledData, err := encBlock.Marshal()
	require.NoError(t, err)

	var unmarshaledBlock Encryption
	err = unmarshaledBlock.Unmarshal(marshaledData)
	require.NoError(t, err)

	require.Equal(t, encBlock, unmarshaledBlock)
}
