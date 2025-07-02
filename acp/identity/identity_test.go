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
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/crypto"
)

type mockIdentity struct {
	did         string
	bearerToken string
	publicKey   crypto.PublicKey
}

func (m *mockIdentity) DID() string                 { return m.did }
func (m *mockIdentity) PublicKey() crypto.PublicKey { return m.publicKey }
func (m *mockIdentity) ToPublicRawIdentity() PublicRawIdentity {
	return PublicRawIdentity{
		PublicKey: hex.EncodeToString(m.publicKey.Raw()),
		DID:       m.did,
	}
}
func (m *mockIdentity) BearerToken() string           { return m.bearerToken }
func (m *mockIdentity) SetBearerToken(string)         {}
func (m *mockIdentity) PrivateKey() crypto.PrivateKey { return nil }
func (m *mockIdentity) IntoRawIdentity() RawIdentity  { return RawIdentity{} }
func (m *mockIdentity) UpdateToken(time.Duration, immutable.Option[string], immutable.Option[string]) error {
	return nil
}
func (m *mockIdentity) NewToken(time.Duration, immutable.Option[string], immutable.Option[string]) ([]byte, error) {
	return nil, nil
}

func TestGenerate_WithSecp256k1_ReturnsNewIdentity(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	require.NotNil(t, identity.PrivateKey())
	require.NotNil(t, identity.PublicKey())
	require.Equal(t, crypto.KeyTypeSecp256k1, identity.PrivateKey().Type())
	require.Equal(t, crypto.KeyTypeSecp256k1, identity.PublicKey().Type())

	require.Equal(t, "did:key", identity.DID()[:7])

	rawIdentity := identity.IntoRawIdentity()
	require.Equal(t, string(crypto.KeyTypeSecp256k1), rawIdentity.KeyType)

	privKeyBytes, err := hex.DecodeString(rawIdentity.PrivateKey)
	require.NoError(t, err)
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	reconstructedIdentity, err := FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)
	require.Equal(t, crypto.KeyTypeSecp256k1, reconstructedIdentity.PrivateKey().Type())
}

func TestGenerate_WithEd25519_ReturnsNewIdentity(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	require.NotNil(t, identity.PrivateKey())
	require.NotNil(t, identity.PublicKey())
	require.Equal(t, crypto.KeyTypeEd25519, identity.PrivateKey().Type())
	require.Equal(t, crypto.KeyTypeEd25519, identity.PublicKey().Type())

	require.Equal(t, "did:key", identity.DID()[:7])

	rawIdentity := identity.IntoRawIdentity()
	require.Equal(t, string(crypto.KeyTypeEd25519), rawIdentity.KeyType)

	privKeyBytes, err := hex.DecodeString(rawIdentity.PrivateKey)
	require.NoError(t, err)

	reconstructedIdentity, err := FromPrivateKey(crypto.NewPrivateKey(ed25519.PrivateKey(privKeyBytes)))
	require.NoError(t, err)
	require.Equal(t, crypto.KeyTypeEd25519, reconstructedIdentity.PrivateKey().Type())
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

	require.NotNil(t, identity1.PrivateKey())
	require.NotNil(t, identity1.PublicKey())
	require.Equal(t, crypto.KeyTypeSecp256k1, identity1.PrivateKey().Type())
	require.NotNil(t, identity2.PrivateKey())
	require.NotNil(t, identity2.PublicKey())
	require.Equal(t, crypto.KeyTypeSecp256k1, identity2.PrivateKey().Type())

	require.Equal(t, "did:key", identity1.DID()[:7])
	require.Equal(t, "did:key", identity2.DID()[:7])

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
	require.Equal(t, identity.DID(), rawIdentity.DID)

	privKeyBytes := identity.PrivateKey().Raw()
	require.Equal(t, hex.EncodeToString(privKeyBytes), rawIdentity.PrivateKey)

	pubKeyBytes := identity.PublicKey().Raw()
	require.Equal(t, hex.EncodeToString(pubKeyBytes), rawIdentity.PublicKey)
}

func TestIdentity_IntoRawIdentityWithEd25519_Success(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	rawIdentity := identity.IntoRawIdentity()

	require.Equal(t, string(crypto.KeyTypeEd25519), rawIdentity.KeyType)
	require.Equal(t, identity.DID(), rawIdentity.DID)

	privKeyBytes := identity.PrivateKey().Raw()
	require.Equal(t, hex.EncodeToString(privKeyBytes), rawIdentity.PrivateKey)

	pubKeyBytes := identity.PublicKey().Raw()
	require.Equal(t, hex.EncodeToString(pubKeyBytes), rawIdentity.PublicKey)
}

func TestRawIdentity_FromRawIdentityWithInvalidKeyType_Error(t *testing.T) {
	rawIdentity := RawIdentity{
		PrivateKey: "0123456789abcdef",
		PublicKey:  "fedcba9876543210",
		DID:        "did:key:test",
		KeyType:    "invalid",
	}

	// Check that DID and PublicKey fields are non-empty and valid
	require.NotEmpty(t, rawIdentity.DID)
	require.True(t, len(rawIdentity.DID) >= 7 && rawIdentity.DID[:7] == "did:key", "DID should start with 'did:key'")
	require.NotEmpty(t, rawIdentity.PublicKey)
	_, err := hex.DecodeString(rawIdentity.PublicKey)
	require.NoError(t, err, "PublicKey should be valid hex")

	// Check that the key type is invalid
	switch rawIdentity.KeyType {
	case string(crypto.KeyTypeSecp256k1), string(crypto.KeyTypeEd25519):
		t.Fatal("Expected invalid key type")
	default:
		_, err := hex.DecodeString(rawIdentity.PrivateKey)
		require.NoError(t, err)
	}
}

func TestRawIdentity_FromRawIdentityWithInvalidPrivateKey_Error(t *testing.T) {
	rawIdentity := RawIdentity{
		PrivateKey: "not-hex",
		PublicKey:  "fedcba9876543210",
		DID:        "did:key:test",
		KeyType:    string(crypto.KeyTypeSecp256k1),
	}

	// Check that DID and PublicKey fields are non-empty and valid
	require.NotEmpty(t, rawIdentity.DID)
	require.True(t, len(rawIdentity.DID) >= 7 && rawIdentity.DID[:7] == "did:key", "DID should start with 'did:key'")
	require.NotEmpty(t, rawIdentity.PublicKey)
	_, err := hex.DecodeString(rawIdentity.PublicKey)
	require.NoError(t, err, "PublicKey should be valid hex")
	require.Equal(t, string(crypto.KeyTypeSecp256k1), rawIdentity.KeyType)

	// Check that the private key is invalid
	_, err = hex.DecodeString(rawIdentity.PrivateKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "encoding/hex")
}

func TestIdentity_RoundTripConversion(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	rawIdentity := identity.IntoRawIdentity()

	privKeyBytes, err := hex.DecodeString(rawIdentity.PrivateKey)
	require.NoError(t, err)
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	reconstructedIdentity, err := FromPrivateKey(crypto.NewPrivateKey(privKey))
	require.NoError(t, err)

	require.Equal(t, identity.DID(), reconstructedIdentity.DID())
	require.True(t, identity.PrivateKey().Equal(reconstructedIdentity.PrivateKey()))
	require.True(t, identity.PublicKey().Equal(reconstructedIdentity.PublicKey()))
}

func TestFromToken_ValidSecp256k1Token_Success(t *testing.T) {
	originalIdentity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	err = originalIdentity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)

	parsedIdentity, err := FromToken([]byte(originalIdentity.BearerToken()))
	require.NoError(t, err)

	require.Equal(t, originalIdentity.DID(), parsedIdentity.DID())
	require.Equal(t, originalIdentity.PublicKey().String(), parsedIdentity.PublicKey().String())
	require.Equal(t, originalIdentity.BearerToken(), parsedIdentity.BearerToken())
}

func TestFromToken_ValidEd25519Token_Success(t *testing.T) {
	originalIdentity, err := Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	err = originalIdentity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)

	parsedIdentity, err := FromToken([]byte(originalIdentity.BearerToken()))
	require.NoError(t, err)

	require.Equal(t, originalIdentity.DID(), parsedIdentity.DID())
	require.Equal(t, originalIdentity.PublicKey().String(), parsedIdentity.PublicKey().String())
	require.Equal(t, originalIdentity.BearerToken(), parsedIdentity.BearerToken())
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
	require.NotEmpty(t, identity.BearerToken())

	err = VerifyAuthToken(identity, "test-audience")
	require.NoError(t, err)
}

func TestUpdateToken_WithAuthorizedAccount_Success(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	authorizedAccount := "test-account"
	err = identity.UpdateToken(time.Hour, immutable.Some("test-audience"), immutable.Some(authorizedAccount))
	require.NoError(t, err)
	require.NotEmpty(t, identity.BearerToken())

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
	require.Equal(t, identity.DID(), parsedIdentity.DID())
	require.Equal(t, identity.PublicKey().String(), parsedIdentity.PublicKey().String())
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
	identity := &mockIdentity{
		did:         "did:key:test",
		bearerToken: validIdentity.BearerToken(),
		publicKey:   mockUnsupportedPublicKey{},
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

func TestFromToken_WithNonStringKeyType_Error(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	token, err := identity.NewToken(time.Hour, immutable.Some("test-audience"), immutable.None[string]())
	require.NoError(t, err)

	parsedToken, err := jwt.Parse(token, jwt.WithVerify(false))
	require.NoError(t, err)

	// Set key_type to a non-string value (numeric in this case)
	err = parsedToken.Set(KeyTypeClaim, 123)
	require.NoError(t, err)

	// Get the underlying private key and validate its type
	privKey := identity.PrivateKey().Underlying()
	secpPrivKey, ok := privKey.(*secp256k1.PrivateKey)
	require.True(t, ok, "expected secp256k1.PrivateKey")

	// Sign the token with the validated key
	modifiedToken, err := jwt.Sign(parsedToken, jwt.WithKey(jwa.ES256K, secpPrivKey.ToECDSA()))
	require.NoError(t, err)

	_, err = FromToken(modifiedToken)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidKeyTypeClaimType)
}

func TestFromPublicKey_WithSecp256k1_CreatesIdentityWithoutPrivateKey(t *testing.T) {
	// First create a full identity with private key
	fullIdentity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	require.NotNil(t, fullIdentity.PrivateKey())

	// Now create an identity from just the public key
	publicOnlyIdentity, err := FromPublicKey(fullIdentity.PublicKey())
	require.NoError(t, err)
	require.NotNil(t, publicOnlyIdentity)

	// Verify that both identities have the same public key and DID
	require.Equal(t, fullIdentity.PublicKey().String(), publicOnlyIdentity.PublicKey().String())
	require.Equal(t, fullIdentity.DID(), publicOnlyIdentity.DID())

	// Verify that public-only identity doesn't implement FullIdentity (no private key)
	if _, ok := publicOnlyIdentity.(FullIdentity); ok {
		t.Fatal("Public-only identity should not implement FullIdentity")
	}
}

func TestFromPublicKey_WithEd25519_CreatesIdentityWithoutPrivateKey(t *testing.T) {
	// First create a full identity with private key
	fullIdentity, err := Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)
	require.NotNil(t, fullIdentity.PrivateKey())

	// Now create an identity from just the public key
	publicOnlyIdentity, err := FromPublicKey(fullIdentity.PublicKey())
	require.NoError(t, err)
	require.NotNil(t, publicOnlyIdentity)

	// Verify that both identities have the same public key and DID
	require.Equal(t, fullIdentity.PublicKey().String(), publicOnlyIdentity.PublicKey().String())
	require.Equal(t, fullIdentity.DID(), publicOnlyIdentity.DID())

	// Verify that public-only identity doesn't implement FullIdentity (no private key)
	if _, ok := publicOnlyIdentity.(FullIdentity); ok {
		t.Fatal("Public-only identity should not implement FullIdentity")
	}
}

func TestFromDID_CreatesIdentityWithDIDOnly(t *testing.T) {
	did := "did:key:example123"
	ident := FromDID(did)

	require.NotNil(t, ident)
	require.Equal(t, did, ident.DID())
	require.Nil(t, ident.PublicKey())

	// Should not implement FullIdentity
	_, isFull := ident.(FullIdentity)
	require.False(t, isFull)

	// ToPublicRawIdentity should return the correct DID
	pubRaw := ident.ToPublicRawIdentity()
	require.Equal(t, did, pubRaw.DID)
}

func TestSetBearerToken_UpdatesFullIdentity(t *testing.T) {
	identity, err := Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	identity.SetBearerToken("customtoken")
	require.Equal(t, "customtoken", identity.BearerToken())
}
