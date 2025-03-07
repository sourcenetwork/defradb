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
	errUnsupportedSignatureType        string = "unsupported signature type"
	errInvalidECDSAPrivKeyBytes        string = "invalid ECDSA private key bytes"
	errUnsupportedECDSAPrivKeyType     string = "unsupported ECDSA private key type"
	errInvalidEd25519PrivKeyLength     string = "invalid Ed25519 private key length"
	errUnsupportedEd25519PrivKeyType   string = "unsupported Ed25519 private key type"
	errSignatureVerification           string = "signature verification failed"
	errUnsupportedPrivKeyType          string = "unsupported private key type"
	errUnsupportedPubKeyType           string = "unsupported public key type"
)

var (
	ErrCipherTextTooShort            = errors.New(errCipherTextTooShort)
	ErrVerificationWithHMACFailed    = errors.New(errVerificationWithHMACFailed)
	ErrNoPublicKeyForDecryption      = errors.New(errNoPublicKeyForDecryption)
	ErrUnsupportedSignatureType      = errors.New(errUnsupportedSignatureType)
	ErrInvalidECDSAPrivKeyBytes      = errors.New(errInvalidECDSAPrivKeyBytes)
	ErrUnsupportedECDSAPrivKeyType   = errors.New(errUnsupportedECDSAPrivKeyType)
	ErrInvalidEd25519PrivKeyLength   = errors.New(errInvalidEd25519PrivKeyLength)
	ErrUnsupportedEd25519PrivKeyType = errors.New(errUnsupportedEd25519PrivKeyType)
	ErrSignatureVerification         = errors.New(errSignatureVerification)
	ErrUnsupportedPrivKeyType        = errors.New(errUnsupportedPrivKeyType)
	ErrUnsupportedPubKeyType         = errors.New(errUnsupportedPubKeyType)
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
