package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

const (
	// EphemeralKeyLength is the size of the ECDH ephemeral key in bytes.
	EphemeralKeyLength = 65
)

// EncryptWithEphemeralKey encrypts a key using a randomly generated ephemeral ECDH key and a provided public key.
// It returns the encrypted key prepended with the ephemeral public key.
func EncryptWithEphemeralKey(plainText, publicKeyBytes []byte) ([]byte, error) {
	ephemeralPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	ephPubKeyBytes := ephemeralPriv.PublicKey().Bytes()
	sharedSecret := sha256.Sum256(append(ephPubKeyBytes, publicKeyBytes...))

	return append(ephPubKeyBytes, xorBytes(plainText, sharedSecret[:])...), nil
}

func xorBytes(data, xor []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ xor[i%len(xor)]
	}
	return result
}

// DecryptWithEphemeralKey decrypts data that was encrypted using EncryptWithEphemeralKey.
// It expects the input to be the ephemeral public key followed by the encrypted data.
func DecryptWithEphemeralKey(encryptedData, publicKeyBytes []byte) ([]byte, error) {
	ephPubKeyBytes := encryptedData[:EphemeralKeyLength]
	cipherText := make([]byte, len(encryptedData)-EphemeralKeyLength)
	copy(cipherText, encryptedData[EphemeralKeyLength:])

	sharedSecret := sha256.Sum256(append(ephPubKeyBytes, publicKeyBytes...))

	return xorBytes(cipherText, sharedSecret[:]), nil
}
