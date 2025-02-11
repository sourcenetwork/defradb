package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignECDSA_WithPrivateKeyStruct(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA, privKey, message)
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
	sig, err := Sign(SignatureTypeECDSA, privKey.Serialize(), message)
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

func TestSign_InvalidPrivateKeyType(t *testing.T) {
	message := []byte("test message")
	_, err := Sign(SignatureTypeECDSA, "invalid key", message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported ECDSA private key type")

	_, err = Sign(SignatureTypeEd25519, "invalid key", message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported Ed25519 private key type")
}
