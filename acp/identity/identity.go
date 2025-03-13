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

	gocrypto "github.com/cyware/ssi-sdk/crypto"
	"github.com/cyware/ssi-sdk/did/key"
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

// didProducer generates a did:key from a public key
type didProducer = func(gocrypto.KeyType, []byte) (*key.DIDKey, error)

// None specifies an anonymous actor.
var None = immutable.None[Identity]()

// BearerTokenSignatureScheme is the signature algorithm used to sign the
// Identity.BearerToken.
const BearerTokenSignatureScheme = jwa.ES256K

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

// DIDFromPublicKey returns a did:key generated from the the given public key.
func DIDFromPublicKey(publicKey *secp256k1.PublicKey) (string, error) {
	return didFromPublicKey(publicKey, key.CreateDIDKey)
}

// didFromPublicKey produces a did from a secp256k1 key and a producer function
func didFromPublicKey(publicKey *secp256k1.PublicKey, producer didProducer) (string, error) {
	bytes := publicKey.SerializeUncompressed()
	did, err := producer(gocrypto.SECP256k1, bytes)
	if err != nil {
		return "", newErrDIDCreation(err, "secp256k1", bytes)
	}
	return did.String(), nil
}

// IntoRawIdentity converts an `Identity` into a `RawIdentity`.
func (identity Identity) IntoRawIdentity() RawIdentity {
	privKeyBytes, _ := identity.PrivateKey.Raw()
	pubKeyBytes, _ := identity.PublicKey.Raw()
	return RawIdentity{
		PrivateKey: hex.EncodeToString(privKeyBytes),
		PublicKey:  hex.EncodeToString(pubKeyBytes),
		DID:        identity.DID,
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
	var signedToken []byte
	pubKeyBytes, err := identity.PublicKey.Raw()
	if err != nil {
		return nil, err
	}
	subject := hex.EncodeToString(pubKeyBytes)
	now := time.Now()

	jwtBuilder := jwt.NewBuilder()
	jwtBuilder = jwtBuilder.Subject(subject)
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

	// For now we only support ECDSA with secp256k1 for bearer tokens
	if identity.PrivateKey.Type() != crypto.KeyTypeSecp256k1 {
		return nil, crypto.ErrUnsupportedSignatureType
	}

	// Convert to ECDSA key for JWT signing
	privKeyBytes, err := identity.PrivateKey.Raw()
	if err != nil {
		return nil, err
	}
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	signedToken, err = jwt.Sign(token, jwt.WithKey(BearerTokenSignatureScheme, privKey.ToECDSA()))
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

	// For now we only support ECDSA with secp256k1 for bearer tokens
	if ident.PublicKey.Type() != crypto.KeyTypeSecp256k1 {
		return crypto.ErrUnsupportedSignatureType
	}

	// Convert to ECDSA key for JWT verification
	pubKeyBytes, err := ident.PublicKey.Raw()
	if err != nil {
		return err
	}
	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		return err
	}

	_, err = jws.Verify(
		[]byte(ident.BearerToken),
		jws.WithKey(BearerTokenSignatureScheme, pubKey.ToECDSA()),
	)
	if err != nil {
		return err
	}

	return nil
}
