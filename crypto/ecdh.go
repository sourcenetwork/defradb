package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func GenerateX25519() (*ecdh.PrivateKey, error) {
	return ecdh.X25519().GenerateKey(rand.Reader)
}

const X25519PublicKeySize = 32

func X25519PublicKeyFromBytes(publicKeyBytes []byte) (*ecdh.PublicKey, error) {
	return ecdh.X25519().NewPublicKey(publicKeyBytes)
}

func EncryptECDH(plaintext []byte, publicKey *ecdh.PublicKey) ([]byte, error) {
	ephemeralPrivate, err := GenerateX25519()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}
	ephemeralPublic := ephemeralPrivate.PublicKey()

	sharedSecret, err := ephemeralPrivate.ECDH(publicKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH failed: %w", err)
	}

	key := sha256.Sum256(sharedSecret)

	cipherText, err := EncryptAES(plaintext, key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return append(ephemeralPublic.Bytes(), cipherText...), nil
}

func DecryptECDH(cipherText []byte, privateKey *ecdh.PrivateKey) ([]byte, error) {
	if len(cipherText) < X25519PublicKeySize+nonceLength {
		return nil, fmt.Errorf("cipherText too short")
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

	key := sha256.Sum256(sharedSecret)

	cipherText = cipherText[X25519PublicKeySize:]
	plainText, err := DecryptAES(cipherText, key[:])

	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plainText, nil
}
