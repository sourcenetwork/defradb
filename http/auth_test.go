// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

func TestAuth_WithSecp256k1_ShouldSucceed(t *testing.T) {
	audience := "abc123"

	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)

	err = identity.UpdateToken(time.Hour, immutable.Some(audience), immutable.None[string]())
	require.NoError(t, err, "failed to update token")

	err = acpIdentity.VerifyAuthToken(identity, audience)
	require.NoError(t, err, "failed to verify auth token")
}

func TestAuth_WithSecp256k1AndNonMatchingAudience_ShouldError(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)

	err = identity.UpdateToken(time.Hour, immutable.Some("valid"), immutable.None[string]())
	require.NoError(t, err, "failed to update token")

	err = acpIdentity.VerifyAuthToken(identity, "invalid")
	assert.Error(t, err, "failed to verify auth token")
}

func TestAuth_WithSecp256k1AndExpiredToken_ShouldError(t *testing.T) {
	audience := "abc123"

	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)

	// negative expiration
	err = identity.UpdateToken(-time.Hour, immutable.Some(audience), immutable.None[string]())
	require.NoError(t, err, "failed to update token")

	err = acpIdentity.VerifyAuthToken(identity, audience)
	assert.Error(t, err, "failed to verify auth token")
}

func TestAuth_WithEd25519_ShouldSucceed(t *testing.T) {
	audience := "abc123"

	privKey, err := crypto.GenerateEd25519()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)

	err = identity.UpdateToken(time.Hour, immutable.Some(audience), immutable.None[string]())
	require.NoError(t, err, "failed to update token")

	err = acpIdentity.VerifyAuthToken(identity, audience)
	require.NoError(t, err, "failed to verify auth token")
}

func TestAuth_WithEd25519AndNonMatchingAudience_ShouldError(t *testing.T) {
	privKey, err := crypto.GenerateEd25519()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)

	err = identity.UpdateToken(time.Hour, immutable.Some("valid"), immutable.None[string]())
	require.NoError(t, err, "failed to update token")

	err = acpIdentity.VerifyAuthToken(identity, "invalid")
	assert.Error(t, err, "failed to verify auth token")
}

func TestAuth_WithEd25519AndExpiredToken_ShouldError(t *testing.T) {
	audience := "abc123"

	privKey, err := crypto.GenerateEd25519()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)

	// negative expiration
	err = identity.UpdateToken(-time.Hour, immutable.Some(audience), immutable.None[string]())
	require.NoError(t, err, "failed to update token")

	err = acpIdentity.VerifyAuthToken(identity, audience)
	assert.Error(t, err, "failed to verify auth token")
}
