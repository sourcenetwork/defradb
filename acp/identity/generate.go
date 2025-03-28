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
	"github.com/sourcenetwork/defradb/crypto"
)

// Generate generates a new identity with the specified key type.
// Supported types are KeyTypeSecp256k1 and KeyTypeEd25519.
func Generate(keyType crypto.KeyType) (Identity, error) {
	privKey, err := crypto.GenerateKey(keyType)
	if err != nil {
		return Identity{}, err
	}

	identity, err := FromPrivateKey(privKey)
	if err != nil {
		return Identity{}, err
	}

	return identity, nil
}
