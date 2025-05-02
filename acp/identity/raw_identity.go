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
	"crypto/ed25519"
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/sourcenetwork/defradb/crypto"
)

// RawIdentity represents an identity in a format suitable for serialization.
type RawIdentity struct {
	// PrivateKey is the actor's private key in HEX format.
	PrivateKey string
	// PublicKey is the actor's public key in HEX format.
	PublicKey string
	// DID is the actor's unique identifier.
	//
	// The address is derived from the actor's public key,
	// using the did:key method
	DID string
	// KeyType is the type of the key
	//
	// Supported values are:
	// - "secp256k1"
	// - "ed25519"
	KeyType string
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
	privateKeyBytes, err := hex.DecodeString(r.PrivateKey)
	if err != nil {
		return Identity{}, err
	}

	var privKey crypto.PrivateKey
	switch r.KeyType {
	case string(crypto.KeyTypeSecp256k1):
		key := secp256k1.PrivKeyFromBytes(privateKeyBytes)
		privKey = crypto.NewPrivateKey(key)
	case string(crypto.KeyTypeEd25519):
		privKey = crypto.NewPrivateKey(ed25519.PrivateKey(privateKeyBytes))
	default:
		return Identity{}, newErrUnsupportedKeyType(r.KeyType)
	}

	pubKey := privKey.GetPublic()

	return Identity{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		DID:        r.DID,
	}, nil
}
