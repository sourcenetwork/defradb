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

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// SignatureType represents the type of signature algorithm to use
type SignatureType int

const (
	// SignatureTypeECDSA256K represents secp256k1 ECDSA signatures using secp256k1 curve
	SignatureTypeECDSA256K SignatureType = iota
	// SignatureTypeEd25519 represents Ed25519 signatures
	SignatureTypeEd25519
)

// Sign signs the provided message using the specified signature type and private key.
// This is a generic function that accepts various types of private keys.
//
// For ECDSA (secp256k1):
// - Returns signature in DER format
// - Accepts private key as either:
//   - *secp256k1.PrivateKey: Direct private key object
//   - []byte: Raw private key bytes that will be parsed into secp256k1.PrivateKey
//
// For Ed25519:
// - Returns standard Ed25519 signature
// - Accepts private key as either:
//   - ed25519.PrivateKey: Direct private key object
//   - []byte: Raw private key bytes (must be ed25519.PrivateKeySize bytes)
//
// Parameters:
//   - sigType: The type of signature algorithm to use (ECDSA or Ed25519)
//   - privKey: The private key to sign with (see above for accepted types)
//   - message: The message to sign
//
// Returns:
//   - []byte: The signature in the format appropriate for the chosen algorithm
//   - error: Any error encountered during signing, including invalid key types
func Sign[T ed25519.PrivateKey | *secp256k1.PrivateKey | []byte](
	sigType SignatureType,
	privKey T,
	message []byte,
) ([]byte, error) {
	switch sigType {
	case SignatureTypeECDSA256K:
		// Type assertion to ensure we're passing a compatible key type
		switch k := any(privKey).(type) {
		case *secp256k1.PrivateKey:
			return SignECDSA256K(k, message)
		case []byte:
			return SignECDSA256K(k, message)
		}
	case SignatureTypeEd25519:
		// Type assertion to ensure we're passing a compatible key type
		switch k := any(privKey).(type) {
		case ed25519.PrivateKey:
			return SignEd25519(k, message)
		case []byte:
			return SignEd25519(k, message)
		}
	}
	return nil, ErrUnsupportedSignatureType
}

// SignECDSA256K signs a message using ECDSA with the secp256k1 curve.
//
// Returns signature in DER format.
// Accepts private key as either:
// - *secp256k1.PrivateKey: Direct private key object
// - []byte: Raw private key bytes that will be parsed into secp256k1.PrivateKey
//
// Parameters:
//   - privKey: The ECDSA private key to sign with
//   - message: The message to sign
//
// Returns:
//   - []byte: The DER-encoded signature
//   - error: Any error encountered during signing
func SignECDSA256K[T *secp256k1.PrivateKey | []byte](
	privKey T,
	message []byte,
) ([]byte, error) {
	var privateKey *secp256k1.PrivateKey

	switch k := any(privKey).(type) {
	case *secp256k1.PrivateKey:
		privateKey = k
	case []byte:
		if len(k) < 32 {
			return nil, ErrInvalidECDSAPrivKeyBytes
		}
		privateKey = secp256k1.PrivKeyFromBytes(k)
		if privateKey == nil {
			return nil, ErrInvalidECDSAPrivKeyBytes
		}
	default:
		return nil, ErrUnsupportedECDSAPrivKeyType
	}

	hash := sha256.Sum256(message)
	signature := ecdsa.Sign(privateKey, hash[:])
	return signature.Serialize(), nil
}

// SignEd25519 signs a message using Ed25519.
//
// Returns standard Ed25519 signature.
// Accepts private key as either:
// - ed25519.PrivateKey: Direct private key object
// - []byte: Raw private key bytes (must be ed25519.PrivateKeySize bytes)
//
// Parameters:
//   - privKey: The Ed25519 private key to sign with
//   - message: The message to sign
//
// Returns:
//   - []byte: The Ed25519 signature
//   - error: Any error encountered during signing
func SignEd25519[T ed25519.PrivateKey | []byte](
	privKey T,
	message []byte,
) ([]byte, error) {
	switch k := any(privKey).(type) {
	case ed25519.PrivateKey:
		return ed25519.Sign(k, message), nil
	case []byte:
		if len(k) != ed25519.PrivateKeySize {
			return nil, ErrInvalidEd25519PrivKeyLength
		}
		return ed25519.Sign(ed25519.PrivateKey(k), message), nil
	default:
		return nil, ErrUnsupportedEd25519PrivKeyType
	}
}

// Verify verifies a signature against a message using the specified signature algorithm.
//
// For ECDSA (secp256k1):
// - Expects signature in DER format
// - Accepts public key as either:
//   - *secp256k1.PublicKey: Direct public key object
//   - []byte: Raw public key bytes that will be parsed
//
// For Ed25519:
// - Expects standard Ed25519 signature
// - Accepts public key as either:
//   - ed25519.PublicKey: Direct public key object
//   - []byte: Raw public key bytes (must be ed25519.PublicKeySize bytes)
//
// Parameters:
//   - sigType: The type of signature algorithm (ECDSA or Ed25519)
//   - pubKey: The public key to verify with (see above for accepted types)
//   - message: The original message that was signed
//   - signature: The signature to verify
//
// Returns:
//   - error: nil if verification succeeds, appropriate error otherwise
func Verify[T *secp256k1.PublicKey | ed25519.PublicKey | []byte](
	sigType SignatureType,
	pubKey T,
	message []byte,
	signature []byte,
) error {
	switch sigType {
	case SignatureTypeECDSA256K:
		switch k := any(pubKey).(type) {
		case *secp256k1.PublicKey:
			return VerifyECDSA256K(k, message, signature)
		case []byte:
			return VerifyECDSA256K(k, message, signature)
		default:
			return ErrUnsupportedECDSAPrivKeyType
		}
	case SignatureTypeEd25519:
		switch k := any(pubKey).(type) {
		case ed25519.PublicKey:
			return VerifyEd25519(k, message, signature)
		case []byte:
			return VerifyEd25519(k, message, signature)
		default:
			return ErrUnsupportedEd25519PrivKeyType
		}
	default:
		return ErrUnsupportedSignatureType
	}
}

// VerifyECDSA256K verifies a signature against a message using ECDSA with the secp256k1 curve.
//
// Expects signature in DER format.
// Accepts public key as either:
// - *secp256k1.PublicKey: Direct public key object
// - []byte: Raw public key bytes that will be parsed
//
// Parameters:
//   - pubKey: The ECDSA public key to verify with
//   - message: The original message that was signed
//   - signature: The DER-encoded signature to verify
//
// Returns:
//   - error: nil if verification succeeds, appropriate error otherwise
func VerifyECDSA256K[T *secp256k1.PublicKey | []byte](
	pubKey T,
	message []byte,
	signature []byte,
) error {
	var publicKey *secp256k1.PublicKey

	switch k := any(pubKey).(type) {
	case *secp256k1.PublicKey:
		publicKey = k
	case []byte:
		var err error
		publicKey, err = secp256k1.ParsePubKey(k)
		if err != nil {
			return err
		}
	default:
		return ErrUnsupportedECDSAPrivKeyType
	}

	sig, err := ecdsa.ParseDERSignature(signature)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(message)
	if !sig.Verify(hash[:], publicKey) {
		return ErrSignatureVerification
	}

	return nil
}

// VerifyEd25519 verifies a signature against a message using Ed25519.
//
// Expects standard Ed25519 signature.
// Accepts public key as either:
// - ed25519.PublicKey: Direct public key object
// - []byte: Raw public key bytes (must be ed25519.PublicKeySize bytes)
//
// Parameters:
//   - pubKey: The Ed25519 public key to verify with
//   - message: The original message that was signed
//   - signature: The Ed25519 signature to verify
//
// Returns:
//   - error: nil if verification succeeds, appropriate error otherwise
func VerifyEd25519[T ed25519.PublicKey | []byte](
	pubKey T,
	message []byte,
	signature []byte,
) error {
	switch k := any(pubKey).(type) {
	case ed25519.PublicKey:
		if !ed25519.Verify(k, message, signature) {
			return ErrSignatureVerification
		}
	case []byte:
		if len(k) != ed25519.PublicKeySize {
			return ErrInvalidEd25519PrivKeyLength
		}
		if !ed25519.Verify(ed25519.PublicKey(k), message, signature) {
			return ErrSignatureVerification
		}
	default:
		return ErrUnsupportedEd25519PrivKeyType
	}

	return nil
}
