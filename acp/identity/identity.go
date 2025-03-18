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
	"encoding/hex"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
)

// AuthorizedAccountClaim is the name of the claim
// field containing the authorized account.
//
// This must be the same as `AuthorizedAccountClaim`
// defined in github.com/sourcenetwork/sourcehub/x/acp/types
//
// The type cannot be directly referenced here due
// to compilation issues with JS targets.
const AuthorizedAccountClaim = "authorized_account"

// None specifies an anonymous actor.
var None = immutable.None[Identity]()

// Identity describes a unique actor.
type Identity struct {
	// PublicKey is the actor's public key.
	PublicKey crypto.PublicKey
	// PrivateKey is the actor's private key.
	PrivateKey crypto.PrivateKey
	// DID is the actor's unique identifier.
	//
	// The address is derived from the actor's public key,
	// using the did:key method
	DID string

	// BearerToken is the signed bearer token that represents this identity.
	BearerToken string
}

// FromPrivateKey returns a new identity using the given private key.
// In order to generate a fresh token for this identity, use the [UpdateToken]
func FromPrivateKey(privateKey crypto.PrivateKey) (Identity, error) {
	pubKey := privateKey.GetPublic()
	did, err := pubKey.DID()
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		DID:        did,
		PrivateKey: privateKey,
		PublicKey:  pubKey,
	}, nil
}

// FromToken constructs a new `Identity` from a bearer token.
func FromToken(data []byte) (Identity, error) {
	token, err := jwt.Parse(data, jwt.WithVerify(false))
	if err != nil {
		return Identity{}, err
	}

	subject, err := hex.DecodeString(token.Subject())
	if err != nil {
		return Identity{}, err
	}

	pubKey, err := secp256k1.ParsePubKey(subject)
	if err != nil {
		return Identity{}, err
	}

	publicKey := crypto.NewPublicKey(pubKey)
	did, err := publicKey.DID()
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		DID:         did,
		PublicKey:   publicKey,
		BearerToken: string(data),
	}, nil
}

// IntoRawIdentity converts an `Identity` into a `RawIdentity`.
func (identity Identity) IntoRawIdentity() RawIdentity {
	privKeyBytes := identity.PrivateKey.Raw()
	pubKeyBytes := identity.PublicKey.Raw()
	return RawIdentity{
		PrivateKey: hex.EncodeToString(privKeyBytes),
		PublicKey:  hex.EncodeToString(pubKeyBytes),
		DID:        identity.DID,
		KeyType:    string(identity.PrivateKey.Type()),
	}
}

// UpdateToken updates the `BearerToken` field of the `Identity`.
//
//   - duration: The [time.Duration] that this identity is valid for.
//   - audience: The audience that this identity is valid for.  This is required
//     by the Defra http client.  For example `github.com/sourcenetwork/defradb`
//   - authorizedAccount: An account that this identity is authorizing to make
//     SourceHub calls on behalf of this actor.  This is currently required when
//     using SourceHub ACP.
func (identity *Identity) UpdateToken(
	duration time.Duration,
	audience immutable.Option[string],
	authorizedAccount immutable.Option[string],
) error {
	signedToken, err := identity.NewToken(duration, audience, authorizedAccount)
	if err != nil {
		return err
	}

	identity.BearerToken = string(signedToken)
	return nil
}

// NewToken creates and returns a new `BearerToken`.
//
//   - duration: The [time.Duration] that this identity is valid for.
//   - audience: The audience that this identity is valid for.  This is required
//     by the Defra http client.  For example `github.com/sourcenetwork/defradb`
//   - authorizedAccount: An account that this identity is authorizing to make
//     SourceHub calls on behalf of this actor.  This is currently required when
//     using SourceHub ACP.
func (identity Identity) NewToken(
	duration time.Duration,
	audience immutable.Option[string],
	authorizedAccount immutable.Option[string],
) ([]byte, error) {
	now := time.Now()

	jwtBuilder := jwt.NewBuilder()
	jwtBuilder = jwtBuilder.Subject(identity.PublicKey.String())
	jwtBuilder = jwtBuilder.Expiration(now.Add(duration))
	jwtBuilder = jwtBuilder.NotBefore(now)
	jwtBuilder = jwtBuilder.Issuer(identity.DID)
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

	// For now we only support ECDSA with secp256k1 or Ed25519 for bearer tokens
	if identity.PrivateKey.Type() != crypto.KeyTypeSecp256k1 && identity.PrivateKey.Type() != crypto.KeyTypeEd25519 {
		return nil, crypto.ErrUnsupportedSignatureType
	}

	privKey := identity.PrivateKey.Underlying()
	if secpPrivKey, ok := privKey.(*secp256k1.PrivateKey); ok {
		privKey = secpPrivKey.ToECDSA()
	}

	signedToken, err := jwt.Sign(token, jwt.WithKey(keyTypeToJWK(identity.PrivateKey.Type()), privKey))
	if err != nil {
		return nil, err
	}

	return signedToken, nil
}

// VerifyAuthToken verifies that the jwt auth token is valid and that the signature
// matches the identity of the subject.
func VerifyAuthToken(ident Identity, audience string) error {
	_, err := jwt.Parse([]byte(ident.BearerToken), jwt.WithVerify(false), jwt.WithAudience(audience))
	if err != nil {
		return err
	}

	// For now we only support ECDSA with secp256k1 or Ed25519 for bearer tokens
	if ident.PublicKey.Type() != crypto.KeyTypeSecp256k1 && ident.PublicKey.Type() != crypto.KeyTypeEd25519 {
		return crypto.ErrUnsupportedSignatureType
	}

	pubKey := ident.PublicKey.Underlying()
	if secpPubkey, ok := pubKey.(*secp256k1.PublicKey); ok {
		pubKey = secpPubkey.ToECDSA()
	}

	_, err = jws.Verify([]byte(ident.BearerToken), jws.WithKey(keyTypeToJWK(ident.PublicKey.Type()), pubKey))
	if err != nil {
		return err
	}

	return nil
}

func keyTypeToJWK(keyType crypto.KeyType) jwa.SignatureAlgorithm {
	if keyType == crypto.KeyTypeEd25519 {
		return jwa.EdDSA
	}
	return jwa.ES256K
}
