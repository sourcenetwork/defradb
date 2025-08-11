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
	"context"

	"github.com/sourcenetwork/defradb/crypto"
)

func BlockVerifySignature(n int, keyTypeStr string, pubKeyStr string, CIDStr string) GoCResult {
	ctx := context.Background()
	keyType := crypto.KeyTypeSecp256k1
	if keyTypeStr != "" {
		keyType = crypto.KeyType(keyTypeStr)
	}
	pubKey, err := crypto.PublicKeyFromString(keyType, pubKeyStr)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = GetNode(n).DB.VerifySignature(ctx, CIDStr, pubKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "Block's signature verified.")
}
