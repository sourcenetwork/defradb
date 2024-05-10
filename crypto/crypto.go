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
)

// GenerateAES256 generates a new random AES-256 bit key.
func GenerateAES256() ([]byte, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	return data, err
}

// GenerateEd25519 generates a new random Ed25519 private key.
func GenerateEd25519() (ed25519.PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	return priv, err
}
