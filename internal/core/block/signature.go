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
	"github.com/ipld/go-ipld-prime/node/bindnode"
)

// SignatureHeader contains the header of the signature.
type SignatureHeader struct {
	// Type is the type of the signature.
	Type string
	// Params are the parameters of the signature.
	//Params ipld.Node
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
