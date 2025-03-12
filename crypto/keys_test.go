// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecp256k1_KeyType(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := wrappedPrivKey.GetPublic()

	assert.Equal(t, KeyTypeSecp256k1, wrappedPrivKey.Type())
	assert.Equal(t, KeyTypeSecp256k1, wrappedPubKey.Type())
}

func TestSecp256k1_RawBytes(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := wrappedPrivKey.GetPublic()

	privBytes, err := wrappedPrivKey.Raw()
	require.NoError(t, err)
	assert.Equal(t, privKey.Serialize(), privBytes)

	pubBytes, err := wrappedPubKey.Raw()
	require.NoError(t, err)
	assert.Equal(t, privKey.PubKey().SerializeCompressed(), pubBytes)
}

func TestSecp256k1_Equals(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)
	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := wrappedPrivKey.GetPublic()

	otherPrivKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)
	otherWrappedPrivKey := NewPrivateKey(otherPrivKey)
	otherWrappedPubKey := otherWrappedPrivKey.GetPublic()

	assert.True(t, wrappedPrivKey.Equals(wrappedPrivKey))
	assert.True(t, wrappedPubKey.Equals(wrappedPubKey))
	assert.False(t, wrappedPrivKey.Equals(otherWrappedPrivKey))
	assert.False(t, wrappedPubKey.Equals(otherWrappedPubKey))
}

func TestSecp256k1_SignAndVerify(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)
	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := wrappedPrivKey.GetPublic()

	message := []byte("test message")
	sig, err := wrappedPrivKey.Sign(message)
	require.NoError(t, err)

	valid, err := wrappedPubKey.Verify(message, sig)
	require.NoError(t, err)
	assert.True(t, valid)

	// Test with wrong message
	valid, err = wrappedPubKey.Verify([]byte("wrong message"), sig)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestSecp256k1_DID(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)
	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := wrappedPrivKey.GetPublic()

	did, err := wrappedPubKey.DID()
	require.NoError(t, err)
	assert.Contains(t, did, "did:key:")
}

func TestEd25519_KeyType(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := NewPublicKey(pubKey)

	assert.Equal(t, KeyTypeEd25519, wrappedPrivKey.Type())
	assert.Equal(t, KeyTypeEd25519, wrappedPubKey.Type())
}

func TestEd25519_RawBytes(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := NewPublicKey(pubKey)

	privBytes, err := wrappedPrivKey.Raw()
	require.NoError(t, err)
	assert.Equal(t, []byte(privKey), privBytes)

	pubBytes, err := wrappedPubKey.Raw()
	require.NoError(t, err)
	assert.Equal(t, []byte(pubKey), pubBytes)
}

func TestEd25519_Equals(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := NewPublicKey(pubKey)

	otherPubKey, otherPrivKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	otherWrappedPrivKey := NewPrivateKey(otherPrivKey)
	otherWrappedPubKey := NewPublicKey(otherPubKey)

	assert.True(t, wrappedPrivKey.Equals(wrappedPrivKey))
	assert.True(t, wrappedPubKey.Equals(wrappedPubKey))
	assert.False(t, wrappedPrivKey.Equals(otherWrappedPrivKey))
	assert.False(t, wrappedPubKey.Equals(otherWrappedPubKey))
}

func TestEd25519_SignAndVerify(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := NewPublicKey(pubKey)

	message := []byte("test message")
	sig, err := wrappedPrivKey.Sign(message)
	require.NoError(t, err)

	valid, err := wrappedPubKey.Verify(message, sig)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = wrappedPubKey.Verify([]byte("wrong message"), sig)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestEd25519_DID(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	wrappedPubKey := NewPublicKey(pubKey)

	did, err := wrappedPubKey.DID()
	require.NoError(t, err)
	assert.Contains(t, did, "did:key:")
}

func TestKeyType_Equality(t *testing.T) {
	secp256k1Key, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)
	wrappedSecp256k1Key := NewPrivateKey(secp256k1Key)

	ed25519Pub, ed25519Priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	wrappedEd25519PrivKey := NewPrivateKey(ed25519Priv)
	wrappedEd25519PubKey := NewPublicKey(ed25519Pub)

	// Different key types should not be equal
	assert.False(t, wrappedSecp256k1Key.Equals(wrappedEd25519PrivKey))
	assert.False(t, wrappedSecp256k1Key.GetPublic().Equals(wrappedEd25519PubKey))
}

func TestSecp256k1_InvalidSignature(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)
	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := wrappedPrivKey.GetPublic()

	message := []byte("test message")

	// Test with invalid signature
	valid, err := wrappedPubKey.Verify(message, []byte("invalid signature"))
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Equal(t, ErrInvalidECDSASignature, err) // This error is still used for parsing failures
}

func TestEd25519_InvalidSignature(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	wrappedPubKey := NewPublicKey(pubKey)

	message := []byte("test message")

	// Test with invalid signature (too short for Ed25519)
	valid, err := wrappedPubKey.Verify(message, []byte("invalid signature"))
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestSecp256k1PrivateKey_NilValidation(t *testing.T) {
	privKey := NewPrivateKey[*secp256k1.PrivateKey](nil)
	assert.Nil(t, privKey, "NewPrivateKey should return nil for nil input")
}

func TestSecp256k1PublicKey_NilValidation(t *testing.T) {
	pubKey := NewPublicKey[*secp256k1.PublicKey](nil)
	assert.Nil(t, pubKey, "NewSecp256k1PublicKey should return nil for nil input")
}

func TestEd25519PrivateKey_NilValidation(t *testing.T) {
	privKey := NewPrivateKey[ed25519.PrivateKey](nil)
	assert.Nil(t, privKey, "NewPrivateKey should return nil for nil input")
}

func TestEd25519PrivateKey_InvalidLengthValidation(t *testing.T) {
	invalidPrivKey := NewPrivateKey[ed25519.PrivateKey](make([]byte, 10))
	assert.Nil(t, invalidPrivKey, "NewPrivateKey should return nil for invalid length key")
}

func TestEd25519PublicKey_NilValidation(t *testing.T) {
	pubKey := NewPublicKey[ed25519.PublicKey](nil)
	assert.Nil(t, pubKey, "NewPublicKey should return nil for nil input")
}

func TestEd25519PublicKey_InvalidLengthValidation(t *testing.T) {
	invalidPubKey := NewPublicKey[ed25519.PublicKey](make([]byte, 10))
	assert.Nil(t, invalidPubKey, "NewPublicKey should return nil for invalid length key")
}

// Test the generic NewPublicKey/NewPrivateKey functions
func TestGenericNewPrivateKey(t *testing.T) {
	// Test with secp256k1
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	genericKey := NewPrivateKey(privKey)
	assert.NotNil(t, genericKey)
	assert.Equal(t, KeyTypeSecp256k1, genericKey.Type())

	// Test with Ed25519
	_, ed25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	genericEd25519Key := NewPrivateKey(ed25519Key)
	assert.NotNil(t, genericEd25519Key)
	assert.Equal(t, KeyTypeEd25519, genericEd25519Key.Type())
}

func TestGenericNewPublicKey(t *testing.T) {
	// Test with secp256k1
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	genericKey := NewPublicKey(privKey.PubKey())
	assert.NotNil(t, genericKey)
	assert.Equal(t, KeyTypeSecp256k1, genericKey.Type())

	// Test with Ed25519
	ed25519Pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	genericEd25519Key := NewPublicKey(ed25519Pub)
	assert.NotNil(t, genericEd25519Key)
	assert.Equal(t, KeyTypeEd25519, genericEd25519Key.Type())
}
