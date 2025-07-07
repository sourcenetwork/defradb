// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build cgo
// +build cgo

package main

/*
#include "defra_structs.h"
*/
import "C"

import (
	"context"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

//export identityNew
func identityNew(cKeyType *C.char) *C.Result {
	keyTypeStr := C.GoString(cKeyType)

	// Create a public key object of the specified type (Secp256k1 by default) and use it to create identity
	keyType := crypto.KeyTypeSecp256k1
	if keyTypeStr != "" {
		keyType = crypto.KeyType(keyTypeStr)
	}
	newIdentity, err := identity.Generate(crypto.KeyType(keyType))
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(newIdentity.IntoRawIdentity())
}

//export nodeIdentity
func nodeIdentity() *C.Result {
	ctx := context.Background()
	identity, err := globalNode.DB.GetNodeIdentity(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	if identity.HasValue() {
		return marshalJSONToCResult(identity.Value())
	}
	return returnC(0, "", "Node has no identity assigned to it.")
}
