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

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export VerifyBlockSignature
func VerifyBlockSignature(nodePtr C.uintptr_t,
	keyType *C.char,
	publicKey *C.char,
	cid *C.char,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	keyTypeStr := C.GoString(keyType)
	pubKeyStr := C.GoString(publicKey)
	cryptoKeyType := crypto.KeyTypeSecp256k1
	if keyTypeStr != "" {
		cryptoKeyType = crypto.KeyType(keyTypeStr)
	}
	pubKey, err := crypto.PublicKeyFromString(cryptoKeyType, pubKeyStr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	verifyOpt := options.WithIdentity(options.VerifySignature(), iIdentity.FromContext(ctx))
	err = store.VerifySignature(ctx, C.GoString(cid), pubKey, verifyOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", "Block's signature verified."))
}
