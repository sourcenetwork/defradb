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
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

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
	// Equal checks whether two keys are the same
	Equal(Key) bool
	// Raw returns the raw bytes of the key
	Raw() []byte
	// String returns a string representation of the key
	String() string
	// Type returns the key type
	Type() KeyType
	// Underlying returns the underlying cryptographic key
	// For example [*secp256k1.PrivateKey] or [ed25519.PrivateKey]
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

// PublicKeyFromString creates a public key from a hex-encoded string and key type.
// This is useful for deserializing public keys.
func PublicKeyFromString(keyType KeyType, keyString string) (PublicKey, error) {
	keyBytes, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, NewErrFailedToParseEphemeralPublicKey(err)
	}

	switch keyType {
	case KeyTypeSecp256k1:
		pubKey, err := secp256k1.ParsePubKey(keyBytes)
		if err != nil {
			return nil, ErrInvalidECDSAPubKey
		}
		return &secp256k1PublicKey{key: pubKey}, nil

	case KeyTypeEd25519:
		if len(keyBytes) != ed25519.PublicKeySize {
			return nil, ErrInvalidEd25519PubKeyLength
		}
		return &ed25519PublicKey{key: keyBytes}, nil

	default:
		return nil, NewErrUnsupportedKeyType(keyType)
	}
}

func (k *secp256k1PrivateKey) Equal(other Key) bool {
	if other.Type() != KeyTypeSecp256k1 {
		return false
	}
	return bytes.Equal(k.key.Serialize(), other.Raw())
}

func (k *secp256k1PrivateKey) Raw() []byte {
	return k.key.Serialize()
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

func (k *secp256k1PublicKey) Equal(other Key) bool {
	if other.Type() != KeyTypeSecp256k1 {
		return false
	}
	return bytes.Equal(k.key.SerializeCompressed(), other.Raw())
}

func (k *secp256k1PublicKey) Raw() []byte {
	return k.key.SerializeCompressed()
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
	did, err := createDIDKey(SECP256k1, k.key.SerializeUncompressed())
	if err != nil {
		return "", NewErrFailedToCreateDIDKey(err)
	}
	return did.String(), nil
}

func (k *secp256k1PublicKey) Underlying() any {
	return k.key
}

func (k *ed25519PrivateKey) Equal(other Key) bool {
	if other.Type() != KeyTypeEd25519 {
		return false
	}
	return bytes.Equal(k.key, other.Raw())
}

func (k *ed25519PrivateKey) Raw() []byte {
	return k.key
}

func (k *ed25519PrivateKey) Type() KeyType {
	return KeyTypeEd25519
}

func (k *ed25519PrivateKey) Sign(data []byte) ([]byte, error) {
	return SignEd25519(k.key, data)
}

func (k *ed25519PrivateKey) GetPublic() PublicKey {
	//nolint:forcetypeassert
	return &ed25519PublicKey{key: k.key.Public().(ed25519.PublicKey)}
}

func (k *ed25519PrivateKey) Underlying() any {
	return k.key
}

func (k *ed25519PublicKey) Equal(other Key) bool {
	if other.Type() != KeyTypeEd25519 {
		return false
	}
	return bytes.Equal(k.key, other.Raw())
}

func (k *ed25519PublicKey) Raw() []byte {
	return k.key
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
	did, err := createDIDKey(Ed25519, k.key)
	if err != nil {
		return "", NewErrFailedToCreateDIDKey(err)
	}
	return did.String(), nil
}

func (k *ed25519PublicKey) Underlying() any {
	return k.key
}

// PrivateKeyFromBytes creates a private key from raw bytes and key type.
// This is useful for deserializing private keys.
func PrivateKeyFromBytes(keyType KeyType, keyBytes []byte) (PrivateKey, error) {
	switch keyType {
	case KeyTypeSecp256k1:
		if len(keyBytes) != secp256k1.PrivKeyBytesLen {
			return nil, ErrInvalidECDSAPrivKeyBytes
		}
		privKey := secp256k1.PrivKeyFromBytes(keyBytes)
		return &secp256k1PrivateKey{key: privKey}, nil

	case KeyTypeEd25519:
		if len(keyBytes) != ed25519.PrivateKeySize {
			return nil, ErrInvalidEd25519PrivKeyLength
		}
		return &ed25519PrivateKey{key: keyBytes}, nil

	default:
		return nil, NewErrUnsupportedKeyType(keyType)
	}
}

// PrivateKeyFromString creates a private key from a hex-encoded string and key type.
// This is useful for deserializing private keys from string representation.
func PrivateKeyFromString(keyType KeyType, keyString string) (PrivateKey, error) {
	keyBytes, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	return PrivateKeyFromBytes(keyType, keyBytes)
}

// GenerateKey generates a new private key of the given type.
func GenerateKey(keyType KeyType) (PrivateKey, error) {
	switch keyType {
	case KeyTypeSecp256k1:
		key, err := GenerateSecp256k1()
		if err != nil {
			return nil, err
		}
		return NewPrivateKey(key), nil
	case KeyTypeEd25519:
		key, err := GenerateEd25519()
		if err != nil {
			return nil, err
		}
		return NewPrivateKey(key), nil
	default:
		return nil, NewErrUnsupportedKeyType(keyType)
	}
}

// GenerateSecp256k1 generates a new secp256k1 private key.
func GenerateSecp256k1() (*secp256k1.PrivateKey, error) {
	return secp256k1.GeneratePrivateKey()
}

// GenerateEd25519 generates a new random Ed25519 private key.
func GenerateEd25519() (ed25519.PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	return priv, err
}
