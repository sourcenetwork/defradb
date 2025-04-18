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
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToGenerateEphemeralKey    string = "failed to generate ephemeral key"
	errFailedECDHOperation             string = "failed ECDH operation"
	errFailedKDFOperationForAESKey     string = "failed KDF operation for AES key"
	errFailedKDFOperationForHMACKey    string = "failed KDF operation for HMAC key"
	errFailedToEncrypt                 string = "failed to encrypt"
	errCipherTextTooShort              string = "cipherText too short"
	errFailedToParseEphemeralPublicKey string = "failed to parse ephemeral public key"
	errVerificationWithHMACFailed      string = "verification with HMAC failed"
	errFailedToDecrypt                 string = "failed to decrypt"
	errNoPublicKeyForDecryption        string = "no public key provided for decryption"
	errInvalidECDSAPrivKeyBytes        string = "invalid ECDSA private key bytes"
	errNilKey                          string = "received nil key"
	errInvalidECDSASignature           string = "invalid ECDSA signature"
	errInvalidECDSAPubKey              string = "invalid secp256k1 public key"
	errInvalidEd25519PrivKeyLength     string = "invalid Ed25519 private key length"
	errInvalidEd25519PubKeyLength      string = "invalid Ed25519 public key length"
	errInvalidEd25519PubKey            string = "invalid Ed25519 public key"
	errSignatureVerification           string = "signature verification failed"
	errUnsupportedPrivKeyType          string = "unsupported private key type"
	errUnsupportedPubKeyType           string = "unsupported public key type"
	errFailedToCreateDIDKey            string = "failed to create DID key"
	errUnsupportedKeyType              string = "unsupported key type"
)

var (
	ErrCipherTextTooShort          = errors.New(errCipherTextTooShort)
	ErrVerificationWithHMACFailed  = errors.New(errVerificationWithHMACFailed)
	ErrNoPublicKeyForDecryption    = errors.New(errNoPublicKeyForDecryption)
	ErrInvalidECDSAPrivKeyBytes    = errors.New(errInvalidECDSAPrivKeyBytes)
	ErrNilKey                      = errors.New(errNilKey)
	ErrInvalidECDSASignature       = errors.New(errInvalidECDSASignature)
	ErrInvalidECDSAPubKey          = errors.New(errInvalidECDSAPubKey)
	ErrInvalidEd25519PrivKeyLength = errors.New(errInvalidEd25519PrivKeyLength)
	ErrInvalidEd25519PubKeyLength  = errors.New(errInvalidEd25519PubKeyLength)
	ErrInvalidEd25519PubKey        = errors.New(errInvalidEd25519PubKey)
	ErrUnsupportedPrivKeyType      = errors.New(errUnsupportedPrivKeyType)
	ErrUnsupportedPubKeyType       = errors.New(errUnsupportedPubKeyType)
	ErrSignatureVerification       = errors.New(errSignatureVerification)
)

func NewErrFailedToGenerateEphemeralKey(inner error) error {
	return errors.Wrap(errFailedToGenerateEphemeralKey, inner)
}

func NewErrFailedECDHOperation(inner error) error {
	return errors.Wrap(errFailedECDHOperation, inner)
}

func NewErrFailedKDFOperationForAESKey(inner error) error {
	return errors.Wrap(errFailedKDFOperationForAESKey, inner)
}

func NewErrFailedKDFOperationForHMACKey(inner error) error {
	return errors.Wrap(errFailedKDFOperationForHMACKey, inner)
}

func NewErrFailedToEncrypt(inner error) error {
	return errors.Wrap(errFailedToEncrypt, inner)
}

func NewErrFailedToParseEphemeralPublicKey(inner error) error {
	return errors.Wrap(errFailedToParseEphemeralPublicKey, inner)
}

func NewErrFailedToDecrypt(inner error) error {
	return errors.Wrap(errFailedToDecrypt, inner)
}

func NewErrFailedToCreateDIDKey(inner error) error {
	return errors.Wrap(errFailedToCreateDIDKey, inner)
}

func NewErrUnsupportedKeyType(keyType KeyType) error {
	return errors.New(errUnsupportedKeyType, errors.NewKV("KeyType", keyType))
}
