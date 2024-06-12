package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const nonceLength = 12

// EncryptAES encrypts data using AES-GCM with a provided key.
func EncryptAES(plainText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
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
