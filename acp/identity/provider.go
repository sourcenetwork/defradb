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

var _ didProvider = (*defaultDIDProvider)(nil)

// didProvider produces a did:key from public keys
type didProvider interface {
	// DIDFromSecp256k1 returns a did:key from a secp256k1 pub key
	DIDFromSecp256k1(key *secp256k1.PublicKey) (string, error)
}

// defaultDIDProvider implements didProvier
type defaultDIDProvider struct{}

func (p *defaultDIDProvider) DIDFromSecp256k1(pubKey *secp256k1.PublicKey) (string, error) {
	bytes := pubKey.SerializeUncompressed()
	did, err := key.CreateDIDKey(crypto.SECP256k1, bytes)
	if err != nil {
		return "", NewErrDIDCreation(err, "secp256k1", bytes)
	}
	return did.String(), nil
}

// identityProvider wraps a didProvider and constructs Identity from key material
type identityProvider struct {
	didProv didProvider
}

// newIdentityProvider returns an identityProvider which uses the defaultDIDProvider
func newIdentityProvider() *identityProvider {
	return &identityProvider{
		didProv: &defaultDIDProvider{},
	}
}

// FromPublicKey returns a new identity using the given public key.
func (p *identityProvider) FromPublicKey(publicKey *secp256k1.PublicKey) (immutable.Option[Identity], error) {
	did, err := p.didProv.DIDFromSecp256k1(publicKey)
	if err != nil {
		return None, err
	}
	return immutable.Some(Identity{
		DID:       did,
		PublicKey: publicKey,
	}), nil
}

// FromPrivateKey returns a new identity using the given private key.
func (p *identityProvider) FromPrivateKey(privateKey *secp256k1.PrivateKey) (immutable.Option[Identity], error) {
	pubKey := privateKey.PubKey()
	did, err := p.didProv.DIDFromSecp256k1(pubKey)
	if err != nil {
		return None, err
	}

	return immutable.Some(Identity{
		DID:        did,
		PublicKey:  pubKey,
		PrivateKey: privateKey,
	}), nil
}
