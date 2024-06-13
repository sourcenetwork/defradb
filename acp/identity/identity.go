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
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
)

// None specifies an anonymous actor.
var None = immutable.None[Identity]()

// Identity describes a unique actor.
type Identity struct {
	// PublicKey is the actor's public key.
	PublicKey *secp256k1.PublicKey
	// PrivateKey is the actor's private key.
	PrivateKey *secp256k1.PrivateKey
	// DID is the actor's unique identifier.
	//
	// The address is derived from the actor's public key,
	// using the did:key method
	DID string
}

// FromPrivateKey returns a new identity using the given private key.
func FromPrivateKey(privateKey *secp256k1.PrivateKey) (immutable.Option[Identity], error) {
	return newIdentityProvider().FromPrivateKey(privateKey)
}

// FromPublicKey returns a new identity using the given public key.
func FromPublicKey(publicKey *secp256k1.PublicKey) (immutable.Option[Identity], error) {
	return newIdentityProvider().FromPublicKey(publicKey)
}

// DIDFromPublicKey returns the unique address of the given public key.
func DIDFromPublicKey(publicKey *secp256k1.PublicKey) (string, error) {
	return generateDID(publicKey, getDefaultDIDProducer())
}
