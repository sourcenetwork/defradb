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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

func IdentityNew(keyTypeStr string) GoCResult {
	// Default key type, if left blank, is Secp256k1
	keyType := crypto.KeyTypeSecp256k1
	if keyTypeStr != "" {
		keyType = crypto.KeyType(keyTypeStr)
	}
	newIdentity, err := identity.Generate(crypto.KeyType(keyType))
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(newIdentity.IntoRawIdentity())
}

func NodeIdentity() GoCResult {
	ctx := context.Background()
	identity, err := globalNode.DB.GetNodeIdentity(ctx)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	if identity.HasValue() {
		return marshalJSONToGoCResult(identity.Value())
	}
	return returnGoC(0, "", "Node has no identity assigned to it.")
}
