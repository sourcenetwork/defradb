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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/hkdf"
)

const X25519PublicKeySize = 32
const HMACSize = 32
const AESKeySize = 32

const minCipherTextSize = 16

// GenerateX25519 generates a new X25519 private key.
func GenerateX25519() (*ecdh.PrivateKey, error) {
	return ecdh.X25519().GenerateKey(rand.Reader)
}

// X25519PublicKeyFromBytes creates a new X25519 public key from the given bytes.
func X25519PublicKeyFromBytes(publicKeyBytes []byte) (*ecdh.PublicKey, error) {
	return ecdh.X25519().NewPublicKey(publicKeyBytes)
}

// EncryptECIES encrypts plaintext using a custom Elliptic Curve Integrated Encryption Scheme (ECIES)
// with X25519 for key agreement, HKDF for key derivation, AES for encryption, and HMAC for authentication.
//
// The function:
// - Generates an ephemeral X25519 key pair
// - Performs ECDH with the provided public key
// - Derives encryption and HMAC keys using HKDF
// - Encrypts the plaintext using a custom AES encryption function
// - Computes an HMAC over the ciphertext
//
// The output format is: [ephemeral public key | encrypted data (including nonce) | HMAC]
//
// Parameters:
//   - plainText: The message to encrypt
//   - publicKey: The recipient's X25519 public key
//   - associatedData: Optional associated data for additional authentication
//
// Returns:
//   - Byte slice containing the encrypted message and necessary metadata for decryption
//   - Error if any step of the encryption process fails
func EncryptECIES(plainText []byte, publicKey *ecdh.PublicKey, associatedData []byte) ([]byte, error) {
	ephemeralPrivate, err := GenerateX25519()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}
	ephemeralPublic := ephemeralPrivate.PublicKey()

	sharedSecret, err := ephemeralPrivate.ECDH(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed ECDH operation: %w", err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)
	aesKey := make([]byte, AESKeySize)
	hmacKey := make([]byte, HMACSize)
	if _, err := kdf.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed KDF operation for AES key: %w", err)
	}
	if _, err := kdf.Read(hmacKey); err != nil {
		return nil, fmt.Errorf("failed KDF operation for HMAC key: %w", err)
	}

	cipherText, _, err := EncryptAES(plainText, aesKey, makeAAD(ephemeralPrivate.Bytes(), associatedData), true)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipherText)
	macSum := mac.Sum(nil)

	result := append(ephemeralPublic.Bytes(), cipherText...)
	result = append(result, macSum...)

	return result, nil
}

// DecryptECIES decrypts ciphertext encrypted with EncryptECIES using the provided private key.
//
// The function:
// - Extracts the ephemeral public key from the ciphertext
// - Performs ECDH with the provided private key
// - Derives decryption and HMAC keys using HKDF
// - Verifies the HMAC
// - Decrypts the message using a custom AES decryption function
//
// The expected input format is: [ephemeral public key | encrypted data (including nonce) | HMAC]
//
// Parameters:
//   - cipherText: The encrypted message, including all necessary metadata
//   - privateKey: The recipient's X25519 private key
//   - associatedData: Optional associated data used during encryption for additional authentication
//
// Returns:
//   - Byte slice containing the decrypted plaintext
//   - Error if any step of the decryption process fails, including authentication failure
func DecryptECIES(cipherText []byte, privateKey *ecdh.PrivateKey, associatedData []byte) ([]byte, error) {
	if len(cipherText) < X25519PublicKeySize+AESNonceSize+HMACSize+minCipherTextSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	ephemeralPublicBytes := cipherText[:X25519PublicKeySize]
	ephemeralPublic, err := ecdh.X25519().NewPublicKey(ephemeralPublicBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ephemeral public key: %w", err)
	}

	sharedSecret, err := privateKey.ECDH(ephemeralPublic)
	if err != nil {
		return nil, fmt.Errorf("failed ECDH operation: %w", err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)
	aesKey := make([]byte, AESKeySize)
	hmacKey := make([]byte, HMACSize)
	if _, err := kdf.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed KDF operation for AES key: %w", err)
	}
	if _, err := kdf.Read(hmacKey); err != nil {
		return nil, fmt.Errorf("failed KDF operation for HMAC key: %w", err)
	}

	macSum := cipherText[len(cipherText)-HMACSize:]
	cipherTextWithNonce := cipherText[X25519PublicKeySize : len(cipherText)-HMACSize]

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipherTextWithNonce)
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(macSum, expectedMAC) {
		return nil, fmt.Errorf("verification with HMAC failed")
	}

	plainText, err := DecryptAES(nil, cipherTextWithNonce, aesKey, makeAAD(ephemeralPublicBytes, associatedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plainText, nil
}

// makeAAD concatenates the ephemeral public key and associated data for use as additional authenticated data.
func makeAAD(ephemeralPublicBytes, associatedData []byte) []byte {
	l := len(ephemeralPublicBytes) + len(associatedData)
	aad := make([]byte, l)
	copy(aad, ephemeralPublicBytes)
	copy(aad[len(ephemeralPublicBytes):], associatedData)
	return aad
}
