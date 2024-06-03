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
	cosmosSecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
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
	// Address is the actor's unique address.
	//
	// The address is derived from the actor's public key.
	Address string
}

// FromPrivateKey returns a new identity using the given private key.
func FromPrivateKey(privateKey *secp256k1.PrivateKey) immutable.Option[Identity] {
	pubKey := privateKey.PubKey()
	return immutable.Some(Identity{
		Address:    AddressFromPublicKey(pubKey),
		PublicKey:  pubKey,
		PrivateKey: privateKey,
	})
}

// FromPublicKey returns a new identity using the given public key.
func FromPublicKey(publicKey *secp256k1.PublicKey) immutable.Option[Identity] {
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
