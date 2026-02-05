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
extern Result RegisterIdentity(char* did, char* privateKeyHex, char* publicKeyHex, char* keyType);
*/
import "C"

import (
	"context"
	"encoding/hex"
	"runtime/cgo"
	"unsafe"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

// RegisterIdentityWithRust registers an identity's signing config with Rust's global identity store.
// This allows Go-created identities to be used for block signing in Rust.
func RegisterIdentityWithRust(ident identity.Identity) error {
	if ident == nil {
		return nil
	}

	// Only full identities have private keys and can be used for signing
	fullIdent, ok := ident.(identity.FullIdentity)
	if !ok {
		// Not a full identity, can't register for signing (no private key)
		return nil
	}

	did := ident.DID()
	privateKeyHex := hex.EncodeToString(fullIdent.PrivateKey().Raw())
	publicKeyHex := hex.EncodeToString(ident.PublicKey().Raw())
	keyType := string(fullIdent.PrivateKey().Type())

	cDid := C.CString(did)
	cPrivateKeyHex := C.CString(privateKeyHex)
	cPublicKeyHex := C.CString(publicKeyHex)
	cKeyType := C.CString(keyType)
	defer C.free(unsafe.Pointer(cDid))
	defer C.free(unsafe.Pointer(cPrivateKeyHex))
	defer C.free(unsafe.Pointer(cPublicKeyHex))
	defer C.free(unsafe.Pointer(cKeyType))

	result := C.RegisterIdentity(cDid, cPrivateKeyHex, cPublicKeyHex, cKeyType)
	res := ConvertAndFreeCResult(result)
	if res.Status != 0 {
		return NewErrFFI(res.Error)
	}
	return nil
}

//export IdentityNew
func IdentityNew(keyType *C.char) C.NewIdentityResult {
	// Default key type, if left blank, is Secp256k1
	cryptoKeyType := crypto.KeyTypeSecp256k1
	keyTypeStr := C.GoString(keyType)
	if keyTypeStr != "" {
		cryptoKeyType = crypto.KeyType(keyTypeStr)
	}
	newIdentity, err := identity.Generate(cryptoKeyType)
	if err != nil {
		return returnNewIdentityResultC(1, err.Error(), nil)
	}

	// Register the identity with Rust so block signing works
	if err := RegisterIdentityWithRust(newIdentity); err != nil {
		return returnNewIdentityResultC(1, "failed to register identity with Rust: "+err.Error(), nil)
	}

	return returnNewIdentityResultC(0, "", newIdentity)
}

//export NodeIdentity
func NodeIdentity(nodePtr C.uintptr_t) C.Result {
	ctx := context.Background()
	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	identity, err := store.GetNodeIdentity(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	if identity.HasValue() {
		return returnC(marshalJSONToGoCResult(identity.Value()))
	}
	return returnC(returnGoC(0, "", "Node has no identity assigned to it."))
}

//export IdentityFree
func IdentityFree(identityPtr C.uintptr_t) {
	_, err := getIdentityFromPointer(identityPtr)
	if err == nil && identityPtr != 0 {
		cgo.Handle(identityPtr).Delete()
	}
}
