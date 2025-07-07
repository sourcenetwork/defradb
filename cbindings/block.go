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

	"github.com/sourcenetwork/defradb/crypto"
)

//export blockVerifySignature
func blockVerifySignature(cKeyType *C.char, cPublicKey *C.char, cCID *C.char) *C.Result {
	ctx := context.Background()
	keyTypeStr := C.GoString(cKeyType)
	pubKeyStr := C.GoString(cPublicKey)
	CIDStr := C.GoString(cCID)

	// Create a public key object of the specified type (Secp256k1 by default)
	keyType := crypto.KeyTypeSecp256k1
	if keyTypeStr != "" {
		keyType = crypto.KeyType(keyTypeStr)
	}
	pubKey, err := crypto.PublicKeyFromString(keyType, pubKeyStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Verify the signature, and either return success status, or an error
	err = globalNode.DB.VerifySignature(ctx, CIDStr, pubKey)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "Block's signature verified.")
}
