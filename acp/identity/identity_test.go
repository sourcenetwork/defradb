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
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/crypto"
)

func TestGenerate_WithSecp256k1_ReturnsNewIdentity(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	require.NotNil(t, identity.PrivateKey)
	require.NotNil(t, identity.PublicKey)
	require.Equal(t, crypto.KeyTypeSecp256k1, identity.PrivateKey.Type())
	require.Equal(t, crypto.KeyTypeSecp256k1, identity.PublicKey.Type())

	require.Equal(t, "did:key", identity.DID[:7])

	rawIdentity := identity.IntoRawIdentity()
	require.Equal(t, string(crypto.KeyTypeSecp256k1), rawIdentity.KeyType)

	privKeyBytes, err := hex.DecodeString(rawIdentity.PrivateKey)
	require.NoError(t, err)
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	reconstructedIdentity, err := FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)
	require.Equal(t, crypto.KeyTypeSecp256k1, reconstructedIdentity.PrivateKey.Type())
}

func TestGenerate_WithEd25519_ReturnsNewIdentity(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	require.NotNil(t, identity.PrivateKey)
	require.NotNil(t, identity.PublicKey)
	require.Equal(t, crypto.KeyTypeEd25519, identity.PrivateKey.Type())
	require.Equal(t, crypto.KeyTypeEd25519, identity.PublicKey.Type())

	require.Equal(t, "did:key", identity.DID[:7])

	rawIdentity := identity.IntoRawIdentity()
	require.Equal(t, string(crypto.KeyTypeEd25519), rawIdentity.KeyType)

	privKeyBytes, err := hex.DecodeString(rawIdentity.PrivateKey)
	require.NoError(t, err)

	reconstructedIdentity, err := FromPrivateKey(crypto.NewPrivateKey(ed25519.PrivateKey(privKeyBytes)))
	require.NoError(t, err)
	require.Equal(t, crypto.KeyTypeEd25519, reconstructedIdentity.PrivateKey.Type())
}

func TestGenerate_WithInvalidType_ReturnsError(t *testing.T) {
	_, err := Generate(crypto.KeyType("invalid_key_type"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported key type")
	require.Contains(t, err.Error(), "invalid_key_type")
}

func TestGenerate_ReturnsUniqueIdentities(t *testing.T) {
	identity1, err1 := Generate(crypto.KeyTypeSecp256k1)
	identity2, err2 := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err1)
	require.NoError(t, err2)

	require.NotNil(t, identity1.PrivateKey)
	require.NotNil(t, identity1.PublicKey)
	require.Equal(t, crypto.KeyTypeSecp256k1, identity1.PrivateKey.Type())
	require.NotNil(t, identity2.PrivateKey)
	require.NotNil(t, identity2.PublicKey)
	require.Equal(t, crypto.KeyTypeSecp256k1, identity2.PrivateKey.Type())

	require.Equal(t, "did:key", identity1.DID[:7])
	require.Equal(t, "did:key", identity2.DID[:7])

	raw1 := identity1.IntoRawIdentity()
	raw2 := identity2.IntoRawIdentity()

	require.NotEqual(t, raw1.PrivateKey, raw2.PrivateKey)
	require.NotEqual(t, raw1.PublicKey, raw2.PublicKey)
	require.NotEqual(t, raw1.DID, raw2.DID)
}

func TestIdentity_IntoRawIdentityWithSecp256k1_Success(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	rawIdentity := identity.IntoRawIdentity()

	require.Equal(t, string(crypto.KeyTypeSecp256k1), rawIdentity.KeyType)
	require.Equal(t, identity.DID, rawIdentity.DID)

	privKeyBytes := identity.PrivateKey.Raw()
	require.Equal(t, hex.EncodeToString(privKeyBytes), rawIdentity.PrivateKey)

	pubKeyBytes := identity.PublicKey.Raw()
	require.Equal(t, hex.EncodeToString(pubKeyBytes), rawIdentity.PublicKey)
}

func TestIdentity_IntoRawIdentityWithEd25519_Success(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	rawIdentity := identity.IntoRawIdentity()

	require.Equal(t, string(crypto.KeyTypeEd25519), rawIdentity.KeyType)
	require.Equal(t, identity.DID, rawIdentity.DID)

	privKeyBytes := identity.PrivateKey.Raw()
	require.Equal(t, hex.EncodeToString(privKeyBytes), rawIdentity.PrivateKey)

	pubKeyBytes := identity.PublicKey.Raw()
	require.Equal(t, hex.EncodeToString(pubKeyBytes), rawIdentity.PublicKey)
}

func TestRawIdentity_FromRawIdentityWithInvalidKeyType_Error(t *testing.T) {
	rawIdentity := RawIdentity{
		PrivateKey: "0123456789abcdef",
		PublicKey:  "fedcba9876543210",
		DID:        "did:key:test",
		KeyType:    "invalid",
	}

	_, err := rawIdentity.IntoIdentity()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedKeyType)
}

func TestRawIdentity_FromRawIdentityWithInvalidPrivateKey_Error(t *testing.T) {
	rawIdentity := RawIdentity{
		PrivateKey: "not-hex",
		PublicKey:  "fedcba9876543210",
		DID:        "did:key:test",
		KeyType:    string(crypto.KeyTypeSecp256k1),
	}

	_, err := rawIdentity.IntoIdentity()
	require.Error(t, err)
	require.Contains(t, err.Error(), "encoding/hex")
}

func TestIdentity_RoundTripConversion(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	rawIdentity := identity.IntoRawIdentity()

	reconstructedIdentity, err := rawIdentity.IntoIdentity()
	require.NoError(t, err)

	require.Equal(t, identity.DID, reconstructedIdentity.DID)
	require.True(t, identity.PrivateKey.Equal(reconstructedIdentity.PrivateKey))
	require.True(t, identity.PublicKey.Equal(reconstructedIdentity.PublicKey))
}

func TestFromToken_ValidSecp256k1Token_Success(t *testing.T) {
	originalIdentity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	err = originalIdentity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)

	parsedIdentity, err := FromToken([]byte(originalIdentity.BearerToken))
	require.NoError(t, err)

	require.Equal(t, originalIdentity.DID, parsedIdentity.DID)
	require.Equal(t, originalIdentity.PublicKey.String(), parsedIdentity.PublicKey.String())
	require.Equal(t, originalIdentity.BearerToken, parsedIdentity.BearerToken)
}

func TestFromToken_ValidEd25519Token_Success(t *testing.T) {
	originalIdentity, err := Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	err = originalIdentity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)

	parsedIdentity, err := FromToken([]byte(originalIdentity.BearerToken))
	require.NoError(t, err)

	require.Equal(t, originalIdentity.DID, parsedIdentity.DID)
	require.Equal(t, originalIdentity.PublicKey.String(), parsedIdentity.PublicKey.String())
	require.Equal(t, originalIdentity.BearerToken, parsedIdentity.BearerToken)
}

func TestFromToken_InvalidToken_Error(t *testing.T) {
	_, err := FromToken([]byte("invalid-token"))
	require.Error(t, err)
}

func TestUpdateToken_WithValidDuration_Success(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	err = identity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)
	require.NotEmpty(t, identity.BearerToken)

	err = VerifyAuthToken(identity, "test-audience")
	require.NoError(t, err)
}

func TestUpdateToken_WithAuthorizedAccount_Success(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	authorizedAccount := "test-account"
	err = identity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.Some(authorizedAccount))
	require.NoError(t, err)
	require.NotEmpty(t, identity.BearerToken)

	err = VerifyAuthToken(identity, "test-audience")
	require.NoError(t, err)
}

func TestNewToken_WithValidParameters_Success(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	token, err := identity.NewToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsedIdentity, err := FromToken(token)
	require.NoError(t, err)
	require.Equal(t, identity.DID, parsedIdentity.DID)
	require.Equal(t, identity.PublicKey.String(), parsedIdentity.PublicKey.String())
}

func TestVerifyAuthToken_WithExpiredToken_Error(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	err = identity.UpdateToken(-time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)

	err = VerifyAuthToken(identity, "test-audience")
	require.Error(t, err)
}

func TestVerifyAuthToken_WithWrongAudience_Error(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	err = identity.UpdateToken(time.Hour, immutable.Some("correct-audience"), immutable.None[string]())
	require.NoError(t, err)

	err = VerifyAuthToken(identity, "wrong-audience")
	require.Error(t, err)
}

func TestVerifyAuthToken_WithNoToken_Error(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	err = VerifyAuthToken(identity, "test-audience")
	require.Error(t, err)
}

func TestVerifyAuthToken_WithUnsupportedKeyType_Error(t *testing.T) {
	// First create a valid token using a supported key type
	validIdentity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	err = validIdentity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)

	// Create an identity with an unsupported key type but using the valid token
	identity := Identity{
		DID:         "did:key:test",
		BearerToken: validIdentity.BearerToken,
		PublicKey:   mockUnsupportedPublicKey{},
	}

	err = VerifyAuthToken(identity, "test-audience")
	require.Error(t, err)
	require.ErrorIs(t, err, crypto.NewErrUnsupportedKeyType("unsupported"))
}

// mockUnsupportedPublicKey implements crypto.PublicKey with an unsupported key type
type mockUnsupportedPublicKey struct{}

func (m mockUnsupportedPublicKey) Equal(other crypto.Key) bool         { return false }
func (m mockUnsupportedPublicKey) Raw() []byte                         { return nil }
func (m mockUnsupportedPublicKey) String() string                      { return "" }
func (m mockUnsupportedPublicKey) Type() crypto.KeyType                { return "unsupported" }
func (m mockUnsupportedPublicKey) Verify([]byte, []byte) (bool, error) { return false, nil }
func (m mockUnsupportedPublicKey) DID() (string, error)                { return "", nil }
func (m mockUnsupportedPublicKey) Underlying() any                     { return nil }
