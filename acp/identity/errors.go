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
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errUnsupportedKeyType      = "unsupported key type"
	errMissingKeyType          = "missing key type in token"
	errInvalidKeyTypeClaimType = "key type claim must be a string"
	errPrivateKeyNotAvailable  = "private key not available"
	errMustBeTokenIdentity     = "identity must be a TokenIdentity"
	errAudienceMismatch        = "auth token audience does not match the request host"
	errMissingAudience         = "auth token is missing the audience claim"
	errInvalidAuthToken        = "invalid auth token"
	errInvalidSubject          = "auth token subject is not a valid public key"
	errTokenExpired            = "auth token has expired"
	errTokenNotYetValid        = "auth token is not yet valid"
)

var (
	// ErrUnsupportedKeyType is returned when attempting to use an unsupported key type.
	ErrUnsupportedKeyType = errors.New(errUnsupportedKeyType)
	// ErrMissingKeyType is returned when a JWT token does not contain the required key_type claim.
	ErrMissingKeyType = errors.New(errMissingKeyType)
	// ErrInvalidKeyTypeClaimType is returned when the key_type claim in a JWT token is not a string.
	ErrInvalidKeyTypeClaimType = errors.New(errInvalidKeyTypeClaimType)
	// ErrPrivateKeyNotAvailable is returned when attempting to use identity with no private key.
	ErrPrivateKeyNotAvailable = errors.New(errPrivateKeyNotAvailable)
	// ErrMustBeTokenIdentity is returned when used identity does not implement TokenIdentity.
	ErrMustBeTokenIdentity = errors.New(errMustBeTokenIdentity)
	// ErrAudienceMismatch is returned when an auth token's audience claim does not match
	// any of the accepted audiences (the request host or an allowed origin).
	// This is an operator-actionable cause.
	ErrAudienceMismatch = errors.New(errAudienceMismatch)
	// ErrMissingAudience is returned when an auth token carries no audience claim.
	// This is an operator-actionable cause.
	ErrMissingAudience = errors.New(errMissingAudience)
	// ErrInvalidAuthToken is returned for any auth token that is malformed, structurally
	// invalid, or whose signature does not verify. Signature failures are deliberately
	// bucketed here so the cause is not distinguishable, to avoid leaking key validity.
	ErrInvalidAuthToken = errors.New(errInvalidAuthToken)
	// ErrInvalidSubject is returned when an auth token's subject claim is not a valid
	// public key for the declared key type. This is an operator-actionable cause.
	ErrInvalidSubject = errors.New(errInvalidSubject)
	// ErrTokenExpired is returned when an auth token's expiry (exp) is in the past.
	// This is an operator-actionable cause.
	ErrTokenExpired = errors.New(errTokenExpired)
	// ErrTokenNotYetValid is returned when an auth token's not-before (nbf) is in the
	// future. This is an operator-actionable cause.
	ErrTokenNotYetValid = errors.New(errTokenNotYetValid)
)
