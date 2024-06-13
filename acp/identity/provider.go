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

// didProducer produces a did:key from public keys
type didProducer func(crypto.KeyType, []byte) (*key.DIDKey, error)

// getDefaultDIDProducer returns the package default didProducer
func getDefaultDIDProducer() didProducer { return key.CreateDIDKey }

// generateDID receives a public key, a didProduce function and returns a did:key string or an error
func generateDID(pubKey *secp256k1.PublicKey, producer didProducer) (string, error) {
	keyType := "secp256k1"
	bytes := pubKey.SerializeUncompressed()
	didKey, err := producer(crypto.SECP256k1, bytes)

	if err != nil {
		return "", NewErrDIDCreation(err, keyType, bytes)
	}

	return didKey.String(), err
}

// identityProvider provides Identity from key material
type identityProvider struct {
	producer didProducer
}

// newIdentityProvider returns an identityProvider which uses the defaultDIDProducer
func newIdentityProvider() *identityProvider {
	return &identityProvider{
		producer: getDefaultDIDProducer(),
	}
}

// FromPublicKey returns a new identity using the given public key.
func (p *identityProvider) FromPublicKey(publicKey *secp256k1.PublicKey) (immutable.Option[Identity], error) {
	did, err := generateDID(publicKey, p.producer)
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
	did, err := generateDID(pubKey, p.producer)
	if err != nil {
		return None, err
	}

	return immutable.Some(Identity{
		DID:        did,
		PublicKey:  pubKey,
		PrivateKey: privateKey,
	}), nil
}
