package encryption

import (
	"crypto/rand"
	"errors"
	"io"
	"os"
	"strings"
)

const nonceLength = 12

var generateNonceFunc = generateNonce

func generateNonce() ([]byte, error) {
	nonce := make([]byte, nonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return nonce, nil
}

// generateTestNonce generates a deterministic nonce for testing.
func generateTestNonce() ([]byte, error) {
	nonce := []byte("deterministic nonce for testing")

	if len(nonce) < nonceLength {
		return nil, errors.New("nonce length is longer than available deterministic nonce")
	}

	return nonce[:nonceLength], nil
}

func init() {
	arg := os.Args[0]
	// If the binary is a test binary, use a deterministic nonce.
	if strings.HasSuffix(arg, ".test") || strings.Contains(arg, "/defradb/tests/") {
		generateNonceFunc = generateTestNonce
	}
}
