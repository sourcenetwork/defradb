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
	"crypto/rand"
)

// GenerateAES256 generates a new random AES-256 bit key.
func GenerateAES256() ([]byte, error) {
	return randomBytes(32)
}

// randomBytes returns a random slice of bytes of the given size.
func randomBytes(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data)
	return data, err
}

// EncryptAES encrypts data using AES-GCM with a provided key and additional data.
// It generates a nonce internally and optionally prepends it to the cipherText.
//
// Parameters:
//   - plainText: The data to be encrypted
//   - key: The AES encryption key
//   - additionalData: Additional authenticated data (AAD) to be used in the encryption
//   - prependNonce: If true, the nonce is prepended to the returned cipherText
//
// Returns:
//   - cipherText: The encrypted data, with the nonce prepended if prependNonce is true
//   - nonce: The generated nonce
//   - error: Any error encountered during the encryption process
func EncryptAES(plainText, key, additionalData []byte, prependNonce bool) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	nonce, err := generateNonceFunc()
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	var cipherText []byte
	if prependNonce {
		cipherText = aesGCM.Seal(nonce, nonce, plainText, additionalData)
	} else {
		cipherText = aesGCM.Seal(nil, nonce, plainText, additionalData)
	}

	return cipherText, nonce, nil
}

// DecryptAES decrypts AES-GCM encrypted data with a provided key and additional data.
// If no separate nonce is provided, it assumes the nonce is prepended to the cipherText.
//
// Parameters:
//   - nonce: The nonce used for decryption. If empty, it's assumed to be prepended to cipherText
//   - cipherText: The data to be decrypted
//   - key: The AES decryption key
//   - additionalData: Additional authenticated data (AAD) used during encryption
//
// Returns:
//   - plainText: The decrypted data
//   - error: Any error encountered during the decryption process, including authentication failures
func DecryptAES(nonce, cipherText, key, additionalData []byte) ([]byte, error) {
	if len(nonce) == 0 {
		if len(cipherText) < AESNonceSize {
			return nil, ErrCipherTextTooShort
		}
		nonce = cipherText[:AESNonceSize]
		cipherText = cipherText[AESNonceSize:]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := aesGCM.Open(nil, nonce, cipherText, additionalData)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
