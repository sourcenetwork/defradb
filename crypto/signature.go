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

// SignatureType represents the type of signature algorithm to use
type SignatureType int

const (
	// SignatureTypeECDSA represents secp256k1 ECDSA signatures
	SignatureTypeECDSA SignatureType = iota
	// SignatureTypeEd25519 represents Ed25519 signatures
	SignatureTypeEd25519
)

// Sign signs the provided message using the specified signature type and private key.
// For ECDSA, it uses secp256k1 curve and returns the signature in R || S format (64 bytes).
// For Ed25519, it uses standard Ed25519 signing.
func Sign(sigType SignatureType, privKey interface{}, message []byte) ([]byte, error) {
	switch sigType {
	case SignatureTypeECDSA:
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
