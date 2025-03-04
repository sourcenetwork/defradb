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
	"testing"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

func makeCompositeBlock(t *testing.T, lsys *linking.LinkSystem) Block {
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
	require.NoError(t, err)

	compositeBlock := Block{
		Delta: crdt.CRDT{
			CompositeDAGDelta: &crdt.CompositeDAGDelta{
				DocID:           []byte("docID"),
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
	require.NoError(t, err)

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
		Heads: []cidlink.Link{
			fieldBlockLink.(cidlink.Link),
		},
	}
	fieldUpdateBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), fieldUpdateBlock.GenerateNode())
	require.NoError(t, err)

	return Block{
		Delta: crdt.CRDT{
			CompositeDAGDelta: &crdt.CompositeDAGDelta{
				DocID:           []byte("docID"),
				Priority:        2,
				SchemaVersionID: "schemaVersionID",
				Status:          1,
			},
		},
		Heads: []cidlink.Link{
			compositeBlockLink.(cidlink.Link),
		},
		Links: []DAGLink{
			{
				Name: "name",
				Link: fieldUpdateBlockLink.(cidlink.Link),
			},
		},
	}
}

func storeBlock(t *testing.T, lsys *linking.LinkSystem, block Block) cidlink.Link {
	blockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), block.GenerateNode())
	require.NoError(t, err)

	return blockLink.(cidlink.Link)
}

func TestBlock(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	rootLink := storeBlock(t, &lsys, makeCompositeBlock(t, &lsys))

	nd, err := lsys.Load(ipld.LinkContext{}, rootLink, BlockSchemaPrototype)
	require.NoError(t, err)

	block, err := GetFromNode(nd)
	require.NoError(t, err)

	b, err := block.Marshal()
	require.NoError(t, err)

	newBlock, err := GetFromBytes(b)
	require.NoError(t, err)

	require.Equal(t, block, newBlock)

	newNode := bindnode.Wrap(block, BlockSchema)
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

	rootLink := storeBlock(t, &lsys, makeCompositeBlock(t, &lsys))

	nd, err := lsys.Load(ipld.LinkContext{}, rootLink, BlockSchemaPrototype)
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

	nd, err := lsys.Load(ipld.LinkContext{}, blockLink, BlockSchemaPrototype)
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

	nd, err := lsys.Load(ipld.LinkContext{}, fieldBlockLink, BlockSchemaPrototype)
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

func TestBlock_IsEncrypted(t *testing.T) {
	tests := []struct {
		name       string
		setupBlock func() Block
		wantResult bool
	}{
		{
			name: "with encryption link",
			setupBlock: func() Block {
				return Block{
					Delta: crdt.CRDT{
						LWWRegDelta: &crdt.LWWRegDelta{},
					},
					Encryption: &cidlink.Link{},
				}
			},
			wantResult: true,
		},
		{
			name: "without encryption link",
			setupBlock: func() Block {
				return Block{
					Delta: crdt.CRDT{
						LWWRegDelta: &crdt.LWWRegDelta{},
					},
				}
			},
			wantResult: false,
		},
		{
			name: "with other fields but no encryption",
			setupBlock: func() Block {
				return Block{
					Delta: crdt.CRDT{
						LWWRegDelta: &crdt.LWWRegDelta{},
					},
					Signature: &cidlink.Link{},
					Heads:     []cidlink.Link{{}},
				}
			},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := tt.setupBlock()
			result := block.IsEncrypted()
			require.Equal(t, tt.wantResult, result)
		})
	}
}

func TestBlock_Clone(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	// Create encryption block and link
	encBlock := Encryption{
		DocID: []byte("docID"),
		Key:   []byte("keyID"),
	}
	encBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), encBlock.GenerateNode())
	require.NoError(t, err, "Failed to store encryption block")
	encLink := encBlockLink.(cidlink.Link)

	// Create signature block and link
	sigBlock := Signature{
		Header: SignatureHeader{
			Type:     SignatureTypeEd25519,
			Identity: []byte("signer-id"),
		},
		Value: []byte("signature-value"),
	}
	sigBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), sigBlock.GenerateNode())
	require.NoError(t, err, "Failed to store signature block")
	sigLink := sigBlockLink.(cidlink.Link)

	// Create a dummy block and get its CID for Heads
	dummyBlock := Block{
		Delta: crdt.CRDT{
			LWWRegDelta: &crdt.LWWRegDelta{
				Data: []byte("dummy"),
			},
		},
	}
	dummyLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), dummyBlock.GenerateNode())
	require.NoError(t, err, "Failed to store dummy block")

	// Create an original block with all fields set
	original := Block{
		Delta: crdt.CRDT{
			LWWRegDelta: &crdt.LWWRegDelta{
				DocID:           []byte("docID"),
				FieldName:       "name",
				Priority:        1,
				SchemaVersionID: "schemaVersionID",
				Data:            []byte("John"),
			},
		},
		Heads: []cidlink.Link{dummyLink.(cidlink.Link)},
		Links: []DAGLink{{
			Name: "testLink",
			Link: dummyLink.(cidlink.Link),
		}},
		Encryption: &encLink,
		Signature:  &sigLink,
	}

	// Serialize the original block
	originalBytes, err := original.Marshal()
	require.NoError(t, err, "Failed to serialize original block")

	// Clone the block
	cloned := original.Clone()

	// Serialize the cloned block
	clonedBytes, err := cloned.Marshal()
	require.NoError(t, err, "Failed to serialize cloned block")

	// Compare serialized forms
	require.Equal(t, originalBytes, clonedBytes, "Serialized blocks should be identical")

	// Modify the original to verify deep copy
	original.Delta.LWWRegDelta.Data = []byte("Jane")

	// Serialize both again
	originalBytes, err = original.Marshal()
	require.NoError(t, err, "Failed to serialize original block after modification")
	clonedBytes, err = cloned.Marshal()
	require.NoError(t, err, "Failed to serialize cloned block after original modification")

	// Verify they are now different
	require.NotEqual(t, originalBytes, clonedBytes, "Modifying original should not affect clone")

	// Verify we can unmarshal both blocks successfully
	var unmarshaledOriginal Block
	var unmarshaledClone Block
	err = unmarshaledOriginal.Unmarshal(originalBytes)
	require.NoError(t, err, "Failed to unmarshal original block")
	err = unmarshaledClone.Unmarshal(clonedBytes)
	require.NoError(t, err, "Failed to unmarshal cloned block")

	// Verify the unmarshaled blocks have the expected values
	require.Equal(t, []byte("Jane"), unmarshaledOriginal.Delta.LWWRegDelta.Data)
	require.Equal(t, []byte("John"), unmarshaledClone.Delta.LWWRegDelta.Data)
}
