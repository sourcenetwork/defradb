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

	"github.com/cyware/ssi-sdk/crypto"
	"github.com/cyware/ssi-sdk/did/key"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sourcenetwork/immutable"
	acptypes "github.com/sourcenetwork/sourcehub/x/acp/bearer_token"
)

// didProducer generates a did:key from a public key
type didProducer = func(crypto.KeyType, []byte) (*key.DIDKey, error)

// None specifies an anonymous actor.
var None = immutable.None[Identity]()

// BearerTokenSignatureScheme is the signature algorithm used to sign the
// Identity.BearerToken.
const BearerTokenSignatureScheme = jwa.ES256K

// Identity describes a unique actor.
type Identity struct {
	// PublicKey is the actor's public key.
	PublicKey *secp256k1.PublicKey
	// PrivateKey is the actor's private key.
	PrivateKey *secp256k1.PrivateKey
	// DID is the actor's unique identifier.
	//
	// The address is derived from the actor's public key,
	// using the did:key method
	DID string

	// BearerToken is the signed bearer token that represents this identity.
	BearerToken string
}

// FromPrivateKey returns a new identity using the given private key.
//
//   - duration: The [time.Duration] that this identity is valid for.
//   - audience: The audience that this identity is valid for.  This is required
//     by the Defra http client.  For example `github.com/sourcenetwork/defradb`
//   - authorizedAccount: An account that this identity is authorizing to make
//     SourceHub calls on behalf of this actor.  This is currently required when
//     using SourceHub ACP.
//   - skipTokenGeneration: If true, BearerToken will not be set.  This parameter is
//     provided as generating and signing the token is relatively slow, and only required
//     by remote Defra clients (CLI, http), or if using SourceHub ACP.
func FromPrivateKey(
	privateKey *secp256k1.PrivateKey,
	duration time.Duration,
	audience immutable.Option[string],
	authorizedAccount immutable.Option[string],
	skipTokenGeneration bool,
) (Identity, error) {
	publicKey := privateKey.PubKey()
	did, err := DIDFromPublicKey(publicKey)
	if err != nil {
		return Identity{}, err
	}

	identity := Identity{
		DID:        did,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}

	if !skipTokenGeneration {
		err = identity.UpdateToken(duration, audience, authorizedAccount)
		if err != nil {
			return Identity{}, err
		}
	}
	return identity, nil
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

	did, err := DIDFromPublicKey(pubKey)
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		DID:         did,
		PublicKey:   pubKey,
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
	did, err := producer(crypto.SECP256k1, bytes)
	if err != nil {
		return "", newErrDIDCreation(err, "secp256k1", bytes)
	}
	return did.String(), nil
}

// IntoRawIdentity converts an `Identity` into a `RawIdentity`.
func (identity Identity) IntoRawIdentity() RawIdentity {
	return newRawIdentity(identity.PrivateKey, identity.PublicKey, identity.DID)
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
	var signedToken []byte
	subject := hex.EncodeToString(identity.PublicKey.SerializeCompressed())
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
		return err
	}

	if authorizedAccount.HasValue() {
		err = token.Set(acptypes.AuthorizedAccountClaim, authorizedAccount.Value())
		if err != nil {
			return err
		}
	}

	signedToken, err = jwt.Sign(token, jwt.WithKey(BearerTokenSignatureScheme, identity.PrivateKey.ToECDSA()))
	if err != nil {
		return err
	}

	identity.BearerToken = string(signedToken)
	return nil
}
