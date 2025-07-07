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
	"fmt"

	"github.com/sourcenetwork/defradb/crypto"
)

const (
	KeyTypeEd25519   = "ed25519"
	KeyTypeSecp256k1 = "secp256k1"
)

//export cryptoGenerateKey
func cryptoGenerateKey(cKeyType *C.char) *C.Result {
	keyTypeStr := C.GoString(cKeyType)

	var keyType crypto.KeyType
	switch keyTypeStr {
	case KeyTypeEd25519:
		keyType = crypto.KeyTypeEd25519
	case KeyTypeSecp256k1:
		keyType = crypto.KeyTypeSecp256k1
	default:
		return returnC(1, fmt.Sprintf(cerrInvalidKeyType, keyTypeStr), "")
	}

	key, err := crypto.GenerateKey(keyType)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", key.String())
}
