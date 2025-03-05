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
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
)

// Encryption contains the encryption information for the block's delta.
type Encryption struct {
	// DocID is the ID of the document that is encrypted with the associated encryption key.
	DocID []byte
	// FieldName is the name of the field that is encrypted with the associated encryption key.
	// It is set if encryption is applied to a field instead of the whole doc.
	// It needs to be a pointer so that it can be translated from and to `optional` in the IPLD schema.
	FieldName *string
	// Encryption key.
	Key []byte
}

// IPLDSchemaBytes returns the IPLD schema representation for the encryption block.
//
// This needs to match the [Encryption] struct or [mustSetSchema] will panic on init.
func (enc *Encryption) IPLDSchemaBytes() []byte {
	return []byte(`
		type Encryption struct {
			docID     Bytes
			fieldName optional String
			key       Bytes
		}
	`)
}

// GetFromBytes returns a block from encoded bytes.
func GetEncryptionBlockFromBytes(b []byte) (*Encryption, error) {
	enc := &Encryption{}
	err := enc.Unmarshal(b)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// GetFromNode returns a block from a node.
func GetEncryptionBlockFromNode(node ipld.Node) (*Encryption, error) {
	encBlock, ok := bindnode.Unwrap(node).(*Encryption)
	if !ok {
		return nil, NewErrNodeToBlock(node)
	}
	return encBlock, nil
}

// Marshal encodes the delta using CBOR encoding.
func (enc *Encryption) Marshal() ([]byte, error) {
	return marshalNode(enc, EncryptionSchema)
}

// Unmarshal decodes the delta from CBOR encoding.
func (enc *Encryption) Unmarshal(b []byte) error {
	return unmarshalNode(b, enc, EncryptionSchema)
}

// GenerateNode generates an IPLD node from the encryption block in its representation form.
func (enc *Encryption) GenerateNode() ipld.Node {
	return bindnode.Wrap(enc, EncryptionSchema).Representation()
}
