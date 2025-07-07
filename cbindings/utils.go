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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

type collectionContextKey struct{}
type schemaNameContextKey struct{}
type identityContextKey struct{}

// Helper function which builds a return struct from Go to C
func returnC(status int, errortext string, valuetext string) *C.Result {
	result := (*C.Result)(C.malloc(C.size_t(unsafe.Sizeof(C.Result{}))))

	result.status = C.int(status)
	result.error = C.CString(errortext)
	result.value = C.CString(valuetext)

	return result
}

// Helper function that attaches an identity to a context, returning the new context
func contextWithIdentity(ctx context.Context, privateKeyHex string) (context.Context, error) {
	if privateKeyHex == "" {
		return ctx, nil
	}
	data, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return ctx, err
	}
	privKey := secp256k1.PrivKeyFromBytes(data)
	newIdentity, err := identity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	if err != nil {
		return ctx, err
	}
	immutableIdentity := immutable.Some[identity.Identity](newIdentity)
	newctx := identity.WithContext(ctx, immutableIdentity)
	return newctx, nil
}

// Helper function that attaches a transaction to a context, returning a new context
func contextWithTransaction(ctx context.Context, cTxnID C.ulonglong) (context.Context, error) {
	TxnIDu64 := uint64(cTxnID)
	if TxnIDu64 == 0 {
		return ctx, nil
	}
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return ctx, fmt.Errorf(cerrTxnDoesNotExist, TxnIDu64)
	}
	txn := tx.(datastore.Txn) //nolint:forcetypeassert
	ctx2 := context.WithValue(ctx, transactionContextKey{}, txn)
	return ctx2, nil
}

// Helper function that seeks to marshall JSON into a CResult
// The Result object will either contain the payload, if it works, or an error if it doesn't
func marshalJSONToCResult(value any) *C.Result {
	dataJSON, err := json.Marshal(value)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}
	return returnC(0, "", string(dataJSON))
}

// Helper function that takes a comma separated const char * and returns an array of Go strings
func splitCommaSeparatedCString(cStr *C.char) []string {
	baseStr := C.GoString(cStr)
	var retArr []string
	if baseStr != "" {
		retArr = strings.Split(baseStr, ",")
	} else {
		retArr = []string{}
	}
	return retArr
}

// Helper function that tries to build a []client.RequestOption from C-string name and JSON variables
func buildRequestOptions(cOpName *C.char, cVars *C.char) ([]client.RequestOption, error) {
	var opts []client.RequestOption
	if cOpName != nil && C.GoString(cOpName) != "" {
		opts = append(opts, client.WithOperationName(C.GoString(cOpName)))
	}
	if cVars != nil && C.GoString(cVars) != "" {
		var variables map[string]any
		if err := json.Unmarshal([]byte(C.GoString(cVars)), &variables); err != nil {
			return nil, fmt.Errorf("invalid JSON in variables: %w", err)
		}
		opts = append(opts, client.WithVariables(variables))
	}
	return opts, nil
}

func loadIdentityFromString(cKeyType *C.char, cPrivKey *C.char) (*identity.FullIdentity, error) {
	goKeyType := C.GoString(cKeyType)
	goPrivKeyStr := C.GoString(cPrivKey)

	if goKeyType == "" || goPrivKeyStr == "" {
		return nil, nil
	}

	// Convert string key type to crypto.KeyType
	var keyType crypto.KeyType
	switch goKeyType {
	case KeyTypeEd25519:
		keyType = crypto.KeyTypeEd25519
	case KeyTypeSecp256k1:
		keyType = crypto.KeyTypeSecp256k1
	default:
		return nil, fmt.Errorf("invalid key type: %s", goKeyType)
	}

	privKey, err := crypto.PrivateKeyFromString(keyType, goPrivKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to construct private key: %w", err)
	}

	id, err := identity.FromPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity from private key: %w", err)
	}

	return &id, nil
}
