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
	"encoding/hex"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/crypto"
)

func TestBuildAuthToken(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity := acp.IdentityFromPrivateKey(privKey)
	token, err := buildAuthToken(identity.Value(), "abc123")
	require.NoError(t, err)

	subject := hex.EncodeToString(privKey.PubKey().SerializeCompressed())
	assert.Equal(t, subject, token.Subject())

	assert.True(t, token.NotBefore().Before(time.Now()))
	assert.True(t, token.Expiration().After(time.Now()))
	assert.Equal(t, []string{"abc123"}, token.Audience())
}

func TestSignAuthTokenErrorsWithPublicIdentity(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity := acp.IdentityFromPublicKey(privKey.PubKey())
	token, err := buildAuthToken(identity.Value(), "abc123")
	require.NoError(t, err)

	_, err = signAuthToken(identity.Value(), token)
	assert.ErrorIs(t, err, ErrMissingIdentityPrivateKey)
}

func TestVerifyAuthToken(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity := acp.IdentityFromPrivateKey(privKey)
	token, err := buildAndSignAuthToken(identity.Value(), "abc123")
	require.NoError(t, err)

	actual, err := verifyAuthToken(token, "abc123")
	require.NoError(t, err)

	expected := acp.IdentityFromPublicKey(privKey.PubKey())
	assert.Equal(t, expected.Value().Address, actual.Value().Address)
}

func TestVerifyAuthTokenErrorsWithNonMatchingAudience(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity := acp.IdentityFromPrivateKey(privKey)
	token, err := buildAndSignAuthToken(identity.Value(), "valid")
	require.NoError(t, err)

	_, err = verifyAuthToken(token, "invalid")
	assert.Error(t, err)
}

func TestVerifyAuthTokenErrorsWithWrongPublicKey(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	otherKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity := acp.IdentityFromPrivateKey(privKey)
	token, err := buildAuthToken(identity.Value(), "123abc")
	require.NoError(t, err)

	// override subject
	subject := hex.EncodeToString(otherKey.PubKey().SerializeCompressed())
	err = token.Set(jwt.SubjectKey, subject)
	require.NoError(t, err)

	data, err := signAuthToken(identity.Value(), token)
	require.NoError(t, err)

	_, err = verifyAuthToken(data, "123abc")
	assert.Error(t, err)
}

func TestVerifyAuthTokenErrorsWithExpired(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity := acp.IdentityFromPrivateKey(privKey)
	token, err := buildAuthToken(identity.Value(), "123abc")
	require.NoError(t, err)

	// override expiration
	err = token.Set(jwt.ExpirationKey, time.Now().Add(-15*time.Minute))
	require.NoError(t, err)

	data, err := signAuthToken(identity.Value(), token)
	require.NoError(t, err)

	_, err = verifyAuthToken(data, "123abc")
	assert.Error(t, err)
}

func TestVerifyAuthTokenErrorsWithNotBefore(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1()
	require.NoError(t, err)

	identity := acp.IdentityFromPrivateKey(privKey)
	token, err := buildAuthToken(identity.Value(), "123abc")
	require.NoError(t, err)

	// override not before
	err = token.Set(jwt.NotBeforeKey, time.Now().Add(15*time.Minute))
	require.NoError(t, err)

	data, err := signAuthToken(identity.Value(), token)
	require.NoError(t, err)

	_, err = verifyAuthToken(data, "123abc")
	assert.Error(t, err)
}
