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

// Identity uniquely identifies an acp actor.
type Identity interface {
	// PublicKey returns the public key of the identity.
	PublicKey() *secp256k1.PublicKey
	// Address returns the bech32 address of the identity.
	Address() string
}

var _ (Identity) = (*PublicKeyIdentity)(nil)

// PublicKeyIdentity is an identity with only a public key.
type PublicKeyIdentity struct {
	pubKey *secp256k1.PublicKey
}

// IdentityFromPublicKey returns a new identity using the given public key.
func IdentityFromPublicKey(pubKey *secp256k1.PublicKey) immutable.Option[Identity] {
	identity := PublicKeyIdentity{pubKey}
	return immutable.Some(Identity(identity))
}

func (i PublicKeyIdentity) PublicKey() *secp256k1.PublicKey {
	return i.pubKey
}

func (i PublicKeyIdentity) Address() string {
	pubKey := cosmosSecp256k1.PubKey{Key: i.pubKey.SerializeCompressed()}
	return types.MustBech32ifyAddressBytes("cosmos", pubKey.Address().Bytes())
}

// PrivateKeyIdentity is an identity with both a private and public key.
type PrivateKeyIdentity struct {
	privKey *secp256k1.PrivateKey
}

// IdentityFromPrivateKey returns a new identity using the given private key.
func IdentityFromPrivateKey(privKey *secp256k1.PrivateKey) immutable.Option[Identity] {
	identity := PrivateKeyIdentity{privKey}
	return immutable.Some(Identity(identity))
}

func (i PrivateKeyIdentity) PublicKey() *secp256k1.PublicKey {
	return i.privKey.PubKey()
}

func (i PrivateKeyIdentity) PrivateKey() *secp256k1.PrivateKey {
	return i.privKey
}

func (i PrivateKeyIdentity) Address() string {
	pubKey := cosmosSecp256k1.PubKey{Key: i.privKey.PubKey().SerializeCompressed()}
	return types.MustBech32ifyAddressBytes("cosmos", pubKey.Address().Bytes())
}
