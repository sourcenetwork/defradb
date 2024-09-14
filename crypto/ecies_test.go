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
	"crypto/ecdh"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptECIES_Errors(t *testing.T) {
	validAssociatedData := []byte("associated data")
	validPrivateKey, _ := GenerateX25519()

	tests := []struct {
		name        string
		plainText   []byte
		publicKey   *ecdh.PublicKey
		opts        []ECIESOption
		expectError string
	}{
		{
			name:        "Invalid public key",
			plainText:   []byte("test data"),
			publicKey:   &ecdh.PublicKey{},
			opts:        []ECIESOption{WithAAD(validAssociatedData)},
			expectError: errFailedECDHOperation,
		},
		{
			name:        "No public key prepended and no private key provided",
			plainText:   []byte("test data"),
			publicKey:   validPrivateKey.PublicKey(),
			opts:        []ECIESOption{WithPubKeyPrepended(false)},
			expectError: errNoPublicKeyForDecryption,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptECIES(tt.plainText, tt.publicKey, tt.opts...)
			if err == nil {
				t.Errorf("Expected an error, but got nil")
			} else if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectError, err)
			}
		})
	}
}

func TestDecryptECIES_Errors(t *testing.T) {
	validPrivateKey, _ := GenerateX25519()
	aad := []byte("associated data")
	validCipherText, _ := EncryptECIES([]byte("test data test data"), validPrivateKey.PublicKey(), WithAAD(aad))

	tests := []struct {
		name        string
		cipherText  []byte
		privateKey  *ecdh.PrivateKey
		opts        []ECIESOption
		expectError string
	}{
		{
			name:        "Ciphertext too short",
			cipherText:  []byte("short"),
			privateKey:  validPrivateKey,
			opts:        []ECIESOption{WithAAD(aad)},
			expectError: errCipherTextTooShort,
		},
		{
			name:        "Invalid private key",
			cipherText:  validCipherText,
			privateKey:  &ecdh.PrivateKey{},
			opts:        []ECIESOption{WithAAD(aad)},
			expectError: errFailedECDHOperation,
		},
		{
			name:        "Tampered ciphertext",
			cipherText:  append(validCipherText, byte(0)),
			privateKey:  validPrivateKey,
			opts:        []ECIESOption{WithAAD(aad)},
			expectError: errVerificationWithHMACFailed,
		},
		{
			name:        "Wrong associated data",
			cipherText:  validCipherText,
			privateKey:  validPrivateKey,
			opts:        []ECIESOption{WithAAD([]byte("wrong data"))},
			expectError: errFailedToDecrypt,
		},
		{
			name:        "No public key prepended and no public key bytes provided",
			cipherText:  validCipherText[X25519PublicKeySize:],
			privateKey:  validPrivateKey,
			opts:        []ECIESOption{WithAAD(aad), WithPubKeyPrepended(false)},
			expectError: errNoPublicKeyForDecryption,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptECIES(tt.cipherText, tt.privateKey, tt.opts...)
			if err == nil || !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got %v", tt.expectError, err)
			}
		})
	}
}

func TestEncryptDecryptECIES_DefaultOptions_Succeeds(t *testing.T) {
	plainText := []byte("Hello, World!")
	recipientPrivateKey := mustGenerateX25519(t)

	cipherText, err := EncryptECIES(plainText, recipientPrivateKey.PublicKey())
	require.NoError(t, err)

	decryptedText, err := DecryptECIES(cipherText, recipientPrivateKey)
	require.NoError(t, err)

	assert.Equal(t, plainText, decryptedText)
}

func TestEncryptDecryptECIES_WithAAD_Succeeds(t *testing.T) {
	plainText := []byte("Secret message")
	aad := []byte("extra authentication data")
	recipientPrivateKey := mustGenerateX25519(t)

	cipherText, err := EncryptECIES(plainText, recipientPrivateKey.PublicKey(), WithAAD(aad))
	require.NoError(t, err)

	decryptedText, err := DecryptECIES(cipherText, recipientPrivateKey, WithAAD(aad))
	require.NoError(t, err)

	assert.Equal(t, plainText, decryptedText)
}

func TestEncryptDecryptECIES_WithCustomPrivateKey_Succeeds(t *testing.T) {
	plainText := []byte("Custom key message")
	recipientPrivateKey := mustGenerateX25519(t)
	senderPrivateKey := mustGenerateX25519(t)

	cipherText, err := EncryptECIES(plainText, recipientPrivateKey.PublicKey(), WithPrivKey(senderPrivateKey))
	require.NoError(t, err)

	require.Equal(t, senderPrivateKey.PublicKey().Bytes(), cipherText[:X25519PublicKeySize])

	decryptedText, err := DecryptECIES(cipherText, recipientPrivateKey)
	require.NoError(t, err)

	assert.Equal(t, plainText, decryptedText)
}

func TestEncryptDecryptECIES_WithoutPublicKeyPrepended_Succeeds(t *testing.T) {
	plainText := []byte("No prepended key")
	recipientPrivateKey := mustGenerateX25519(t)
	senderPrivateKey := mustGenerateX25519(t)

	cipherText, err := EncryptECIES(plainText, recipientPrivateKey.PublicKey(),
		WithPubKeyPrepended(false),
		WithPrivKey(senderPrivateKey))
	require.NoError(t, err)

	// In a real scenario, the public key would be transmitted separately
	senderPublicKeyBytes := senderPrivateKey.PublicKey().Bytes()

	decryptedText, err := DecryptECIES(cipherText, recipientPrivateKey,
		WithPubKeyPrepended(false),
		WithPubKeyBytes(senderPublicKeyBytes))
	require.NoError(t, err)

	assert.Equal(t, plainText, decryptedText)
}

func TestEncryptDecryptECIES_DifferentAAD_FailsToDecrypt(t *testing.T) {
	plainText := []byte("AAD test message")
	encryptAAD := []byte("encryption AAD")
	decryptAAD := []byte("decryption AAD")
	recipientPrivateKey := mustGenerateX25519(t)

	cipherText, err := EncryptECIES(plainText, recipientPrivateKey.PublicKey(), WithAAD(encryptAAD))
	require.NoError(t, err)

	_, err = DecryptECIES(cipherText, recipientPrivateKey, WithAAD(decryptAAD))
	assert.Error(t, err, "Decryption should fail with different AAD")
}

func mustGenerateX25519(t *testing.T) *ecdh.PrivateKey {
	key, err := GenerateX25519()
	require.NoError(t, err)
	return key
}
