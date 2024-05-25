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
	"crypto/rand"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// GenerateSecp256k1 generates a new secp256k1 private key.
func GenerateSecp256k1() (*secp256k1.PrivateKey, error) {
	return secp256k1.GeneratePrivateKey()
}

// GenerateAES256 generates a new random AES-256 bit key.
func GenerateAES256() ([]byte, error) {
	return RandomBytes(32)
}

// GenerateEd25519 generates a new random Ed25519 private key.
func GenerateEd25519() (ed25519.PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	return priv, err
}

// RandomBytes returns a random slice of bytes of the given size.
func RandomBytes(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data)
	return data, err
}
