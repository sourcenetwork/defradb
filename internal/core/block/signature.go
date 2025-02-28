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
	SignatureTypeECDSA   = "ECDSA"
	SignatureTypeEd25519 = "Ed25519"
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
	encBlock, ok := bindnode.Unwrap(node).(*Signature)
	if !ok {
		return nil, NewErrNodeToBlock(node)
	}
	return encBlock, nil
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
// It returns nil if:
// - The block has no signature (optional signature)
// - The signature is valid
// It returns an error if:
// - The signature block cannot be loaded
// - The signature verification fails
func VerifyBlockSignature(block *Block, lsys *linking.LinkSystem) error {
	if block.Signature == nil {
		return nil
	}

	// Load the signature block
	nd, err := lsys.Load(ipld.LinkContext{}, *block.Signature, SignatureSchemaPrototype)
	if err != nil {
		return ErrSignatureNotFound
	}

	sigBlock, err := GetSignatureBlockFromNode(nd)
	if err != nil {
		return ErrSignatureNotFound
	}

	// Generate a new node from the block without the signature field
	// This is what was originally signed
	blockToVerify := *block
	blockToVerify.Signature = nil

	// Marshal the node to get the bytes that were signed
	signedBytes, err := marshalNode(&blockToVerify, BlockSchema)
	if err != nil {
		return err
	}

	var sigType crypto.SignatureType
	switch sigBlock.Header.Type {
	case SignatureTypeEd25519:
		sigType = crypto.SignatureTypeEd25519
	case SignatureTypeECDSA:
		sigType = crypto.SignatureTypeECDSA
	default:
		return crypto.ErrUnsupportedSignatureType
	}

	return crypto.Verify(sigType, sigBlock.Header.Identity, signedBytes, sigBlock.Value)
}
