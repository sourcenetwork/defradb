// Current Implementation Analysis:
// 1. Key Generation: Correctly uses X25519 for key generation.
// 2. ECDH: Properly performs the ECDH operation.
// 3. Key Derivation: Uses SHA-256 on the shared secret, which is simplistic.
// 4. Encryption: Uses AES (implementation not shown).
// 5. MAC: Not implemented.

// Improvements Needed:
// 1. Use a proper Key Derivation Function (KDF)
// 2. Implement HMAC for message authentication
// 3. Use authenticated encryption (e.g., AES-GCM) instead of AES
// 4. Standardize the output format

// Here's an improved version of the EncryptECDH and DecryptECDH functions:

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

func GenerateX25519() (*ecdh.PrivateKey, error) {
	return ecdh.X25519().GenerateKey(rand.Reader)
}

func X25519PublicKeyFromBytes(publicKeyBytes []byte) (*ecdh.PublicKey, error) {
	return ecdh.X25519().NewPublicKey(publicKeyBytes)
}

func EncryptECIES(plainText []byte, publicKey *ecdh.PublicKey) ([]byte, error) {
	ephemeralPrivate, err := GenerateX25519()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}
	ephemeralPublic := ephemeralPrivate.PublicKey()

	sharedSecret, err := ephemeralPrivate.ECDH(publicKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH failed: %w", err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)
	aesKey := make([]byte, AESKeySize)
	hmacKey := make([]byte, HMACSize)
	if _, err := kdf.Read(aesKey); err != nil {
		return nil, fmt.Errorf("KDF failed for AES key: %w", err)
	}
	if _, err := kdf.Read(hmacKey); err != nil {
		return nil, fmt.Errorf("KDF failed for HMAC key: %w", err)
	}

	cipherText, err := EncryptAES(plainText, aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipherText[AESNonceSize:])
	macSum := mac.Sum(nil)

	result := append(ephemeralPublic.Bytes(), cipherText...)
	result = append(result, macSum...)

	return result, nil
}

func DecryptECIES(cipherText []byte, privateKey *ecdh.PrivateKey) ([]byte, error) {
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
		return nil, fmt.Errorf("ECDH failed: %w", err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)
	aesKey := make([]byte, AESKeySize)
	hmacKey := make([]byte, HMACSize)
	if _, err := kdf.Read(aesKey); err != nil {
		return nil, fmt.Errorf("KDF failed for AES key: %w", err)
	}
	if _, err := kdf.Read(hmacKey); err != nil {
		return nil, fmt.Errorf("KDF failed for HMAC key: %w", err)
	}

	macSum := cipherText[len(cipherText)-HMACSize:]
	cipherText = cipherText[:len(cipherText)-HMACSize]

	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipherText[X25519PublicKeySize+AESNonceSize:])
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(macSum, expectedMAC) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	plainText, err := DecryptAES(cipherText[X25519PublicKeySize:], aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plainText, nil
}
