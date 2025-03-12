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

	"github.com/sourcenetwork/defradb/crypto"
)

// RawIdentity holds the raw bytes that make up an actor's identity.
type RawIdentity struct {
	// PrivateKey is a hex-encoded private key
	PrivateKey string

	// PublicKey is a hex-encoded public key
	PublicKey string

	// DID is `did:key` key generated from the public key address.
	DID string
}

// PublicRawIdentity holds the raw bytes that make up an actor's identity that can be shared publicly.
type PublicRawIdentity struct {
	// PublicKey is a hex-encoded public key
	PublicKey string

	// DID is `did:key` key generated from the public key address.
	DID string
}

func (r RawIdentity) Public() PublicRawIdentity {
	return PublicRawIdentity{
		PublicKey: r.PublicKey,
		DID:       r.DID,
	}
}

// IntoIdentity converts a RawIdentity into an Identity.
func (r RawIdentity) IntoIdentity() (Identity, error) {
	// For now we only support secp256k1 keys
	privateKeyBytes, err := hex.DecodeString(r.PrivateKey)
	if err != nil {
		return Identity{}, err
	}

	privateKey := secp256k1.PrivKeyFromBytes(privateKeyBytes)
	privKey := crypto.NewPrivateKey(privateKey)
	pubKey := privKey.GetPublic()

	return Identity{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		DID:        r.DID,
	}, nil
}
