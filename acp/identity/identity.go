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
	"time"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/immutable"
)

const (
	// AuthorizedAccountClaim is the name of the claim
	// field containing the authorized account.
	//
	// This must be the same as `AuthorizedAccountClaim`
	// defined in github.com/sourcenetwork/sourcehub/x/acp/types
	//
	// The type cannot be directly referenced here due
	// to compilation issues with JS targets.
	AuthorizedAccountClaim = "authorized_account"

	// KeyTypeClaim is the name of the claim field containing
	// the type of key used to sign the token. This is used
	// to determine the appropriate verification algorithm
	// when validating the token signature.
	KeyTypeClaim = "key_type"
)

// None specifies an anonymous actor.
var None = immutable.None[Identity]()

// Identity describes a unique actor with basic identity information.
// This is the base interface that all identity types implement.
type Identity interface {
	// PublicKey returns the actor's public key.
	PublicKey() crypto.PublicKey
	// DID returns the actor's unique identifier.
	//
	// The address is derived from the actor's public key, using the did:key method
	DID() string
	// ToPublicRawIdentity converts an `Identity` into a `PublicRawIdentity`.
	ToPublicRawIdentity() PublicRawIdentity
}

// TokenIdentity describes an identity that has a bearer token.
type TokenIdentity interface {
	Identity
	// BearerToken returns the signed bearer token that represents this identity.
	BearerToken() string
}

// FullIdentity describes a complete identity with both basic identity information,
// access to the private key, and the ability to manage bearer tokens.
type FullIdentity interface {
	TokenIdentity
	// PrivateKey returns the actor's private key.
	PrivateKey() crypto.PrivateKey
	// IntoRawIdentity converts an `Identity` into a `RawIdentity`.
	IntoRawIdentity() (RawIdentity, error)
	// NewToken creates and returns a new `BearerToken`.
	NewToken(
		duration time.Duration,
		audience immutable.Option[string],
		authorizedAccount immutable.Option[string],
	) ([]byte, error)
	// SetBearerToken sets the bearer token for this identity.
	SetBearerToken(token string)
	// UpdateToken updates the `BearerToken` field of the `Identity`.
	UpdateToken(
		duration time.Duration,
		audience immutable.Option[string],
		authorizedAccount immutable.Option[string],
	) error
}
