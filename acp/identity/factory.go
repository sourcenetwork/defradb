// Copyright 2025 Democratized Data Foundation
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
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/sourcenetwork/defradb/crypto"
)

// FromDID returns an Identity with only a DID and no public key.
func FromDID(did string) Identity {
	return &baseIdentity{did: did}
}

// FromPrivateKey returns a new full identity using the given private key.
func FromPrivateKey(privateKey crypto.PrivateKey) (FullIdentity, error) {
	publicKey := privateKey.GetPublic()
	did, err := publicKey.DID()
	if err != nil {
		return nil, err
	}

	return &fullIdentity{
		identity: identity{
			publicKey: publicKey,
			did:       did,
		},
		privateKey: privateKey,
	}, nil
}

// FromPublicKey returns a new identity using only the given public key.
func FromPublicKey(publicKey crypto.PublicKey) (Identity, error) {
	did, err := publicKey.DID()
	if err != nil {
		return nil, err
	}

	return &identity{
		publicKey: publicKey,
		did:       did,
	}, nil
}

// FromToken constructs a new identity from a bearer token.
// The returned identity implements FullIdentity but cannot create or update tokens
// since it doesn't have access to the private key.
func FromToken(data []byte) (TokenIdentity, error) {
	token, err := jwt.Parse(data, jwt.WithVerify(false))
	if err != nil {
		return nil, err
	}

	keyTypeStr, ok := token.Get(KeyTypeClaim)
	if !ok {
		return nil, ErrMissingKeyType
	}

	keyTypeValue, ok := keyTypeStr.(string)
	if !ok {
		return nil, ErrInvalidKeyTypeClaimType
	}

	publicKey, err := crypto.PublicKeyFromString(crypto.KeyType(keyTypeValue), token.Subject())
	if err != nil {
		return nil, err
	}

	did, err := publicKey.DID()
	if err != nil {
		return nil, err
	}

	return &fullIdentity{
		identity: identity{
			publicKey: publicKey,
			did:       did,
		},
		bearerToken: string(data),
	}, nil
}

// Generate generates a new full identity with the specified key type.
// Supported types are KeyTypeSecp256k1 and KeyTypeEd25519.
func Generate(keyType crypto.KeyType) (FullIdentity, error) {
	privKey, err := crypto.GenerateKey(keyType)
	if err != nil {
		return nil, err
	}

	identity, err := FromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	return identity, nil
}
