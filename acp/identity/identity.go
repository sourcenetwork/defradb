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
	"github.com/cyware/ssi-sdk/crypto"
	"github.com/cyware/ssi-sdk/did/key"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
)

// didProducer is a concrete function which
// produces a did from a given key type and pub key bytes
//
// Currently the did production operation is
// infalliable, but in order to assure the correct error
// is being returned, this pkg private variable
// can be used in tests to locally mock the producer function
var didProducer = key.CreateDIDKey

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
	pubKey := privateKey.PubKey()
	did, err := DIDFromPublicKey(pubKey)
	if err != nil {
		return None, err
	}

	return immutable.Some(Identity{
		DID:        did,
		PublicKey:  pubKey,
		PrivateKey: privateKey,
	}), nil
}

// FromPublicKey returns a new identity using the given public key.
func FromPublicKey(publicKey *secp256k1.PublicKey) (immutable.Option[Identity], error) {
	did, err := DIDFromPublicKey(publicKey)
	if err != nil {
		return None, err
	}
	return immutable.Some(Identity{
		DID:       did,
		PublicKey: publicKey,
	}), nil
}

// DIDFromPublicKey returns the unique address of the given public key.
func DIDFromPublicKey(publicKey *secp256k1.PublicKey) (string, error) {
	bytes := publicKey.SerializeUncompressed()
	did, err := didProducer(crypto.SECP256k1, bytes)
	if err != nil {
		return "", NewErrDIDCreation(err, "secp256k1", bytes)
	}
	return did.String(), nil
}
