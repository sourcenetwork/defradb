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
	"crypto/ed25519"
	"crypto/sha256"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignECDSA_WithPrivateKeyStruct(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey, message)
	require.NoError(t, err)

	// Parse the DER signature
	signature, err := ecdsa.ParseDERSignature(sig)
	require.NoError(t, err)

	// Verify the signature
	hash := sha256.Sum256(message)
	assert.True(t, signature.Verify(hash[:], privKey.PubKey()))
}

func TestSignECDSA_WithPrivateKeyBytes(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey.Serialize(), message)
	require.NoError(t, err)

	// Parse the DER signature
	signature, err := ecdsa.ParseDERSignature(sig)
	require.NoError(t, err)

	// Verify the signature
	hash := sha256.Sum256(message)
	assert.True(t, signature.Verify(hash[:], privKey.PubKey()))
}

func TestSignEd25519_WithPrivateKeyStruct(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeEd25519, privKey, message)
	require.NoError(t, err)
	assert.Equal(t, ed25519.SignatureSize, len(sig))
	assert.True(t, ed25519.Verify(pubKey, message, sig))
}

func TestSignEd25519_WithPrivateKeyBytes(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeEd25519, []byte(privKey), message)
	require.NoError(t, err)
	assert.Equal(t, ed25519.SignatureSize, len(sig))
	assert.True(t, ed25519.Verify(pubKey, message, sig))
}

func TestSign_InvalidSignatureType(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	_, err = Sign(SignatureType(99), privKey, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported signature type")
}

func TestSignECDSA256K_WithPrivateKeyStruct(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignECDSA256K(privKey, message)
	require.NoError(t, err)

	signature, err := ecdsa.ParseDERSignature(sig)
	require.NoError(t, err)

	hash := sha256.Sum256(message)
	assert.True(t, signature.Verify(hash[:], privKey.PubKey()))
}

func TestSignECDSA256K_WithPrivateKeyBytes(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignECDSA256K(privKey.Serialize(), message)
	require.NoError(t, err)

	signature, err := ecdsa.ParseDERSignature(sig)
	require.NoError(t, err)

	hash := sha256.Sum256(message)
	assert.True(t, signature.Verify(hash[:], privKey.PubKey()))
}

func TestSignEd25519_Direct_WithPrivateKeyStruct(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignEd25519(privKey, message)
	require.NoError(t, err)
	assert.Equal(t, ed25519.SignatureSize, len(sig))
	assert.True(t, ed25519.Verify(pubKey, message, sig))
}

func TestSignEd25519_Direct_WithPrivateKeyBytes(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignEd25519([]byte(privKey), message)
	require.NoError(t, err)
	assert.Equal(t, ed25519.SignatureSize, len(sig))
	assert.True(t, ed25519.Verify(pubKey, message, sig))
}

func TestSign_InvalidPrivateKeyType(t *testing.T) {
	message := []byte("test message")
	_, err := Sign(SignatureTypeECDSA256K, []byte("invalid key"), message)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidECDSAPrivKeyBytes)

	_, err = Sign(SignatureTypeEd25519, []byte("invalid key"), message)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEd25519PrivKeyLength)
}

func TestVerifyECDSA_WithPublicKeyStruct(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeECDSA256K, privKey.PubKey(), message, sig)
	require.NoError(t, err)
}

func TestVerifyECDSA_WithPublicKeyBytes(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeECDSA256K, privKey.PubKey().SerializeCompressed(), message, sig)
	require.NoError(t, err)
}

func TestVerifyEd25519_WithPublicKeyStruct(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeEd25519, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeEd25519, pubKey, message, sig)
	require.NoError(t, err)
}

func TestVerifyEd25519_WithPublicKeyBytes(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeEd25519, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeEd25519, []byte(pubKey), message, sig)
	require.NoError(t, err)
}

func TestVerifyECDSA256K_Direct_WithPublicKeyStruct(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignECDSA256K(privKey, message)
	require.NoError(t, err)

	err = VerifyECDSA256K(privKey.PubKey(), message, sig)
	require.NoError(t, err)
}

func TestVerifyECDSA256K_Direct_WithPublicKeyBytes(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignECDSA256K(privKey, message)
	require.NoError(t, err)

	err = VerifyECDSA256K(privKey.PubKey().SerializeCompressed(), message, sig)
	require.NoError(t, err)
}

func TestVerifyEd25519_Direct_WithPublicKeyStruct(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignEd25519(privKey, message)
	require.NoError(t, err)

	err = VerifyEd25519(pubKey, message, sig)
	require.NoError(t, err)
}

func TestVerifyEd25519_Direct_WithPublicKeyBytes(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := SignEd25519(privKey, message)
	require.NoError(t, err)

	err = VerifyEd25519([]byte(pubKey), message, sig)
	require.NoError(t, err)
}

func TestVerifyECDSA_TamperedMessage(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	originalMessage := []byte("original message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey, originalMessage)
	require.NoError(t, err)

	tamperedMessage := []byte("tampered message")

	err = Verify(SignatureTypeECDSA256K, pubKey, tamperedMessage, sig)
	require.ErrorIs(t, err, ErrSignatureVerification)
}

func TestVerifyEd25519_TamperedMessage(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	originalMessage := []byte("original message")
	sig, err := Sign(SignatureTypeEd25519, privKey, originalMessage)
	require.NoError(t, err)

	tamperedMessage := []byte("tampered message")

	err = Verify(SignatureTypeEd25519, pubKey, tamperedMessage, sig)
	require.ErrorIs(t, err, ErrSignatureVerification)
}

func TestVerifyECDSA_TamperedSignature(t *testing.T) {
	priv, err := GenerateSecp256k1()
	require.NoError(t, err)
	pubKey := priv.PubKey()
	message := []byte("test message")

	sig, err := Sign(SignatureTypeECDSA256K, priv, message)
	require.NoError(t, err)

	signature, err := ecdsa.ParseDERSignature(sig)
	require.NoError(t, err)

	// Create a new ModNScalar with a slightly different value
	one := new(secp256k1.ModNScalar).SetInt(1)
	r := signature.R()
	r.Add(one)
	s := signature.S()

	modifiedSig := ecdsa.NewSignature(&r, &s)

	err = Verify(SignatureTypeECDSA256K, pubKey, message, modifiedSig.Serialize())
	require.ErrorIs(t, err, ErrSignatureVerification)
}

func TestVerifyEd25519_TamperedSignature(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	message := []byte("test message")

	sig, err := Sign(SignatureTypeEd25519, privKey, message)
	require.NoError(t, err)

	tamperedSig := make([]byte, len(sig))
	copy(tamperedSig, sig)
	tamperedSig[0] ^= 0xff // Flip bits in first byte to tamper with signature

	err = Verify(SignatureTypeEd25519, pubKey, message, tamperedSig)
	require.ErrorIs(t, err, ErrSignatureVerification)
}

func TestVerifyECDSA_WrongPublicKey(t *testing.T) {
	correctPriv, err := GenerateSecp256k1()
	require.NoError(t, err)

	wrongPriv, err := GenerateSecp256k1()
	require.NoError(t, err)
	wrongPub := wrongPriv.PubKey()

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, correctPriv, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeECDSA256K, wrongPub, message, sig)
	require.ErrorIs(t, err, ErrSignatureVerification)
}

func TestVerifyEd25519_WrongPublicKey(t *testing.T) {
	_, correctPriv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	wrongPub, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeEd25519, correctPriv, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeEd25519, wrongPub, message, sig)
	require.ErrorIs(t, err, ErrSignatureVerification)
}

func TestVerify_InvalidSignatureType(t *testing.T) {
	// Test with an invalid signature type
	invalidSigType := SignatureType(99)
	err := Verify(invalidSigType, []byte("any"), []byte("any"), []byte("any"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedSignatureType)
	require.Contains(t, err.Error(), "unsupported signature type")
}
