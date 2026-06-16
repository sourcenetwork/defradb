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
	"slices"
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

// fullIdentity is the concrete implementation of the FullIdentity interface,
// holding both public and private keys and a bearer token.
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
func (f *fullIdentity) PrivateKey() crypto.PrivateKey {
	return f.privateKey
}

// IntoRawIdentity converts a fullIdentity into a RawIdentity struct.
func (f *fullIdentity) IntoRawIdentity() RawIdentity {
	privKeyBytes := f.privateKey.Raw()
	keyType := string(f.privateKey.Type())
	pubKeyBytes := f.publicKey.Raw()

	return RawIdentity{
		PrivateKey: hex.EncodeToString(privKeyBytes),
		PublicKey:  hex.EncodeToString(pubKeyBytes),
		DID:        f.did,
		KeyType:    keyType,
	}
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

// VerifyAuthToken verifies the jwt auth token's signature and that its audience
// claim matches any one of the given audiences. At least one must be provided.
func VerifyAuthToken(ident TokenIdentity, audiences ...string) error {
	// Validate temporal claims (expiry, not-before); audience is checked below
	// since jwx's WithAudience only accepts a single value.
	token, err := jwt.Parse([]byte(ident.BearerToken()), jwt.WithVerify(false), jwt.WithValidate(true))
	if err != nil {
		return err
	}

	if !audienceMatches(token.Audience(), audiences) {
		return ErrTokenAudienceMismatch
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

// audienceMatches reports whether the token's audience claim contains at least
// one of the accepted audiences.
func audienceMatches(tokenAudience []string, accepted []string) bool {
	for _, aud := range accepted {
		if slices.Contains(tokenAudience, aud) {
			return true
		}
	}
	return false
}

// keyTypeToJWK maps a crypto.KeyType to the corresponding JWA signature algorithm.
func keyTypeToJWK(keyType crypto.KeyType) jwa.SignatureAlgorithm {
	if keyType == crypto.KeyTypeEd25519 {
		return jwa.EdDSA
	}
	return jwa.ES256K
}

// cloneFullIdentity creates a deep copy of the given fullIdentity.
func cloneFullIdentity(orig *fullIdentity) *fullIdentity {
	if orig == nil {
		return nil
	}
	return &fullIdentity{
		identity: identity{
			did:       orig.did,
			publicKey: orig.publicKey,
		},
		bearerToken: orig.bearerToken,
		privateKey:  orig.privateKey,
	}
}

// CloneIdentity creates a deep copy of the given identity.
// This exists so as to allow parallel test actions to avoid race conditions.
// Specifically, the MustGetCanonicallyOrderedCollections function requires this functionality
func CloneIdentity(orig Identity) Identity {
	if f, ok := orig.(*fullIdentity); ok {
		return cloneFullIdentity(f)
	}
	return orig
}
