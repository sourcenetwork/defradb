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

import "github.com/sourcenetwork/defradb/crypto"

// RawIdentity holds the raw bytes that make up an actor's identity.
type RawIdentity struct {
	// An actor's private key.
	PrivateKey []byte

	// An actor's corresponding public key address.
	PublicKey []byte

	// An actor's DID. Generated from the public key address.
	DID string
}

// Generate generates a new identity.
func Generate() (RawIdentity, error) {
	privateKey, err := crypto.GenerateSecp256k1()
	if err != nil {
		return RawIdentity{}, err
	}

	maybeNewIdentity, err := FromPrivateKey(privateKey)
	if err != nil {
		return RawIdentity{}, err
	}

	if !maybeNewIdentity.HasValue() {
		return RawIdentity{}, ErrFailedToGenerateIdentityFromPrivateKey
	}

	newIdentity := maybeNewIdentity.Value()

	return RawIdentity{
		PrivateKey: newIdentity.PrivateKey.Serialize(),
		PublicKey:  newIdentity.PublicKey.SerializeUncompressed(),
		DID:        newIdentity.DID,
	}, nil
}
