// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// EncryptAES encrypts data using AES-GCM with a provided key.
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

	buf := make([]byte, base64.StdEncoding.EncodedLen(len(cipherText)))
	base64.StdEncoding.Encode(buf, cipherText)

	return buf, nil
}

// DecryptAES decrypts AES-GCM encrypted data with a provided key.
func DecryptAES(cipherTextBase64, key []byte) ([]byte, error) {
	cipherText := make([]byte, base64.StdEncoding.DecodedLen(len(cipherTextBase64)))
	n, err := base64.StdEncoding.Decode(cipherText, []byte(cipherTextBase64))

	if err != nil {
		return nil, err
	}

	cipherText = cipherText[:n]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < nonceLength {
		return nil, fmt.Errorf("cipherText too short")
	}

	nonce := cipherText[:nonceLength]
	cipherText = cipherText[nonceLength:]

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
