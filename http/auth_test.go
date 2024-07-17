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

func TestVerifyAuthToken(t *testing.T) {
	audience := "abc123"

	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(
		privKey,
		time.Hour,
		immutable.Some(audience),
		immutable.None[string](),
		false,
	)
	require.NoError(t, err)

	err = verifyAuthToken(identity, audience)
	require.NoError(t, err)
}

func TestVerifyAuthTokenErrorsWithNonMatchingAudience(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(
		privKey,
		time.Hour,
		immutable.Some("valid"),
		immutable.None[string](),
		false,
	)
	require.NoError(t, err)

	err = verifyAuthToken(identity, "invalid")
	assert.Error(t, err)
}

func TestVerifyAuthTokenErrorsWithExpired(t *testing.T) {
	audience := "abc123"

	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity, err := acpIdentity.FromPrivateKey(
		privKey,
		// negative expiration
		-time.Hour,
		immutable.Some(audience),
		immutable.None[string](),
		false,
	)
	require.NoError(t, err)

	err = verifyAuthToken(identity, "123abc")
	assert.Error(t, err)
}
