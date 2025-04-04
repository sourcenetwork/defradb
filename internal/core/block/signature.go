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
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	"github.com/ipld/go-ipld-prime/node/bindnode"

	"github.com/sourcenetwork/defradb/crypto"
)

const (
	SignatureTypeECDSA256K = "ES256K"
	SignatureTypeEd25519   = "EdDSA"
)

// SignatureHeader contains the header of the signature.
type SignatureHeader struct {
	// Type is the type of the signature.
	Type string
	// Identity is the identity of the signer.
	Identity []byte
}

// Signature contains the block's signature.
type Signature struct {
	// Header is the header of the signature.
	Header SignatureHeader
	// Value is the signature value.
	Value []byte
}

// IPLDSchemaBytes returns the IPLD schema representation for the signature header block.
//
// This needs to match the [SignatureHeader] struct or [mustSetSchema] will panic on init.
func (sig *SignatureHeader) IPLDSchemaBytes() []byte {
	return []byte(`
		type SignatureHeader struct {
			type     String
			identity Bytes
		}
	`)
}

// IPLDSchemaBytes returns the IPLD schema representation for the signature block.
//
// This needs to match the [Signature] struct or [mustSetSchema] will panic on init.
func (sig *Signature) IPLDSchemaBytes() []byte {
	return []byte(`
		type Signature struct {
			header SignatureHeader
			value  Bytes
		}
	`)
}

// GetFromBytes returns a block from encoded bytes.
func GetSignatureBlockFromBytes(b []byte) (*Signature, error) {
	enc := &Signature{}
	err := enc.Unmarshal(b)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// GetFromNode returns a block from a node.
func GetSignatureBlockFromNode(node ipld.Node) (*Signature, error) {
	sigBlock, ok := bindnode.Unwrap(node).(*Signature)
	if !ok {
		return nil, NewErrNodeToBlock(node)
	}
	return sigBlock, nil
}

// Marshal encodes the delta using CBOR encoding.
func (sig *Signature) Marshal() ([]byte, error) {
	return marshalNode(sig, SignatureSchema)
}

// Unmarshal decodes the delta from CBOR encoding.
func (sig *Signature) Unmarshal(b []byte) error {
	return unmarshalNode(b, sig, SignatureSchema)
}

// GenerateNode generates an IPLD node from the encryption block in its representation form.
func (sig *Signature) GenerateNode() ipld.Node {
	return bindnode.Wrap(sig, SignatureSchema).Representation()
}

// verifySignature performs the cryptographic verification and returns appropriate results
func verifySignature(pubKey crypto.PublicKey, signedBytes, sigValue []byte) error {
	valid, err := pubKey.Verify(signedBytes, sigValue)
	if err != nil {
		return err
	}
	if !valid {
		return crypto.ErrSignatureVerification
	}
	return nil
}

// VerifyBlockSignature verifies the signature of a block using the link system.
// The first return value is true if it actually ran cryptographic verification.
func VerifyBlockSignature(block *Block, lsys *linking.LinkSystem) (bool, error) {
	if block.Signature == nil {
		return false, nil
	}

	signedBytes, sigBlock, err := getSignedDataAndSignature(block, lsys)
	if err != nil {
		return false, err
	}

	pubKey, err := getPublicKeyFromSignature(sigBlock)
	if err != nil {
		return false, err
	}

	return true, verifySignature(pubKey, signedBytes, sigBlock.Value)
}

// VerifyBlockSignatureWithKey verifies the signature of a block using a public key.
// The first return value is true if it actually ran cryptographic verification.
func VerifyBlockSignatureWithKey(block *Block, lsys *linking.LinkSystem, pubKey crypto.PublicKey) (bool, error) {
	if block.Signature == nil {
		return false, nil
	}

	signedBytes, sigBlock, err := getSignedDataAndSignature(block, lsys)
	if err != nil {
		return false, err
	}

	// Verify that the identity matches the signature's identity
	if string(sigBlock.Header.Identity) != pubKey.String() {
		return false, ErrSignaturePubKeyMismatch
	}

	return true, verifySignature(pubKey, signedBytes, sigBlock.Value)
}

func getSignedDataAndSignature(block *Block, lsys *linking.LinkSystem) ([]byte, *Signature, error) {
	signedBytes, err := getBlockBytesToSign(block)
	if err != nil {
		return nil, nil, err
	}

	nd, err := lsys.Load(ipld.LinkContext{}, *block.Signature, SignatureSchemaPrototype)
	if err != nil {
		return nil, nil, NewErrCouldNotLoadSignatureBlock(err)
	}

	sigBlock, err := GetSignatureBlockFromNode(nd)
	if err != nil {
		return nil, nil, NewErrCouldNotLoadSignatureBlock(err)
	}
	return signedBytes, sigBlock, nil
}

// getBlockBytesToSign returns the bytes to sign for a block
func getBlockBytesToSign(block *Block) ([]byte, error) {
	blockToVerify := *block
	blockToVerify.Signature = nil

	signedBytes, err := marshalNode(&blockToVerify, BlockSchema)
	if err != nil {
		return nil, err
	}

	return signedBytes, nil
}

// getPublicKeyFromSignature extracts the public key from a signature block
func getPublicKeyFromSignature(sigBlock *Signature) (crypto.PublicKey, error) {
	var keyType crypto.KeyType
	switch sigBlock.Header.Type {
	case SignatureTypeEd25519:
		keyType = crypto.KeyTypeEd25519
	case SignatureTypeECDSA256K:
		keyType = crypto.KeyTypeSecp256k1
	default:
		return nil, crypto.ErrUnsupportedPrivKeyType
	}

	return crypto.PublicKeyFromString(keyType, string(sigBlock.Header.Identity))
}
