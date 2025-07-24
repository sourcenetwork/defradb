// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

import (
	"fmt"

	"github.com/sourcenetwork/defradb/crypto"
)

const (
	KeyTypeEd25519   = "ed25519"
	KeyTypeSecp256k1 = "secp256k1"
)

func CryptoGenerateKey(keyTypeStr string) GoCResult {
	var keyType crypto.KeyType
	switch keyTypeStr {
	case KeyTypeEd25519:
		keyType = crypto.KeyTypeEd25519
	case KeyTypeSecp256k1:
		keyType = crypto.KeyTypeSecp256k1
	default:
		return returnGoC(1, fmt.Sprintf(errInvalidKeyType, keyTypeStr), "")
	}
	key, err := crypto.GenerateKey(keyType)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", key.String())
}
