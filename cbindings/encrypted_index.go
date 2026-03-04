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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export NewEncryptedIndex
func NewEncryptedIndex(
	nodePtr C.uintptr_t,
	collectionName *C.char,
	fieldName *C.char,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	desc := client.EncryptedIndexDescription{
		FieldName: C.GoString(fieldName),
	}
	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	getColOpt := options.WithIdentity(options.GetCollectionByName(), iIdentity.FromContext(ctx))
	col, err := store.GetCollectionByName(ctx, C.GoString(collectionName), getColOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	addOpt := options.WithIdentity(options.NewEncryptedIndex(), iIdentity.FromContext(ctx))
	descWithID, err := col.NewEncryptedIndex(ctx, desc, addOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(descWithID))
}

//export ListEncryptedIndexes
func ListEncryptedIndexes(nodePtr C.uintptr_t, collectionName *C.char, identityPtr C.uintptr_t) C.Result {
	ctx, err := contextWithIdentity(context.Background(), identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	colName := C.GoString(collectionName)
	switch {
	// Get the encrypted indices associated with a given collection
	case colName != "":
		getColOpt := options.WithIdentity(options.GetCollectionByName(), iIdentity.FromContext(ctx))
		col, err := store.GetCollectionByName(ctx, colName, getColOpt)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		opts := options.WithIdentity(
			options.ListCollectionEncryptedIndexes(),
			iIdentity.FromContext(ctx),
		)
		indices, err := col.ListEncryptedIndexes(ctx, opts)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(marshalJSONToGoCResult(indices))
	// Get all of the encrypted indices, because no collection was specified
	default:
		opts := options.WithIdentity(
			options.ListAllEncryptedIndexes(),
			iIdentity.FromContext(ctx),
		)
		indices, err := store.ListAllEncryptedIndexes(ctx, opts)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(marshalJSONToGoCResult(indices))
	}
}

//export DeleteEncryptedIndex
func DeleteEncryptedIndex(
	nodePtr C.uintptr_t, collectionName *C.char, fieldName *C.char, identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	getColOpt := options.WithIdentity(options.GetCollectionByName(), iIdentity.FromContext(ctx))
	col, err := store.GetCollectionByName(ctx, C.GoString(collectionName), getColOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	deleteOpt := options.WithIdentity(options.DeleteEncryptedIndex(), iIdentity.FromContext(ctx))
	err = col.DeleteEncryptedIndex(ctx, C.GoString(fieldName), deleteOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
