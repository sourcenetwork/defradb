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
	"encoding/hex"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/immutable"
)

// baseIdentity is a minimal implementation of the Identity interface that only has a DID.
type baseIdentity struct {
	did string
}

// identity is the concrete implementation of the Identity interface with a public key and DID.
type identity struct {
	did       string
	publicKey crypto.PublicKey
}

// fullIdentity is the concrete implementation of the FullIdentity interface, holding both public and private keys and a bearer token.
type fullIdentity struct {
	identity
	bearerToken string
	privateKey  crypto.PrivateKey
}

// Ensure interface implementations
var _ Identity = (*baseIdentity)(nil)
var _ Identity = (*identity)(nil)
var _ FullIdentity = (*fullIdentity)(nil)

// DID returns the decentralized identifier of the base identity.
func (b *baseIdentity) DID() string {
	return b.did
}

// PublicKey returns nil for baseIdentity, as it does not have a public key.
func (b *baseIdentity) PublicKey() crypto.PublicKey {
	return nil
}

// ToPublicRawIdentity returns a PublicRawIdentity with only the DID for baseIdentity.
func (b *baseIdentity) ToPublicRawIdentity() PublicRawIdentity {
	return PublicRawIdentity{DID: b.did}
}

// PublicKey returns the actor's public key for identity.
func (i *identity) PublicKey() crypto.PublicKey {
	return i.publicKey
}

// DID returns the decentralized identifier of the identity.
func (i *identity) DID() string {
	return i.did
}

// ToPublicRawIdentity converts an identity into a `PublicRawIdentity`.
func (i *identity) ToPublicRawIdentity() PublicRawIdentity {
	return PublicRawIdentity{
		PublicKey: hex.EncodeToString(i.publicKey.Raw()),
		DID:       i.did,
	}
}

// BearerToken returns the signed bearer token that represents this full identity.
func (f *fullIdentity) BearerToken() string {
	return f.bearerToken
}

// PrivateKey returns the actor's private key for fullIdentity.
func (p *fullIdentity) PrivateKey() crypto.PrivateKey {
	return p.privateKey
}

// IntoRawIdentity converts a fullIdentity into a RawIdentity struct.
func (p *fullIdentity) IntoRawIdentity() (RawIdentity, error) {
	privKeyBytes := p.privateKey.Raw()
	keyType := string(p.privateKey.Type())
	pubKeyBytes := p.publicKey.Raw()

	return RawIdentity{
		PrivateKey: hex.EncodeToString(privKeyBytes),
		PublicKey:  hex.EncodeToString(pubKeyBytes),
		DID:        p.did,
		KeyType:    keyType,
	}, nil
}

// NewToken creates and returns a new signed bearer token for the fullIdentity.
func (f *fullIdentity) NewToken(
	duration time.Duration,
	audience immutable.Option[string],
	authorizedAccount immutable.Option[string],
) ([]byte, error) {
	if f.privateKey == nil {
		return nil, ErrPrivateKeyNotAvailable
	}

	now := time.Now()

	jwtBuilder := jwt.NewBuilder()
	jwtBuilder = jwtBuilder.Subject(f.publicKey.String())
	jwtBuilder = jwtBuilder.Expiration(now.Add(duration))
	jwtBuilder = jwtBuilder.NotBefore(now)
	jwtBuilder = jwtBuilder.Issuer(f.did)
	jwtBuilder = jwtBuilder.IssuedAt(now)

	if audience.HasValue() {
		jwtBuilder = jwtBuilder.Audience([]string{audience.Value()})
	}

	token, err := jwtBuilder.Build()
	if err != nil {
		return nil, err
	}

	if authorizedAccount.HasValue() {
		err = token.Set(AuthorizedAccountClaim, authorizedAccount.Value())
		if err != nil {
			return nil, err
		}
	}

	err = token.Set(KeyTypeClaim, string(f.privateKey.Type()))
	if err != nil {
		return nil, err
	}

	// For now we only support ECDSA with secp256k1 or Ed25519 for bearer tokens
	if f.privateKey.Type() != crypto.KeyTypeSecp256k1 && f.privateKey.Type() != crypto.KeyTypeEd25519 {
		return nil, crypto.NewErrUnsupportedKeyType(f.privateKey.Type())
	}

	privKey := f.privateKey.Underlying()
	if secpPrivKey, ok := privKey.(*secp256k1.PrivateKey); ok {
		privKey = secpPrivKey.ToECDSA()
	}

	signedToken, err := jwt.Sign(token, jwt.WithKey(keyTypeToJWK(f.privateKey.Type()), privKey))
	if err != nil {
		return nil, err
	}

	return signedToken, nil
}

// SetBearerToken sets the bearerToken to the specified token for fullIdentity.
func (f *fullIdentity) SetBearerToken(token string) {
	f.bearerToken = token
}

// UpdateToken updates the bearerToken field of the fullIdentity by generating a new token.
func (f *fullIdentity) UpdateToken(
	duration time.Duration,
	audience immutable.Option[string],
	authorizedAccount immutable.Option[string],
) error {
	signedToken, err := f.NewToken(duration, audience, authorizedAccount)
	if err != nil {
		return err
	}

	f.bearerToken = string(signedToken)
	return nil
}

// VerifyAuthToken verifies that the jwt auth token is valid and that the signature
// matches the identity of the subject.
func VerifyAuthToken(ident TokenIdentity, audience string) error {
	_, err := jwt.Parse([]byte(ident.BearerToken()), jwt.WithVerify(false), jwt.WithAudience(audience))
	if err != nil {
		return err
	}

	// For now we only support ECDSA with secp256k1 or Ed25519 for bearer tokens
	if ident.PublicKey().Type() != crypto.KeyTypeSecp256k1 && ident.PublicKey().Type() != crypto.KeyTypeEd25519 {
		return crypto.NewErrUnsupportedKeyType(ident.PublicKey().Type())
	}

	pubKey := ident.PublicKey().Underlying()
	if secpPubkey, ok := pubKey.(*secp256k1.PublicKey); ok {
		pubKey = secpPubkey.ToECDSA()
	}

	_, err = jws.Verify([]byte(ident.BearerToken()), jws.WithKey(keyTypeToJWK(ident.PublicKey().Type()), pubKey))
	if err != nil {
		return err
	}

	return nil
}

// keyTypeToJWK maps a crypto.KeyType to the corresponding JWA signature algorithm.
func keyTypeToJWK(keyType crypto.KeyType) jwa.SignatureAlgorithm {
	if keyType == crypto.KeyTypeEd25519 {
		return jwa.EdDSA
	}
	return jwa.ES256K
}
