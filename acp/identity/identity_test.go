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

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"

	defracrypto "github.com/sourcenetwork/defradb/crypto"
)

func TestGenerate_WithSecp256k1_ReturnsNewRawIdentity(t *testing.T) {
	newIdentity, err := Generate(defracrypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	require.NotEmpty(t, newIdentity.PrivateKey)
	require.NotEmpty(t, newIdentity.PublicKey)
	require.Equal(t, string(defracrypto.KeyTypeSecp256k1), newIdentity.KeyType)

	require.Equal(t, newIdentity.DID[:7], "did:key")

	privKeyBytes, err := hex.DecodeString(newIdentity.PrivateKey)
	require.NoError(t, err)
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	identity, err := FromPrivateKey(defracrypto.NewPrivateKey(privKey))
	require.NoError(t, err)
	require.Equal(t, defracrypto.KeyTypeSecp256k1, identity.PrivateKey.Type())
}

func TestGenerate_WithEd25519_ReturnsNewRawIdentity(t *testing.T) {
	newIdentity, err := Generate(defracrypto.KeyTypeEd25519)
	require.NoError(t, err)

	require.NotEmpty(t, newIdentity.PrivateKey)
	require.NotEmpty(t, newIdentity.PublicKey)
	require.Equal(t, string(defracrypto.KeyTypeEd25519), newIdentity.KeyType)

	require.Equal(t, newIdentity.DID[:7], "did:key")

	privKeyBytes, err := hex.DecodeString(newIdentity.PrivateKey)
	require.NoError(t, err)

	identity, err := FromPrivateKey(defracrypto.NewPrivateKey(ed25519.PrivateKey(privKeyBytes)))
	require.NoError(t, err)
	require.Equal(t, defracrypto.KeyTypeEd25519, identity.PrivateKey.Type())
}

func TestGenerate_WithInvalidType_ReturnsError(t *testing.T) {
	_, err := Generate(defracrypto.KeyType("invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported key type: invalid")
}

func TestGenerate_ReturnsUniqueRawIdentities(t *testing.T) {
	newIdentity1, err1 := Generate(defracrypto.KeyTypeSecp256k1)
	newIdentity2, err2 := Generate(defracrypto.KeyTypeSecp256k1)
	require.NoError(t, err1)
	require.NoError(t, err2)

	// Check that both private and public key are not empty.
	require.NotEmpty(t, newIdentity1.PrivateKey)
	require.NotEmpty(t, newIdentity1.PublicKey)
	require.Equal(t, string(defracrypto.KeyTypeSecp256k1), newIdentity1.KeyType)
	require.NotEmpty(t, newIdentity2.PrivateKey)
	require.NotEmpty(t, newIdentity2.PublicKey)
	require.Equal(t, string(defracrypto.KeyTypeSecp256k1), newIdentity2.KeyType)

	// Check leading `did:key` prefix.
	require.Equal(t, newIdentity1.DID[:7], "did:key")
	require.Equal(t, newIdentity2.DID[:7], "did:key")

	// Check both are different.
	require.NotEqual(t, newIdentity1.PrivateKey, newIdentity2.PrivateKey)
	require.NotEqual(t, newIdentity1.PublicKey, newIdentity2.PublicKey)
	require.NotEqual(t, newIdentity1.DID, newIdentity2.DID)
}

func TestRawIdentity_IntoIdentityWithSecp256k1_Success(t *testing.T) {
	rawIdentity, err := Generate(defracrypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	identity, err := rawIdentity.IntoIdentity()
	require.NoError(t, err)

	require.Equal(t, defracrypto.KeyTypeSecp256k1, identity.PrivateKey.Type())
	require.Equal(t, defracrypto.KeyTypeSecp256k1, identity.PublicKey.Type())
	require.Equal(t, rawIdentity.DID, identity.DID)

	privKeyBytes := identity.PrivateKey.Raw()
	require.Equal(t, rawIdentity.PrivateKey, hex.EncodeToString(privKeyBytes))

	pubKeyBytes := identity.PublicKey.Raw()
	require.Equal(t, rawIdentity.PublicKey, hex.EncodeToString(pubKeyBytes))
}

func TestRawIdentity_IntoIdentityWithEd25519_Success(t *testing.T) {
	rawIdentity, err := Generate(defracrypto.KeyTypeEd25519)
	require.NoError(t, err)

	identity, err := rawIdentity.IntoIdentity()
	require.NoError(t, err)

	require.Equal(t, defracrypto.KeyTypeEd25519, identity.PrivateKey.Type())
	require.Equal(t, defracrypto.KeyTypeEd25519, identity.PublicKey.Type())
	require.Equal(t, rawIdentity.DID, identity.DID)

	privKeyBytes := identity.PrivateKey.Raw()
	require.Equal(t, rawIdentity.PrivateKey, hex.EncodeToString(privKeyBytes))

	pubKeyBytes := identity.PublicKey.Raw()
	require.Equal(t, rawIdentity.PublicKey, hex.EncodeToString(pubKeyBytes))
}

func TestRawIdentity_IntoIdentityWithInvalidKeyType_Error(t *testing.T) {
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

func TestRawIdentity_IntoIdentityWithInvalidPrivateKey_Error(t *testing.T) {
	rawIdentity := RawIdentity{
		PrivateKey: "not-hex",
		PublicKey:  "fedcba9876543210",
		DID:        "did:key:test",
		KeyType:    string(defracrypto.KeyTypeSecp256k1),
	}

	_, err := rawIdentity.IntoIdentity()
	require.Error(t, err)
	require.Contains(t, err.Error(), "encoding/hex")
}
