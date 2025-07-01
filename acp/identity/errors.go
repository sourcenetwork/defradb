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
)
