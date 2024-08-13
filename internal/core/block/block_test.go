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

func TestBlockMarshal_IsEncryptedNotSet_ShouldNotContainIsEcryptedField(t *testing.T) {
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

	b, err := fieldBlock.Marshal()
	require.NoError(t, err)
	require.NotContains(t, string(b), "isEncrypted")
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

func TestBlock_Validate(t *testing.T) {
	tests := []struct {
		name          string
		encryption    *Encryption
		expectedError error
	}{
		{
			name:          "NotEncrypted type is valid",
			encryption:    &Encryption{Type: NotEncrypted, KeyID: []byte{1}},
			expectedError: nil,
		},
		{
			name:          "DocumentEncrypted type is valid",
			encryption:    &Encryption{Type: DocumentEncrypted, KeyID: []byte{1}},
			expectedError: nil,
		},
		{
			name:          "FieldEncrypted type is valid",
			encryption:    &Encryption{Type: FieldEncrypted, KeyID: []byte{1}},
			expectedError: nil,
		},
		{
			name:          "Nil Encryption is valid",
			encryption:    nil,
			expectedError: nil,
		},
		{
			name:          "Invalid encryption type",
			encryption:    &Encryption{Type: EncryptionType(99), KeyID: []byte{1}},
			expectedError: ErrInvalidBlockEncryptionType,
		},
		{
			name:          "Invalid encryption key id parameter",
			encryption:    &Encryption{Type: DocumentEncrypted, KeyID: []byte{}},
			expectedError: ErrInvalidBlockEncryptionKeyID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				Encryption: tt.encryption,
			}
			err := b.Validate()
			require.Equal(t, tt.expectedError, err)
		})
	}
}
