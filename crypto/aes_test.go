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
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptAES(t *testing.T) {
	validKey := make([]byte, 32) // AES-256
	_, err := rand.Read(validKey)
	require.NoError(t, err)
	validPlaintext := []byte("Hello, World!")
	validAAD := []byte("Additional Authenticated Data")

	tests := []struct {
		name           string
		plainText      []byte
		key            []byte
		additionalData []byte
		prependNonce   bool
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Valid encryption with prepended nonce",
			plainText:      validPlaintext,
			key:            validKey,
			additionalData: validAAD,
			prependNonce:   true,
			expectError:    false,
		},
		{
			name:           "Valid encryption without prepended nonce",
			plainText:      validPlaintext,
			key:            validKey,
			additionalData: validAAD,
			prependNonce:   false,
			expectError:    false,
		},
		{
			name:           "Invalid key size",
			plainText:      validPlaintext,
			key:            make([]byte, 31), // Invalid key size
			additionalData: validAAD,
			prependNonce:   true,
			expectError:    true,
			errorContains:  "invalid key size",
		},
		{
			name:           "Nil plaintext",
			plainText:      nil,
			key:            validKey,
			additionalData: validAAD,
			prependNonce:   true,
			expectError:    false, // AES-GCM can encrypt nil/empty plaintext
		},
		{
			name:           "Nil additional data",
			plainText:      validPlaintext,
			key:            validKey,
			additionalData: nil,
			prependNonce:   true,
			expectError:    false, // Nil AAD is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherText, nonce, err := EncryptAES(tt.plainText, tt.key, tt.additionalData, tt.prependNonce)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				if tt.prependNonce {
					require.Greater(t, len(cipherText), len(nonce), "Ciphertext length not greater than nonce length")
				} else {
					require.Equal(t, AESNonceSize, len(nonce), "Nonce length != AESNonceSize")
				}
			}
		})
	}
}

func TestDecryptAES(t *testing.T) {
	validKey := make([]byte, 32) // AES-256
	_, err := rand.Read(validKey)
	require.NoError(t, err)
	validPlaintext := []byte("Hello, World!")
	validAAD := []byte("Additional Authenticated Data")
	validCiphertext, validNonce, _ := EncryptAES(validPlaintext, validKey, validAAD, true)

	tests := []struct {
		name           string
		nonce          []byte
		cipherText     []byte
		key            []byte
		additionalData []byte
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Valid decryption",
			nonce:          nil, // Should be extracted from cipherText
			cipherText:     validCiphertext,
			key:            validKey,
			additionalData: validAAD,
			expectError:    false,
		},
		{
			name:           "Invalid key size",
			nonce:          validNonce,
			cipherText:     validCiphertext[AESNonceSize:],
			key:            make([]byte, 31), // Invalid key size
			additionalData: validAAD,
			expectError:    true,
			errorContains:  "invalid key size",
		},
		{
			name:           "Ciphertext too short",
			nonce:          nil,
			cipherText:     make([]byte, AESNonceSize-1), // Too short to contain nonce
			key:            validKey,
			additionalData: validAAD,
			expectError:    true,
			errorContains:  errCipherTextTooShort,
		},
		{
			name:           "Invalid additional data",
			nonce:          validNonce,
			cipherText:     validCiphertext[AESNonceSize:],
			key:            validKey,
			additionalData: []byte("Wrong AAD"),
			expectError:    true,
			errorContains:  "message authentication failed",
		},
		{
			name:  "Tampered ciphertext",
			nonce: validNonce,
			// Flip a byte in the ciphertext to corrupt it.
			cipherText:     append([]byte{^validCiphertext[AESNonceSize]}, validCiphertext[AESNonceSize+1:]...),
			key:            validKey,
			additionalData: validAAD,
			expectError:    true,
			errorContains:  "message authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plainText, err := DecryptAES(tt.nonce, tt.cipherText, tt.key, tt.additionalData)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				require.True(t, bytes.Equal(plainText, validPlaintext), "Decrypted plaintext does not match original")
			}
		})
	}
}
