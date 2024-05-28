// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

import (
	cosmosSecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
)

// NoIdentity specifies an anonymous actor.
var NoIdentity = immutable.None[Identity]()

type Identity struct {
	// PublicKey is the identity public key.
	PublicKey *secp256k1.PublicKey
	// PrivateKey is the identity private key.
	PrivateKey *secp256k1.PrivateKey
	// Address is the identity address.
	Address string
}

// IdentityFromPublicKey returns a new identity using the given private key.
func IdentityFromPrivateKey(privateKey *secp256k1.PrivateKey) immutable.Option[Identity] {
	pubKey := privateKey.PubKey()
	return immutable.Some(Identity{
		Address:    AddressFromPublicKey(pubKey),
		PublicKey:  pubKey,
		PrivateKey: privateKey,
	})
}

// IdentityFromPublicKey returns a new identity using the given public key.
func IdentityFromPublicKey(publicKey *secp256k1.PublicKey) immutable.Option[Identity] {
	return immutable.Some(Identity{
		Address:   AddressFromPublicKey(publicKey),
		PublicKey: publicKey,
	})
}

// AddressFromPublicKey returns the unique address of the given public key.
func AddressFromPublicKey(publicKey *secp256k1.PublicKey) string {
	pub := cosmosSecp256k1.PubKey{Key: publicKey.SerializeCompressed()}
	// conversion from well known types should never cause a panic
	return types.MustBech32ifyAddressBytes("cosmos", pub.Address().Bytes())
}
