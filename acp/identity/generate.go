// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package identity

import (
	"crypto/ed25519"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/sourcenetwork/defradb/crypto"
)

// Generate generates a new identity with a secp256k1 key pair.
// This is kept for backward compatibility.
func Generate() (RawIdentity, error) {
	return GenerateWithType(crypto.KeyTypeSecp256k1)
}

// GenerateWithType generates a new identity with the specified key type.
// Supported types are KeyTypeSecp256k1 and KeyTypeEd25519.
func GenerateWithType(keyType crypto.KeyType) (RawIdentity, error) {
	var privKey crypto.PrivateKey
	var err error

	switch keyType {
	case crypto.KeyTypeSecp256k1:
		var key *secp256k1.PrivateKey
		key, err = crypto.GenerateSecp256k1()
		if err != nil {
			return RawIdentity{}, err
		}
		privKey = crypto.NewPrivateKey(key)
	case crypto.KeyTypeEd25519:
		var key ed25519.PrivateKey
		key, err = crypto.GenerateEd25519()
		if err != nil {
			return RawIdentity{}, err
		}
		privKey = crypto.NewPrivateKey(key)
	default:
		return RawIdentity{}, fmt.Errorf("unsupported key type: %s", keyType)
	}

	identity, err := FromPrivateKey(privKey)
	if err != nil {
		return RawIdentity{}, err
	}

	return identity.IntoRawIdentity(), nil
}
