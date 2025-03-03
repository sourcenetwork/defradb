package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignECDSA_WithPrivateKeyStruct(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey, message)
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
	sig, err := Sign(SignatureTypeECDSA256K, privKey.Serialize(), message)
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
	_, err := Sign(SignatureTypeECDSA256K, "invalid key", message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported ECDSA private key type")

	_, err = Sign(SignatureTypeEd25519, "invalid key", message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported Ed25519 private key type")
}

func TestVerifyECDSA_WithPublicKeyStruct(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeECDSA256K, privKey.PubKey(), message, sig)
	require.NoError(t, err)
}

func TestVerifyECDSA_WithPublicKeyBytes(t *testing.T) {
	privKey, err := GenerateSecp256k1()
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeECDSA256K, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeECDSA256K, privKey.PubKey().SerializeCompressed(), message, sig)
	require.NoError(t, err)
}

func TestVerifyEd25519_WithPublicKeyStruct(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeEd25519, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeEd25519, pubKey, message, sig)
	require.NoError(t, err)
}

func TestVerifyEd25519_WithPublicKeyBytes(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	sig, err := Sign(SignatureTypeEd25519, privKey, message)
	require.NoError(t, err)

	err = Verify(SignatureTypeEd25519, []byte(pubKey), message, sig)
	require.NoError(t, err)
}

func TestVerify_TamperedMessage(t *testing.T) {
	tests := []struct {
		name      string
		sigType   SignatureType
		setupKeys func(t *testing.T) (pubKey, privKey interface{})
	}{
		{
			name:    "ECDSA tampered message",
			sigType: SignatureTypeECDSA256K,
			setupKeys: func(t *testing.T) (pubKey, privKey interface{}) {
				priv, err := GenerateSecp256k1()
				require.NoError(t, err)
				return priv.PubKey(), priv
			},
		},
		{
			name:    "Ed25519 tampered message",
			sigType: SignatureTypeEd25519,
			setupKeys: func(t *testing.T) (pubKey, privKey interface{}) {
				pub, priv, err := ed25519.GenerateKey(nil)
				require.NoError(t, err)
				return pub, priv
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey, privKey := tt.setupKeys(t)
			message := []byte("original message")
			tamperedMessage := []byte("tampered message")

			sig, err := Sign(tt.sigType, privKey, message)
			require.NoError(t, err)

			err = Verify(tt.sigType, pubKey, tamperedMessage, sig)
			require.ErrorIs(t, err, ErrSignatureVerification)
		})
	}
}

func TestVerify_TamperedSignature(t *testing.T) {
	tests := []struct {
		name      string
		sigType   SignatureType
		setupKeys func(t *testing.T) (pubKey, privKey interface{})
	}{
		{
			name:    "ECDSA tampered signature",
			sigType: SignatureTypeECDSA256K,
			setupKeys: func(t *testing.T) (pubKey, privKey interface{}) {
				priv, err := GenerateSecp256k1()
				require.NoError(t, err)
				return priv.PubKey(), priv
			},
		},
		{
			name:    "Ed25519 tampered signature",
			sigType: SignatureTypeEd25519,
			setupKeys: func(t *testing.T) (pubKey, privKey interface{}) {
				pub, priv, err := ed25519.GenerateKey(nil)
				require.NoError(t, err)
				return pub, priv
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey, privKey := tt.setupKeys(t)
			message := []byte("test message")

			sig, err := Sign(tt.sigType, privKey, message)
			require.NoError(t, err)

			switch tt.sigType {
			case SignatureTypeECDSA256K:
				// For ECDSA, parse the DER signature first, modify it, then serialize back
				signature, err := ecdsa.ParseDERSignature(sig)
				require.NoError(t, err)

				// Create a new ModNScalar with a slightly different value
				one := new(secp256k1.ModNScalar).SetInt(1)
				r := signature.R()
				r.Add(one)

				s := signature.S()
				modifiedSig := ecdsa.NewSignature(&r, &s)
				err = Verify(tt.sigType, pubKey, message, modifiedSig.Serialize())
				require.ErrorIs(t, err, ErrSignatureVerification)

			case SignatureTypeEd25519:
				// For Ed25519, we can directly modify the signature bytes
				sig[0] ^= 0xff
				err = Verify(tt.sigType, pubKey, message, sig)
				require.ErrorIs(t, err, ErrSignatureVerification)
			}
		})
	}
}

func TestVerify_WrongPublicKey(t *testing.T) {
	tests := []struct {
		name      string
		sigType   SignatureType
		setupKeys func(t *testing.T) (correctPub, correctPriv, wrongPub interface{})
		expectErr error
	}{
		{
			name:    "ECDSA wrong public key",
			sigType: SignatureTypeECDSA256K,
			setupKeys: func(t *testing.T) (correctPub, correctPriv, wrongPub interface{}) {
				priv1, err := GenerateSecp256k1()
				require.NoError(t, err)
				priv2, err := GenerateSecp256k1()
				require.NoError(t, err)
				return priv1.PubKey(), priv1, priv2.PubKey()
			},
			expectErr: ErrSignatureVerification,
		},
		{
			name:    "Ed25519 wrong public key",
			sigType: SignatureTypeEd25519,
			setupKeys: func(t *testing.T) (correctPub, correctPriv, wrongPub interface{}) {
				pub1, priv1, err := ed25519.GenerateKey(nil)
				require.NoError(t, err)
				pub2, _, err := ed25519.GenerateKey(nil)
				require.NoError(t, err)
				return pub1, priv1, pub2
			},
			expectErr: ErrSignatureVerification,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, correctPriv, wrongPub := tt.setupKeys(t)
			message := []byte("test message")

			sig, err := Sign(tt.sigType, correctPriv, message)
			require.NoError(t, err)

			err = Verify(tt.sigType, wrongPub, message, sig)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}

func TestVerify_InvalidInputs(t *testing.T) {
	tests := []struct {
		name      string
		sigType   SignatureType
		pubKey    interface{}
		message   []byte
		signature []byte
		expectErr error
	}{
		{
			name:      "Invalid signature type",
			sigType:   SignatureType(99),
			pubKey:    []byte("any"),
			message:   []byte("any"),
			signature: []byte("any"),
			expectErr: ErrUnsupportedSignatureType,
		},
		{
			name:      "ECDSA invalid public key type",
			sigType:   SignatureTypeECDSA256K,
			pubKey:    "invalid type",
			message:   []byte("any"),
			signature: []byte("any"),
			expectErr: ErrUnsupportedECDSAPrivKeyType,
		},
		{
			name:      "Ed25519 invalid public key type",
			sigType:   SignatureTypeEd25519,
			pubKey:    "invalid type",
			message:   []byte("any"),
			signature: []byte("any"),
			expectErr: ErrUnsupportedEd25519PrivKeyType,
		},
		{
			name:      "Ed25519 wrong public key length",
			sigType:   SignatureTypeEd25519,
			pubKey:    []byte("wrong length"),
			message:   []byte("any"),
			signature: []byte("any"),
			expectErr: ErrInvalidEd25519PrivKeyLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.sigType, tt.pubKey, tt.message, tt.signature)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}
