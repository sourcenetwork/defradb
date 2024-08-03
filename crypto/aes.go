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
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// EncryptAES encrypts data using AES-GCM with a provided key.
// The nonce is prepended to the cipherText.
func EncryptAES(plainText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce, err := generateNonceFunc()
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	cipherText := aesGCM.Seal(nonce, nonce, plainText, nil)

	return cipherText, nil
}

// DecryptAES decrypts AES-GCM encrypted data with a provided key.
// The nonce is expected to be prepended to the cipherText.
func DecryptAES(cipherText, key []byte) ([]byte, error) {
	if len(cipherText) < AESNonceSize {
		// TODO return typed error
		return nil, fmt.Errorf("cipherText too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := cipherText[:AESNonceSize]
	cipherText = cipherText[AESNonceSize:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
