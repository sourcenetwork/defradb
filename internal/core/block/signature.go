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

// VerifyBlockSignature verifies the signature of a block.
// The first return value is true if it actually ran cryptographic verification.
func VerifyBlockSignature(block *Block, lsys *linking.LinkSystem) (bool, error) {
	if block.Signature == nil {
		return false, nil
	}

	nd, err := lsys.Load(ipld.LinkContext{}, *block.Signature, SignatureSchemaPrototype)
	if err != nil {
		return false, NewErrCouldNotLoadSignatureBlock(err)
	}

	sigBlock, err := GetSignatureBlockFromNode(nd)
	if err != nil {
		return false, NewErrCouldNotLoadSignatureBlock(err)
	}

	blockToVerify := *block
	blockToVerify.Signature = nil

	signedBytes, err := marshalNode(&blockToVerify, BlockSchema)
	if err != nil {
		return false, err
	}

	// Convert the hex-encoded public key to a crypto.PublicKey using the new function
	var keyType crypto.KeyType
	switch sigBlock.Header.Type {
	case SignatureTypeEd25519:
		keyType = crypto.KeyTypeEd25519
	case SignatureTypeECDSA256K:
		keyType = crypto.KeyTypeSecp256k1
	default:
		return false, crypto.ErrUnsupportedPrivKeyType
	}

	pubKey, err := crypto.PublicKeyFromString(keyType, string(sigBlock.Header.Identity))
	if err != nil {
		return false, err
	}

	valid, err := pubKey.Verify(signedBytes, sigBlock.Value)

	if err != nil {
		// We return true for 'verified' because we did run cryptographic verification
		return true, err
	}

	if !valid {
		return true, crypto.ErrSignatureVerification
	}

	return true, nil
}
