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
)

func TestEncryptECIES_Errors(t *testing.T) {
	validAssociatedData := []byte("associated data")

	tests := []struct {
		name           string
		plainText      []byte
		publicKey      *ecdh.PublicKey
		associatedData []byte
		expectError    string
	}{
		{
			name:           "Invalid public key",
			plainText:      []byte("test data"),
			publicKey:      &ecdh.PublicKey{},
			associatedData: validAssociatedData,
			expectError:    "failed ECDH operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptECIES(tt.plainText, tt.publicKey, tt.associatedData)
			if err == nil {
				t.Errorf("Expected an error, but got nil")
			} else if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectError, err)
			}
		})
	}
}

func TestDecryptECIES_Errors(t *testing.T) {
	// Setup
	validPrivateKey, _ := GenerateX25519()
	validCipherText, _ := EncryptECIES([]byte("test data"), validPrivateKey.PublicKey(), []byte("associated data"))

	tests := []struct {
		name           string
		cipherText     []byte
		privateKey     *ecdh.PrivateKey
		associatedData []byte
		expectError    string
	}{
		{
			name:           "Ciphertext too short",
			cipherText:     []byte("short"),
			privateKey:     validPrivateKey,
			associatedData: []byte("associated data"),
			expectError:    "ciphertext too short",
		},
		{
			name:           "Invalid private key",
			cipherText:     validCipherText,
			privateKey:     &ecdh.PrivateKey{},
			associatedData: []byte("associated data"),
			expectError:    "failed ECDH operation",
		},
		{
			name:           "Tampered ciphertext",
			cipherText:     append(validCipherText, byte(0)),
			privateKey:     validPrivateKey,
			associatedData: []byte("associated data"),
			expectError:    "verification with HMAC failed",
		},
		{
			name:           "Wrong associated data",
			cipherText:     validCipherText,
			privateKey:     validPrivateKey,
			associatedData: []byte("wrong data"),
			expectError:    "failed to decrypt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptECIES(tt.cipherText, tt.privateKey, tt.associatedData)
			if err == nil || !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got %v", tt.expectError, err)
			}
		})
	}
}
