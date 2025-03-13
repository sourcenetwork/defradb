// Copyright 2024 Democratized Data Foundation
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
	"encoding/hex"

	"github.com/cyware/ssi-sdk/crypto"
	"github.com/cyware/ssi-sdk/did/key"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// KeyType represents the type of cryptographic key
type KeyType string

const (
	// KeyTypeSecp256k1 represents a secp256k1 key
	KeyTypeSecp256k1 KeyType = "secp256k1"
	// KeyTypeEd25519 represents an Ed25519 key
	KeyTypeEd25519 KeyType = "ed25519"
)

// Key represents a cryptographic key
type Key interface {
	// Equals checks whether two keys are the same
	Equals(Key) bool
	// Raw returns the raw bytes of the key
	Raw() ([]byte, error)
	// String returns a string representation of the key
	String() string
	// Type returns the key type
	Type() KeyType
	// Underlying returns the underlying key implementation
	Underlying() any
}

// PublicKey represents a public key
type PublicKey interface {
	Key
	// Verify verifies a signature against a message.
	// Returns true if the signature is valid, false otherwise.
	Verify(data []byte, sig []byte) (bool, error)
	// DID returns the DID key representation
	DID() (string, error)
}

// PrivateKey represents a private key
type PrivateKey interface {
	Key
	// Sign signs a message
	Sign([]byte) ([]byte, error)
	// GetPublic returns the corresponding public key
	GetPublic() PublicKey
}

// secp256k1PrivateKey wraps secp256k1.PrivateKey to implement PrivateKey interface
type secp256k1PrivateKey struct {
	key *secp256k1.PrivateKey
}

// String implements PrivateKey.
func (k *secp256k1PrivateKey) String() string {
	return hex.EncodeToString(k.key.Serialize())
}

// secp256k1PublicKey wraps secp256k1.PublicKey to implement PublicKey interface
type secp256k1PublicKey struct {
	key *secp256k1.PublicKey
}

// ed25519PrivateKey wraps ed25519.PrivateKey to implement PrivateKey interface
type ed25519PrivateKey struct {
	key ed25519.PrivateKey
}

// String implements PrivateKey.
func (k *ed25519PrivateKey) String() string {
	return hex.EncodeToString(k.key)
}

// ed25519PublicKey wraps ed25519.PublicKey to implement PublicKey interface
type ed25519PublicKey struct {
	key ed25519.PublicKey
}

// NewPrivateKey creates a new private key of a specific type based on the input
func NewPrivateKey[T *secp256k1.PrivateKey | ed25519.PrivateKey](key T) PrivateKey {
	switch k := any(key).(type) {
	case *secp256k1.PrivateKey:
		if k == nil {
			return nil
		}
		return &secp256k1PrivateKey{key: k}
	case ed25519.PrivateKey:
		if k == nil || len(k) != ed25519.PrivateKeySize {
			return nil
		}
		return &ed25519PrivateKey{key: k}
	default:
		return nil
	}
}

// NewPublicKey creates a new public key of a specific type based on the input
func NewPublicKey[T *secp256k1.PublicKey | ed25519.PublicKey](key T) PublicKey {
	switch k := any(key).(type) {
	case *secp256k1.PublicKey:
		if k == nil {
			return nil
		}
		return &secp256k1PublicKey{key: k}
	case ed25519.PublicKey:
		if k == nil || len(k) != ed25519.PublicKeySize {
			return nil
		}
		return &ed25519PublicKey{key: k}
	default:
		return nil
	}
}

func (k *secp256k1PrivateKey) Equals(other Key) bool {
	if other.Type() != KeyTypeSecp256k1 {
		return false
	}
	otherBytes, err := other.Raw()
	if err != nil {
		return false
	}
	myBytes := k.key.Serialize()
	return string(myBytes) == string(otherBytes)
}

func (k *secp256k1PrivateKey) Raw() ([]byte, error) {
	return k.key.Serialize(), nil
}

func (k *secp256k1PrivateKey) Type() KeyType {
	return KeyTypeSecp256k1
}

func (k *secp256k1PrivateKey) Sign(data []byte) ([]byte, error) {
	return SignECDSA256K(k.key, data)
}

func (k *secp256k1PrivateKey) GetPublic() PublicKey {
	return &secp256k1PublicKey{key: k.key.PubKey()}
}

func (k *secp256k1PrivateKey) Underlying() any {
	return k.key
}

func (k *secp256k1PublicKey) Equals(other Key) bool {
	if other.Type() != KeyTypeSecp256k1 {
		return false
	}
	otherBytes, err := other.Raw()
	if err != nil {
		return false
	}
	myBytes := k.key.SerializeCompressed()
	return string(myBytes) == string(otherBytes)
}

func (k *secp256k1PublicKey) Raw() ([]byte, error) {
	return k.key.SerializeCompressed(), nil
}

func (k *secp256k1PublicKey) Type() KeyType {
	return KeyTypeSecp256k1
}

func (k *secp256k1PublicKey) Verify(data []byte, sig []byte) (bool, error) {
	parsedSig, err := ecdsa.ParseDERSignature(sig)
	if err != nil {
		return false, ErrInvalidECDSASignature
	}

	hash := sha256.Sum256(data)
	return parsedSig.Verify(hash[:], k.key), nil
}

func (k *secp256k1PublicKey) String() string {
	return hex.EncodeToString(k.key.SerializeCompressed())
}

func (k *secp256k1PublicKey) DID() (string, error) {
	did, err := key.CreateDIDKey(crypto.SECP256k1, k.key.SerializeUncompressed())
	if err != nil {
		return "", NewErrFailedToCreateDIDKey(err)
	}
	return did.String(), nil
}

func (k *secp256k1PublicKey) Underlying() any {
	return k.key
}

func (k *ed25519PrivateKey) Equals(other Key) bool {
	if other.Type() != KeyTypeEd25519 {
		return false
	}
	otherBytes, err := other.Raw()
	if err != nil {
		return false
	}
	return string(k.key) == string(otherBytes)
}

func (k *ed25519PrivateKey) Raw() ([]byte, error) {
	return k.key, nil
}

func (k *ed25519PrivateKey) Type() KeyType {
	return KeyTypeEd25519
}

func (k *ed25519PrivateKey) Sign(data []byte) ([]byte, error) {
	return SignEd25519(k.key, data)
}

func (k *ed25519PrivateKey) GetPublic() PublicKey {
	return &ed25519PublicKey{key: k.key.Public().(ed25519.PublicKey)}
}

func (k *ed25519PrivateKey) Underlying() any {
	return k.key
}

func (k *ed25519PublicKey) Equals(other Key) bool {
	if other.Type() != KeyTypeEd25519 {
		return false
	}
	otherBytes, err := other.Raw()
	if err != nil {
		return false
	}
	return string(k.key) == string(otherBytes)
}

func (k *ed25519PublicKey) Raw() ([]byte, error) {
	return k.key, nil
}

func (k *ed25519PublicKey) Type() KeyType {
	return KeyTypeEd25519
}

func (k *ed25519PublicKey) Verify(data []byte, sig []byte) (bool, error) {
	return ed25519.Verify(k.key, data, sig), nil
}

func (k *ed25519PublicKey) String() string {
	return hex.EncodeToString(k.key)
}

func (k *ed25519PublicKey) DID() (string, error) {
	did, err := key.CreateDIDKey(crypto.Ed25519, k.key)
	if err != nil {
		return "", NewErrFailedToCreateDIDKey(err)
	}
	return did.String(), nil
}

func (k *ed25519PublicKey) Underlying() any {
	return k.key
}
