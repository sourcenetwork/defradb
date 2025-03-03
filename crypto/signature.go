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
func Sign(sigType SignatureType, privKey any, message []byte) ([]byte, error) {
	switch sigType {
	case SignatureTypeECDSA256K:
		var privateKey *secp256k1.PrivateKey
		switch k := privKey.(type) {
		case *secp256k1.PrivateKey:
			privateKey = k
		case []byte:
			privateKey = secp256k1.PrivKeyFromBytes(k)
			if privateKey == nil {
				return nil, ErrInvalidECDSAPrivKeyBytes
			}
		default:
			return nil, ErrUnsupportedECDSAPrivKeyType
		}

		// Hash the message with SHA256
		hash := sha256.Sum256(message)

		// Sign the hash
		signature := ecdsa.Sign(privateKey, hash[:])
		return signature.Serialize(), nil

	case SignatureTypeEd25519:
		switch k := privKey.(type) {
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

	default:
		return nil, ErrUnsupportedSignatureType
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
func Verify(sigType SignatureType, pubKey any, message, signature []byte) error {
	switch sigType {
	case SignatureTypeECDSA256K:
		var publicKey *secp256k1.PublicKey
		switch k := pubKey.(type) {
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

	case SignatureTypeEd25519:
		switch k := pubKey.(type) {
		case ed25519.PublicKey:
			if !ed25519.Verify(k, message, signature) {
				return ErrSignatureVerification
			}
			return nil
		case []byte:
			if len(k) != ed25519.PublicKeySize {
				return ErrInvalidEd25519PrivKeyLength
			}
			if !ed25519.Verify(ed25519.PublicKey(k), message, signature) {
				return ErrSignatureVerification
			}
			return nil
		default:
			return ErrUnsupportedEd25519PrivKeyType
		}

	default:
		return ErrUnsupportedSignatureType
	}
}
