// Copyright 2025 Democratized Data Foundation
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
	"crypto/ed25519"
	"crypto/sha256"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Sign signs the provided message using the appropriate signature algorithm based on the key type.
//
// This is a generic function that accepts different types of private keys and
// automatically uses the correct signing algorithm based on the key type.
//
// For ECDSA (secp256k1):
// - Returns signature in DER format
// - Accepts private key as *secp256k1.PrivateKey
//
// For Ed25519:
// - Returns standard Ed25519 signature
// - Accepts private key as ed25519.PrivateKey
//
// Parameters:
//   - privKey: The private key to sign with
//   - message: The message to sign
//
// Returns:
//   - []byte: The signature in the format appropriate for the key type
//   - error: Any error encountered during signing, including invalid key types
func Sign[T ed25519.PrivateKey | *secp256k1.PrivateKey](
	privKey T,
	message []byte,
) ([]byte, error) {
	switch k := any(privKey).(type) {
	case *secp256k1.PrivateKey:
		return SignECDSA256K(k, message)
	case ed25519.PrivateKey:
		return SignEd25519(k, message)
	default:
		// This should never happen due to type constraints on T
		return nil, ErrUnsupportedPrivKeyType
	}
}

// SignECDSA256K signs a message using ECDSA with the secp256k1 curve.
//
// Returns signature in DER format.
// Accepts private key as *secp256k1.PrivateKey: Direct private key object
//
// Parameters:
//   - privKey: The ECDSA private key to sign with
//   - message: The message to sign
//
// Returns:
//   - []byte: The DER-encoded signature
//   - error: Any error encountered during signing
func SignECDSA256K(
	privKey *secp256k1.PrivateKey,
	message []byte,
) ([]byte, error) {
	if privKey == nil {
		return nil, ErrInvalidECDSAPrivKeyBytes
	}

	hash := sha256.Sum256(message)
	signature := ecdsa.Sign(privKey, hash[:])
	return signature.Serialize(), nil
}

// SignEd25519 signs a message using Ed25519.
//
// Returns standard Ed25519 signature.
// Accepts private key as ed25519.PrivateKey: Direct private key object
//
// Parameters:
//   - privKey: The Ed25519 private key to sign with
//   - message: The message to sign
//
// Returns:
//   - []byte: The Ed25519 signature
//   - error: Any error encountered during signing
func SignEd25519(
	privKey ed25519.PrivateKey,
	message []byte,
) ([]byte, error) {
	if privKey == nil || len(privKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidEd25519PrivKeyLength
	}
	return ed25519.Sign(privKey, message), nil
}

// Verify verifies a signature against a message using the appropriate signature algorithm based on the key type.
//
// For ECDSA (secp256k1):
// - Expects signature in DER format
// - Accepts public key as *secp256k1.PublicKey
//
// For Ed25519:
// - Expects standard Ed25519 signature
// - Accepts public key as ed25519.PublicKey
//
// Parameters:
//   - pubKey: The public key to verify with
//   - message: The original message that was signed
//   - signature: The signature to verify
//
// Returns:
//   - error: nil if verification succeeds, appropriate error otherwise
func Verify[T *secp256k1.PublicKey | ed25519.PublicKey](
	pubKey T,
	message []byte,
	signature []byte,
) error {
	switch k := any(pubKey).(type) {
	case *secp256k1.PublicKey:
		return VerifyECDSA256K(k, message, signature)
	case ed25519.PublicKey:
		return VerifyEd25519(k, message, signature)
	default:
		// This should never happen due to type constraints on T
		return ErrUnsupportedPubKeyType
	}
}

// VerifyECDSA256K verifies a signature against a message using ECDSA with the secp256k1 curve.
//
// Expects signature in DER format.
// Accepts public key as *secp256k1.PublicKey: Direct public key object
//
// Parameters:
//   - pubKey: The ECDSA public key to verify with
//   - message: The original message that was signed
//   - signature: The DER-encoded signature to verify
//
// Returns:
//   - error: nil if verification succeeds, appropriate error otherwise
func VerifyECDSA256K(
	pubKey *secp256k1.PublicKey,
	message []byte,
	signature []byte,
) error {
	if pubKey == nil {
		return ErrUnsupportedECDSAPrivKeyType
	}

	sig, err := ecdsa.ParseDERSignature(signature)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(message)
	if !sig.Verify(hash[:], pubKey) {
		return ErrSignatureVerification
	}

	return nil
}

// VerifyEd25519 verifies a signature against a message using Ed25519.
//
// Expects standard Ed25519 signature.
// Accepts public key as ed25519.PublicKey: Direct public key object
//
// Parameters:
//   - pubKey: The Ed25519 public key to verify with
//   - message: The original message that was signed
//   - signature: The Ed25519 signature to verify
//
// Returns:
//   - error: nil if verification succeeds, appropriate error otherwise
func VerifyEd25519(
	pubKey ed25519.PublicKey,
	message []byte,
	signature []byte,
) error {
	if pubKey == nil || len(pubKey) != ed25519.PublicKeySize {
		return ErrInvalidEd25519PrivKeyLength
	}

	if !ed25519.Verify(pubKey, message, signature) {
		return ErrSignatureVerification
	}

	return nil
}
