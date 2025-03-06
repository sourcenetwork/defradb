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
	"crypto/ed25519"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

type signatureTestKeys struct {
	ed25519Pub  ed25519.PublicKey
	ed25519Priv ed25519.PrivateKey
	ecdsaKey    *secp256k1.PrivateKey
}

func setupTestKeys(t *testing.T) *signatureTestKeys {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	ecdsaKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	return &signatureTestKeys{
		ed25519Pub:  pubKey,
		ed25519Priv: privKey,
		ecdsaKey:    ecdsaKey,
	}
}

func createSignedBlock(t *testing.T, lsys *linking.LinkSystem, block *Block, sigType string, keys *signatureTestKeys) {
	blockBytes, err := block.Marshal()
	require.NoError(t, err)

	var sigBlock *Signature
	switch sigType {
	case SignatureTypeEd25519:
		sigBlock = &Signature{
			Header: SignatureHeader{
				Type:     SignatureTypeEd25519,
				Identity: keys.ed25519Pub,
			},
			Value: ed25519.Sign(keys.ed25519Priv, blockBytes),
		}
	case SignatureTypeECDSA256K:
		sig, err := crypto.Sign(crypto.SignatureTypeECDSA256K, keys.ecdsaKey, blockBytes)
		require.NoError(t, err)
		sigBlock = &Signature{
			Header: SignatureHeader{
				Type:     SignatureTypeECDSA256K,
				Identity: keys.ecdsaKey.PubKey().SerializeCompressed(),
			},
			Value: sig,
		}
	}

	sigBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), sigBlock.GenerateNode())
	require.NoError(t, err)
	sigLink, ok := sigBlockLink.(cidlink.Link)
	require.True(t, ok)
	block.Signature = &sigLink
}

func setupTestEnv(t *testing.T) (*linking.LinkSystem, *signatureTestKeys) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)
	return &lsys, setupTestKeys(t)
}

func TestSignatureBlockUnmarshal_InvalidCBOR_Error(t *testing.T) {
	invalidData := []byte("invalid CBOR data")
	var sigBlock Signature
	err := sigBlock.Unmarshal(invalidData)
	require.Error(t, err)
}

func TestSignatureBlockUnmarshal_ValidInput_Succeed(t *testing.T) {
	sigBlock := Signature{
		Header: SignatureHeader{
			Type:     SignatureTypeEd25519,
			Identity: []byte("signer-id"),
		},
		Value: []byte("signature-value"),
	}

	marshaledData, err := sigBlock.Marshal()
	require.NoError(t, err)

	var unmarshaledBlock Signature
	err = unmarshaledBlock.Unmarshal(marshaledData)
	require.NoError(t, err)

	require.Equal(t, sigBlock, unmarshaledBlock)
}

func TestBlockMarshal_IfSignatureNotSet_ShouldNotContainSignatureField(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	sigBlock := Signature{
		Header: SignatureHeader{
			Type:     SignatureTypeECDSA256K,
			Identity: []byte("pubkey-bytes"),
		},
		Value: []byte("signature-bytes"),
	}

	sigBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), sigBlock.GenerateNode())
	require.NoError(t, err)

	link, ok := sigBlockLink.(cidlink.Link)
	require.True(t, ok)

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
		Signature: &link,
	}

	blockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), block.GenerateNode())
	require.NoError(t, err)

	nd, err := lsys.Load(ipld.LinkContext{}, blockLink, BlockSchemaPrototype)
	require.NoError(t, err)

	loadedBlock, err := GetFromNode(nd)
	require.NoError(t, err)

	require.NotNil(t, loadedBlock.Signature)

	nd, err = lsys.Load(ipld.LinkContext{}, loadedBlock.Signature, SignatureSchemaPrototype)
	require.NoError(t, err)

	loadedSigBlock, err := GetSignatureBlockFromNode(nd)
	require.NoError(t, err)

	require.Equal(t, sigBlock, *loadedSigBlock)
}

func TestBlockWithSignatureAndEncryption(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetReadStorage(&store)
	lsys.SetWriteStorage(&store)

	// Create encryption block
	encBlock := Encryption{
		DocID: []byte("docID"),
		Key:   []byte("keyID"),
	}
	encBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), encBlock.GenerateNode())
	require.NoError(t, err)
	encLink, ok := encBlockLink.(cidlink.Link)
	require.True(t, ok)

	// Create signature block
	sigBlock := Signature{
		Header: SignatureHeader{
			Type:     SignatureTypeEd25519,
			Identity: []byte("signer-id"),
		},
		Value: []byte("signature-value"),
	}
	sigBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), sigBlock.GenerateNode())
	require.NoError(t, err)
	sigLink, ok := sigBlockLink.(cidlink.Link)
	require.True(t, ok)

	// Create block with both encryption and signature
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
		Encryption: &encLink,
		Signature:  &sigLink,
	}

	blockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), block.GenerateNode())
	require.NoError(t, err)

	nd, err := lsys.Load(ipld.LinkContext{}, blockLink, BlockSchemaPrototype)
	require.NoError(t, err)

	loadedBlock, err := GetFromNode(nd)
	require.NoError(t, err)

	// Verify both links are preserved
	require.NotNil(t, loadedBlock.Encryption)
	require.NotNil(t, loadedBlock.Signature)

	// Load and verify encryption block
	nd, err = lsys.Load(ipld.LinkContext{}, loadedBlock.Encryption, EncryptionSchemaPrototype)
	require.NoError(t, err)
	loadedEncBlock, err := GetEncryptionBlockFromNode(nd)
	require.NoError(t, err)
	require.Equal(t, encBlock, *loadedEncBlock)

	// Load and verify signature block
	nd, err = lsys.Load(ipld.LinkContext{}, loadedBlock.Signature, SignatureSchemaPrototype)
	require.NoError(t, err)
	loadedSigBlock, err := GetSignatureBlockFromNode(nd)
	require.NoError(t, err)
	require.Equal(t, sigBlock, *loadedSigBlock)
}

func TestVerifyBlockSignature_NoSignature(t *testing.T) {
	lsys, _ := setupTestEnv(t)
	block := makeCompositeBlock(t, lsys)
	storeBlock(t, lsys, block)
	err := VerifyBlockSignature(&block, lsys)
	require.NoError(t, err)
}

func TestVerifyBlockSignature_ValidEd25519(t *testing.T) {
	lsys, keys := setupTestEnv(t)
	block := makeCompositeBlock(t, lsys)
	createSignedBlock(t, lsys, &block, SignatureTypeEd25519, keys)
	storeBlock(t, lsys, block)
	err := VerifyBlockSignature(&block, lsys)
	require.NoError(t, err)
}

func TestVerifyBlockSignature_ValidECDSA(t *testing.T) {
	lsys, keys := setupTestEnv(t)
	block := makeCompositeBlock(t, lsys)
	createSignedBlock(t, lsys, &block, SignatureTypeECDSA256K, keys)
	err := VerifyBlockSignature(&block, lsys)
	require.NoError(t, err)
}

func TestVerifyBlockSignature_InvalidLink(t *testing.T) {
	lsys, _ := setupTestEnv(t)
	block := makeCompositeBlock(t, lsys)
	block.Signature = &cidlink.Link{} // Invalid CID
	err := VerifyBlockSignature(&block, lsys)
	require.ErrorIs(t, err, ErrSignatureNotFound)
}

func TestVerifyBlockSignature_TamperedData(t *testing.T) {
	lsys, keys := setupTestEnv(t)
	block := makeCompositeBlock(t, lsys)
	createSignedBlock(t, lsys, &block, SignatureTypeEd25519, keys)

	// Tamper with the data after signing
	block.Links = append(block.Links, block.Links[0])

	err := VerifyBlockSignature(&block, lsys)
	require.ErrorIs(t, err, crypto.ErrSignatureVerification)
}

func TestVerifyBlockSignature_UnsupportedType(t *testing.T) {
	lsys, _ := setupTestEnv(t)
	block := makeCompositeBlock(t, lsys)

	// Create signature block with unsupported type
	sigBlock := &Signature{
		Header: SignatureHeader{
			Type:     "UnsupportedType",
			Identity: []byte("any"),
		},
		Value: []byte("any"),
	}

	sigBlockLink, err := lsys.Store(ipld.LinkContext{}, GetLinkPrototype(), sigBlock.GenerateNode())
	require.NoError(t, err)
	sigLink, ok := sigBlockLink.(cidlink.Link)
	require.True(t, ok)
	block.Signature = &sigLink

	err = VerifyBlockSignature(&block, lsys)
	require.ErrorIs(t, err, crypto.ErrUnsupportedSignatureType)
}
