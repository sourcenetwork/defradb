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

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"context"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

//export IdentityNew
func IdentityNew(keyType *C.char) C.NewIdentityResult {
	// Default key type, if left blank, is Secp256k1
	cryptoKeyType := crypto.KeyTypeSecp256k1
	keyTypeStr := C.GoString(keyType)
	if keyTypeStr != "" {
		cryptoKeyType = crypto.KeyType(keyTypeStr)
	}
	newIdentity, err := identity.Generate(crypto.KeyType(cryptoKeyType))
	if err != nil {
		return returnNewIdentityResultC(1, err.Error(), nil)
	}
	return returnNewIdentityResultC(0, "", newIdentity)
}

//export NodeIdentity
func NodeIdentity(nodePtr C.uintptr_t) *C.Result {
	ctx := context.Background()
	store := getStoreFromPointer(nodePtr)
	identity, err := store.GetNodeIdentity(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	if identity.HasValue() {
		return returnC(marshalJSONToGoCResult(identity.Value()))
	}
	return returnC(returnGoC(0, "", "Node has no identity assigned to it."))
}
