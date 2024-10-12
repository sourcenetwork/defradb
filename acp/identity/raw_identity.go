// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package identity

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// RawIdentity holds the raw bytes that make up an actor's identity.
type RawIdentity struct {
	// PrivateKey is a secp256k1 private key that is a 256-bit big-endian
	// binary-encoded number, padded to a length of 32 bytes in HEX format.
	PrivateKey string

	// PublicKey is a compressed 33-byte secp256k1 public key in HEX format.
	PublicKey string

	// DID is `did:key` key generated from the public key address.
	DID string
}

// PublicRawIdentity holds the raw bytes that make up an actor's identity that can be shared publicly.
type PublicRawIdentity struct {
	// PublicKey is a compressed 33-byte secp256k1 public key in HEX format.
	PublicKey string

	// DID is `did:key` key generated from the public key address.
	DID string
}

func newRawIdentity(privateKey *secp256k1.PrivateKey, publicKey *secp256k1.PublicKey, did string) RawIdentity {
	return RawIdentity{
		PrivateKey: hex.EncodeToString(privateKey.Serialize()),
		PublicKey:  hex.EncodeToString(publicKey.SerializeCompressed()),
		DID:        did,
	}
}

func (r RawIdentity) Public() PublicRawIdentity {
	return PublicRawIdentity{
		PublicKey: r.PublicKey,
		DID:       r.DID,
	}
}
