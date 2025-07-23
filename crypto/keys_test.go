// Copyright 2025 Democratized Data Foundation
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
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

	privBytes := wrappedPrivKey.Raw()
	assert.Equal(t, privKey.Serialize(), privBytes)

	pubBytes := wrappedPubKey.Raw()
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

	assert.True(t, wrappedPrivKey.Equal(wrappedPrivKey))
	assert.True(t, wrappedPubKey.Equal(wrappedPubKey))
	assert.False(t, wrappedPrivKey.Equal(otherWrappedPrivKey))
	assert.False(t, wrappedPubKey.Equal(otherWrappedPubKey))
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

	privBytes := wrappedPrivKey.Raw()
	assert.Equal(t, []byte(privKey), privBytes)

	pubBytes := wrappedPubKey.Raw()
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

	assert.True(t, wrappedPrivKey.Equal(wrappedPrivKey))
	assert.True(t, wrappedPubKey.Equal(wrappedPubKey))
	assert.False(t, wrappedPrivKey.Equal(otherWrappedPrivKey))
	assert.False(t, wrappedPubKey.Equal(otherWrappedPubKey))
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
	assert.False(t, wrappedSecp256k1Key.Equal(wrappedEd25519PrivKey))
	assert.False(t, wrappedSecp256k1Key.GetPublic().Equal(wrappedEd25519PubKey))
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
	assert.Equal(t, ErrInvalidECDSASignature, err)
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
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	genericKey := NewPrivateKey(privKey)
	assert.NotNil(t, genericKey)
	assert.Equal(t, KeyTypeSecp256k1, genericKey.Type())

	_, ed25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	genericEd25519Key := NewPrivateKey(ed25519Key)
	assert.NotNil(t, genericEd25519Key)
	assert.Equal(t, KeyTypeEd25519, genericEd25519Key.Type())
}

func TestGenericNewPublicKey(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	genericKey := NewPublicKey(privKey.PubKey())
	assert.NotNil(t, genericKey)
	assert.Equal(t, KeyTypeSecp256k1, genericKey.Type())

	ed25519Pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	genericEd25519Key := NewPublicKey(ed25519Pub)
	assert.NotNil(t, genericEd25519Key)
	assert.Equal(t, KeyTypeEd25519, genericEd25519Key.Type())
}

func TestSecp256k1_Underlying(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := wrappedPrivKey.GetPublic()

	underlying := wrappedPrivKey.Underlying()
	assert.NotNil(t, underlying)
	assert.IsType(t, &secp256k1.PrivateKey{}, underlying)
	assert.Equal(t, privKey, underlying)

	underlying = wrappedPubKey.Underlying()
	assert.NotNil(t, underlying)
	assert.IsType(t, &secp256k1.PublicKey{}, underlying)
	assert.Equal(t, privKey.PubKey(), underlying)
}

func TestEd25519_Underlying(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := NewPublicKey(pubKey)

	underlying := wrappedPrivKey.Underlying()
	assert.NotNil(t, underlying)
	assert.IsType(t, ed25519.PrivateKey{}, underlying)
	assert.Equal(t, privKey, underlying)

	underlying = wrappedPubKey.Underlying()
	assert.NotNil(t, underlying)
	assert.IsType(t, ed25519.PublicKey{}, underlying)
	assert.Equal(t, pubKey, underlying)
}

func TestEd25519_GetPublic(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	wrappedPrivKey := NewPrivateKey(privKey)
	wrappedPubKey := NewPublicKey(pubKey)

	publicKey := wrappedPrivKey.GetPublic()
	assert.NotNil(t, publicKey)
	assert.Equal(t, KeyTypeEd25519, publicKey.Type())
	assert.Equal(t, pubKey, publicKey.Underlying())
	assert.True(t, publicKey.Equal(wrappedPubKey))
}

func TestPublicKeyFromString_ValidSecp256k1Key(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	pubKey := privKey.PubKey()
	wrappedKey := NewPublicKey(pubKey)
	keyString := wrappedKey.String()

	parsedKey, err := PublicKeyFromString(KeyTypeSecp256k1, keyString)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)

	assert.Equal(t, KeyTypeSecp256k1, parsedKey.Type())
	assert.True(t, wrappedKey.Equal(parsedKey))

	origBytes := wrappedKey.Raw()
	parsedBytes := parsedKey.Raw()
	assert.Equal(t, origBytes, parsedBytes)
}

func TestPublicKeyFromString_ValidEd25519Key(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	wrappedKey := NewPublicKey(pubKey)
	keyString := wrappedKey.String()

	parsedKey, err := PublicKeyFromString(KeyTypeEd25519, keyString)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)

	assert.Equal(t, KeyTypeEd25519, parsedKey.Type())
	assert.True(t, wrappedKey.Equal(parsedKey))

	origBytes := wrappedKey.Raw()
	parsedBytes := parsedKey.Raw()
	assert.Equal(t, origBytes, parsedBytes)
}

func TestPublicKeyFromString_InvalidHexString(t *testing.T) {
	// Not hex encoded
	parsedKey, err := PublicKeyFromString(KeyTypeSecp256k1, "not-hex-data")
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
}

func TestPublicKeyFromString_InvalidKeyType(t *testing.T) {
	// Valid hex but wrong key type
	parsedKey, err := PublicKeyFromString("unknown-type", "deadbeef")
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.ErrorIs(t, NewErrUnsupportedKeyType(KeyType("unknown-type")), err)
	assert.ErrorContains(t, err, "unknown-type")
}

func TestPublicKeyFromString_InvalidSecp256k1KeyData(t *testing.T) {
	// Valid hex but invalid key data for secp256k1
	parsedKey, err := PublicKeyFromString(KeyTypeSecp256k1, "deadbeef")
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Equal(t, ErrInvalidECDSAPubKey, err)
}

func TestPublicKeyFromString_InvalidEd25519KeyLength(t *testing.T) {
	// Valid hex but wrong length for Ed25519
	parsedKey, err := PublicKeyFromString(KeyTypeEd25519, "deadbeef")
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Equal(t, ErrInvalidEd25519PubKeyLength, err)
}

func TestPrivateKeyFromBytes_ValidSecp256k1Key(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	keyBytes := privKey.Serialize()
	parsedKey, err := PrivateKeyFromBytes(KeyTypeSecp256k1, keyBytes)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)

	assert.Equal(t, KeyTypeSecp256k1, parsedKey.Type())

	parsedBytes := parsedKey.Raw()
	assert.Equal(t, keyBytes, parsedBytes)

	wrappedKey := NewPrivateKey(privKey)
	assert.True(t, parsedKey.GetPublic().Equal(wrappedKey.GetPublic()))
}

func TestPrivateKeyFromBytes_ValidEd25519Key(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	keyBytes := []byte(privKey)
	parsedKey, err := PrivateKeyFromBytes(KeyTypeEd25519, keyBytes)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)

	assert.Equal(t, KeyTypeEd25519, parsedKey.Type())

	parsedBytes := parsedKey.Raw()
	assert.Equal(t, keyBytes, parsedBytes)

	wrappedPubKey := NewPublicKey(pubKey)
	assert.True(t, parsedKey.GetPublic().Equal(wrappedPubKey))
}

func TestPrivateKeyFromBytes_InvalidKeyType(t *testing.T) {
	// Valid bytes but wrong key type
	parsedKey, err := PrivateKeyFromBytes(KeyType("unknown-type"), []byte{1, 2, 3, 4})
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.ErrorIs(t, NewErrUnsupportedKeyType(KeyType("unknown-type")), err)
	assert.ErrorContains(t, err, "unknown-type")
}

func TestPrivateKeyFromBytes_InvalidSecp256k1KeyLength(t *testing.T) {
	// Invalid length for secp256k1
	parsedKey, err := PrivateKeyFromBytes(KeyTypeSecp256k1, []byte{1, 2, 3, 4})
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Equal(t, ErrInvalidECDSAPrivKeyBytes, err)
}

func TestPrivateKeyFromBytes_InvalidEd25519KeyLength(t *testing.T) {
	// Invalid length for Ed25519
	parsedKey, err := PrivateKeyFromBytes(KeyTypeEd25519, []byte{1, 2, 3, 4})
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Equal(t, ErrInvalidEd25519PrivKeyLength, err)
}

func TestPrivateKeyFromString_Secp256k1ValidKey(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)

	keyString := hex.EncodeToString(privKey.Serialize())
	parsedKey, err := PrivateKeyFromString(KeyTypeSecp256k1, keyString)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)

	assert.Equal(t, KeyTypeSecp256k1, parsedKey.Type())
	assert.Equal(t, keyString, parsedKey.String())
}

func TestPrivateKeyFromString_Ed25519ValidKey(t *testing.T) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	keyString := hex.EncodeToString(privKey)
	parsedKey, err := PrivateKeyFromString(KeyTypeEd25519, keyString)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)

	assert.Equal(t, KeyTypeEd25519, parsedKey.Type())
	assert.Equal(t, keyString, parsedKey.String())
}

func TestPrivateKeyFromString_InvalidHexString(t *testing.T) {
	// Not hex encoded
	parsedKey, err := PrivateKeyFromString(KeyTypeSecp256k1, "not-hex-data")
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Contains(t, err.Error(), "encoding/hex")
}

func TestGenerateKey_Secp256k1(t *testing.T) {
	key, err := GenerateKey(KeyTypeSecp256k1)
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.Equal(t, KeyTypeSecp256k1, key.Type())

	pubKey := key.GetPublic()
	require.NotNil(t, pubKey)
	assert.Equal(t, KeyTypeSecp256k1, pubKey.Type())

	message := []byte("test message")
	sig, err := key.Sign(message)
	require.NoError(t, err)

	valid, err := pubKey.Verify(message, sig)
	require.NoError(t, err)
	assert.True(t, valid)

	assert.IsType(t, &secp256k1.PrivateKey{}, key.Underlying())
	assert.IsType(t, &secp256k1.PublicKey{}, pubKey.Underlying())
}

func TestGenerateKey_Ed25519(t *testing.T) {
	key, err := GenerateKey(KeyTypeEd25519)
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.Equal(t, KeyTypeEd25519, key.Type())

	pubKey := key.GetPublic()
	require.NotNil(t, pubKey)
	assert.Equal(t, KeyTypeEd25519, pubKey.Type())

	message := []byte("test message")
	sig, err := key.Sign(message)
	require.NoError(t, err)

	valid, err := pubKey.Verify(message, sig)
	require.NoError(t, err)
	assert.True(t, valid)

	assert.IsType(t, ed25519.PrivateKey{}, key.Underlying())
	assert.IsType(t, ed25519.PublicKey{}, pubKey.Underlying())
}

func TestGenerateKey_InvalidKeyType(t *testing.T) {
	key, err := GenerateKey("invalid-key-type")
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.ErrorIs(t, err, NewErrUnsupportedKeyType(KeyType("invalid-key-type")))
	assert.ErrorContains(t, err, "invalid-key-type")
}

func TestSecp256r1_KeyType(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	wrappedPubKey := NewPublicKey(&privKey.PublicKey)
	assert.Equal(t, KeyTypeSecp256r1, wrappedPubKey.Type())
}

func TestSecp256r1_RawBytes(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	wrappedPubKey := NewPublicKey(&privKey.PublicKey)
	pubBytes := wrappedPubKey.Raw()
	assert.Equal(t, 33, len(pubBytes))
	assert.True(t, pubBytes[0] == 0x02 || pubBytes[0] == 0x03)
}

func TestSecp256r1_Equals(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	wrappedPubKey := NewPublicKey(&privKey.PublicKey)
	otherPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	otherWrappedPubKey := NewPublicKey(&otherPrivKey.PublicKey)
	assert.True(t, wrappedPubKey.Equal(wrappedPubKey))
	assert.False(t, wrappedPubKey.Equal(otherWrappedPubKey))
}

func TestSecp256r1_Verify(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	wrappedPubKey := NewPublicKey(&privKey.PublicKey)
	message := []byte("test message")
	hash := sha256.Sum256(message)
	sig, err := ecdsa.SignASN1(rand.Reader, privKey, hash[:])
	require.NoError(t, err)

	valid, err := wrappedPubKey.Verify(message, sig)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, NewErrUnsupportedKeyType(KeyTypeSecp256r1))
}

func TestSecp256r1_DID(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	wrappedPubKey := NewPublicKey(&privKey.PublicKey)
	did, err := wrappedPubKey.DID()
	require.NoError(t, err)
	assert.Contains(t, did, "did:key:")
}

func TestSecp256r1_Underlying(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	wrappedPubKey := NewPublicKey(&privKey.PublicKey)
	underlying := wrappedPubKey.Underlying()
	assert.NotNil(t, underlying)
	assert.IsType(t, &ecdsa.PublicKey{}, underlying)
}

func TestPublicKeyFromString_ValidSecp256r1Key(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	pubKey := NewPublicKey(&privKey.PublicKey)
	wrappedKey := pubKey
	keyString := wrappedKey.String()
	parsedKey, err := PublicKeyFromString(KeyTypeSecp256r1, keyString)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)
	assert.Equal(t, KeyTypeSecp256r1, parsedKey.Type())
	assert.True(t, wrappedKey.Equal(parsedKey))

	origBytes := wrappedKey.Raw()
	parsedBytes := parsedKey.Raw()
	assert.Equal(t, origBytes, parsedBytes)
}

func TestPublicKeyFromString_ValidSecp256r1UncompressedKey(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	ecdhKey, err := privKey.PublicKey.ECDH()
	require.NoError(t, err)

	uncompressedBytes := ecdhKey.Bytes()
	keyString := hex.EncodeToString(uncompressedBytes)
	parsedKey, err := PublicKeyFromString(KeyTypeSecp256r1, keyString)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)

	assert.Equal(t, KeyTypeSecp256r1, parsedKey.Type())
	assert.True(t, NewPublicKey(&privKey.PublicKey).Equal(parsedKey))
}

func TestPublicKeyFromString_InvalidSecp256r1KeyLength(t *testing.T) {
	parsedKey, err := PublicKeyFromString(KeyTypeSecp256r1, "0102030405")
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Equal(t, ErrInvalidECDSAPubKey, err)
}

func TestPublicKeyFromString_InvalidSecp256r1CompressedPrefix(t *testing.T) {
	invalidKey := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f2021"
	parsedKey, err := PublicKeyFromString(KeyTypeSecp256r1, invalidKey)
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Equal(t, ErrInvalidECDSAPubKey, err)
}

func TestPublicKeyFromString_InvalidSecp256r1UncompressedPrefix(t *testing.T) {
	invalidKey := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f2021" + "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f2021"
	parsedKey, err := PublicKeyFromString(KeyTypeSecp256r1, invalidKey)
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.Equal(t, ErrInvalidECDSAPubKey, err)
}

func TestPrivateKeyFromBytes_Secp256r1NotSupported(t *testing.T) {
	parsedKey, err := PrivateKeyFromBytes(KeyTypeSecp256r1, make([]byte, 32))
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.ErrorIs(t, err, NewErrUnsupportedKeyType(KeyTypeSecp256r1))
}

func TestPrivateKeyFromString_Secp256r1NotSupported(t *testing.T) {
	keyString := hex.EncodeToString(make([]byte, 32))
	parsedKey, err := PrivateKeyFromString(KeyTypeSecp256r1, keyString)
	assert.Error(t, err)
	assert.Nil(t, parsedKey)
	assert.ErrorIs(t, err, NewErrUnsupportedKeyType(KeyTypeSecp256r1))
}

func TestGenerateKey_Secp256r1NotSupported(t *testing.T) {
	key, err := GenerateKey(KeyTypeSecp256r1)
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.ErrorIs(t, err, NewErrUnsupportedKeyType(KeyTypeSecp256r1))
}

func TestNewPrivateKey_Secp256r1NotSupported(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	wrappedKey := NewPrivateKey(privKey)
	assert.Nil(t, wrappedKey, "secp256r1 private keys should not be supported")
}

func TestSecp256r1_DID_Comprehensive(t *testing.T) {
	t.Run("DID from compressed key", func(t *testing.T) {
		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		wrappedPubKey := NewPublicKey(&privKey.PublicKey)
		did, err := wrappedPubKey.DID()
		require.NoError(t, err)
		assert.Contains(t, did, "did:key:")
		assert.True(t, len(did) > 20)
	})

	t.Run("DID from uncompressed key", func(t *testing.T) {
		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		ecdhKey, err := privKey.PublicKey.ECDH()
		require.NoError(t, err)

		uncompressedBytes := ecdhKey.Bytes()
		keyString := hex.EncodeToString(uncompressedBytes)
		parsedKey, err := PublicKeyFromString(KeyTypeSecp256r1, keyString)
		require.NoError(t, err)

		did, err := parsedKey.DID()
		require.NoError(t, err)
		assert.Contains(t, did, "did:key:")
	})

	t.Run("DID from key with nil Y coordinate", func(t *testing.T) {
		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		compressedBytes := elliptic.MarshalCompressed(elliptic.P256(), privKey.PublicKey.X, privKey.PublicKey.Y)
		compressedKey := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     privKey.PublicKey.X,
			Y:     nil,
		}

		wrappedKey := &secp256r1PublicKey{
			key:             compressedKey,
			compressedBytes: compressedBytes,
		}

		did, err := wrappedKey.DID()
		require.NoError(t, err)
		assert.Contains(t, did, "did:key:")
	})

	t.Run("DID consistency between compressed and uncompressed", func(t *testing.T) {
		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		compressedKey := NewPublicKey(&privKey.PublicKey)
		compressedDID, err := compressedKey.DID()
		require.NoError(t, err)

		ecdhKey, err := privKey.PublicKey.ECDH()
		require.NoError(t, err)

		uncompressedBytes := ecdhKey.Bytes()
		keyString := hex.EncodeToString(uncompressedBytes)
		uncompressedKey, err := PublicKeyFromString(KeyTypeSecp256r1, keyString)
		require.NoError(t, err)

		uncompressedDID, err := uncompressedKey.DID()
		require.NoError(t, err)

		assert.Equal(t, compressedDID, uncompressedDID)
	})
}
