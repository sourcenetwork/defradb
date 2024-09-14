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

type ECIESOption func(*eciesOptions)

type eciesOptions struct {
	associatedData    []byte
	privateKey        *ecdh.PrivateKey
	publicKeyBytes    []byte
	noPubKeyPrepended bool
}

// WithAAD sets the associated data to use for authentication.
func WithAAD(aad []byte) ECIESOption {
	return func(o *eciesOptions) {
		o.associatedData = aad
	}
}

// WithPrivKey sets the private key to use for encryption.
//
// If not set, a new ephemeral key will be generated.
// This option has no effect on decryption.
func WithPrivKey(privKey *ecdh.PrivateKey) ECIESOption {
	return func(o *eciesOptions) {
		o.privateKey = privKey
	}
}

// WithPubKeyBytes sets the public key bytes to use for decryption.
//
// If not set, the cipherText is assumed to have the public key X25519 prepended.
// This option has no effect on encryption.
func WithPubKeyBytes(pubKeyBytes []byte) ECIESOption {
	return func(o *eciesOptions) {
		o.publicKeyBytes = pubKeyBytes
	}
}

// WithPubKeyPrepended sets whether the public key should is prepended to the cipherText.
//
// Upon encryption, if set to true (default value), the public key is prepended to the cipherText.
// Otherwise it's not and in this case a private key should be provided with the WithPrivKey option.
//
// Upon decryption, if set to true (default value), the public key is expected to be prepended to the cipherText.
// Otherwise it's not and in this case the public key bytes should be provided with the WithPubKeyBytes option.
func WithPubKeyPrepended(prepended bool) ECIESOption {
	return func(o *eciesOptions) {
		o.noPubKeyPrepended = !prepended
	}
}

// EncryptECIES encrypts plaintext using a custom Elliptic Curve Integrated Encryption Scheme (ECIES)
// with X25519 for key agreement, HKDF for key derivation, AES for encryption, and HMAC for authentication.
//
// The function:
// - Uses or generates an ephemeral X25519 key pair
// - Performs ECDH with the provided public key
// - Derives encryption and HMAC keys using HKDF
// - Encrypts the plaintext using a custom AES encryption function
// - Computes an HMAC over the ciphertext
//
// The default output format is: [ephemeral public key | encrypted data (including nonce) | HMAC]
// This can be modified using options.
//
// Parameters:
//   - plainText: The message to encrypt
//   - publicKey: The recipient's X25519 public key
//   - opts: Optional ECIESOption functions to customize the encryption process
//
// Available options:
//   - WithAAD(aad []byte): Sets the associated data for additional authentication
//   - WithPrivKey(privKey *ecdh.PrivateKey): Uses the provided private key instead of generating a new one
//   - WithPubKeyPrepended(prepended bool): Controls whether the public key is prepended to the ciphertext
//
// Returns:
//   - Byte slice containing the encrypted message and necessary metadata for decryption
//   - Error if any step of the encryption process fails
//
// Example usage:
//
//	cipherText, err := EncryptECIES(plainText, recipientPublicKey,
//	    WithAAD(additionalData),
//	    WithPrivKey(senderPrivateKey),
//	    WithPubKeyPrepended(false))
func EncryptECIES(plainText []byte, publicKey *ecdh.PublicKey, opts ...ECIESOption) ([]byte, error) {
	options := &eciesOptions{}
	for _, opt := range opts {
		opt(options)
	}

	ourPrivateKey := options.privateKey
	if ourPrivateKey == nil {
		if options.noPubKeyPrepended {
			return nil, ErrNoPublicKeyForDecryption
		}
		var err error
		ourPrivateKey, err = GenerateX25519()
		if err != nil {
			return nil, NewErrFailedToGenerateEphemeralKey(err)
		}
	}
	ourPublicKey := ourPrivateKey.PublicKey()

	sharedSecret, err := ourPrivateKey.ECDH(publicKey)
	if err != nil {
		return nil, NewErrFailedECDHOperation(err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)
	aesKey := make([]byte, AESKeySize)
	hmacKey := make([]byte, HMACSize)
	if _, err := kdf.Read(aesKey); err != nil {
		return nil, NewErrFailedKDFOperationForAESKey(err)
	}
	if _, err := kdf.Read(hmacKey); err != nil {
		return nil, NewErrFailedKDFOperationForHMACKey(err)
	}

	cipherText, _, err := EncryptAES(plainText, aesKey, makeAAD(ourPublicKey.Bytes(), options.associatedData), true)
	if err != nil {
		return nil, NewErrFailedToEncrypt(err)
	}

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipherText)
	macSum := mac.Sum(nil)

	var result []byte
	if options.noPubKeyPrepended {
		result = cipherText
	} else {
		result = append(ourPublicKey.Bytes(), cipherText...)
	}
	result = append(result, macSum...)

	return result, nil
}

// DecryptECIES decrypts ciphertext encrypted with EncryptECIES using the provided private key.
//
// The function:
// - Extracts or uses the provided ephemeral public key
// - Performs ECDH with the provided private key
// - Derives decryption and HMAC keys using HKDF
// - Verifies the HMAC
// - Decrypts the message using a custom AES decryption function
//
// The default expected input format is: [ephemeral public key | encrypted data (including nonce) | HMAC]
// This can be modified using options.
//
// Parameters:
//   - cipherText: The encrypted message, including all necessary metadata
//   - privateKey: The recipient's X25519 private key
//   - opts: Optional ECIESOption functions to customize the decryption process
//
// Available options:
//   - WithAAD(aad []byte): Sets the associated data used during encryption for additional authentication
//   - WithPubKeyBytes(pubKeyBytes []byte): Provides the public key bytes if not prepended to the ciphertext
//   - WithPubKeyPrepended(prepended bool): Indicates whether the public key is prepended to the ciphertext
//
// Returns:
//   - Byte slice containing the decrypted plaintext
//   - Error if any step of the decryption process fails, including authentication failure
//
// Example usage:
//
//	plainText, err := DecryptECIES(cipherText, recipientPrivateKey,
//	    WithAAD(additionalData),
//	    WithPubKeyBytes(senderPublicKeyBytes),
//	    WithPubKeyPrepended(false))
func DecryptECIES(cipherText []byte, ourPrivateKey *ecdh.PrivateKey, opts ...ECIESOption) ([]byte, error) {
	options := &eciesOptions{}
	for _, opt := range opts {
		opt(options)
	}

	minLength := X25519PublicKeySize + AESNonceSize + HMACSize + minCipherTextSize
	if options.noPubKeyPrepended {
		minLength -= X25519PublicKeySize
	}

	if len(cipherText) < minLength {
		return nil, ErrCipherTextTooShort
	}

	publicKeyBytes := options.publicKeyBytes
	if options.publicKeyBytes == nil {
		if options.noPubKeyPrepended {
			return nil, ErrNoPublicKeyForDecryption
		}
		publicKeyBytes = cipherText[:X25519PublicKeySize]
		cipherText = cipherText[X25519PublicKeySize:]
	}
	publicKey, err := ecdh.X25519().NewPublicKey(publicKeyBytes)
	if err != nil {
		return nil, NewErrFailedToParseEphemeralPublicKey(err)
	}

	sharedSecret, err := ourPrivateKey.ECDH(publicKey)
	if err != nil {
		return nil, NewErrFailedECDHOperation(err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)
	aesKey := make([]byte, AESKeySize)
	hmacKey := make([]byte, HMACSize)
	if _, err := kdf.Read(aesKey); err != nil {
		return nil, NewErrFailedKDFOperationForAESKey(err)
	}
	if _, err := kdf.Read(hmacKey); err != nil {
		return nil, NewErrFailedKDFOperationForHMACKey(err)
	}

	macSum := cipherText[len(cipherText)-HMACSize:]
	cipherTextWithNonce := cipherText[:len(cipherText)-HMACSize]

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipherTextWithNonce)
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(macSum, expectedMAC) {
		return nil, ErrVerificationWithHMACFailed
	}

	plainText, err := DecryptAES(nil, cipherTextWithNonce, aesKey, makeAAD(publicKeyBytes, options.associatedData))
	if err != nil {
		return nil, NewErrFailedToDecrypt(err)
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
