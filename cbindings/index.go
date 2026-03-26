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
	"strings"

	"github.com/sourcenetwork/defradb/client"
	defraOpts "github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export NewIndex
func NewIndex(
	nodePtr C.uintptr_t,
	indexName *C.char,
	fieldsStr *C.char,
	isUnique C.int,
	options C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	fieldsArg := splitCommaSeparatedString(C.GoString(fieldsStr))
	collectionName := C.GoString(options.name)

	// Parse the fields into an object, considering whether they are each ascending or descending
	var fields []client.IndexedFieldDescription
	for _, field := range fieldsArg {
		// For each field, parse it into a field name and ascension order, separated by a colon
		// If there is no colon, assume the ascension order is ASC by default
		const asc = "ASC"
		const desc = "DESC"
		parts := strings.Split(field, ":")
		fieldName := parts[0]
		order := asc
		if len(parts) == 2 {
			order = strings.ToUpper(parts[1])
			if order != asc && order != desc {
				return returnC(returnGoC(1, errInvalidAscensionOrder, ""))
			}
		} else if len(parts) > 2 {
			return returnC(returnGoC(1, NewErrInvalidIndexFieldDescription(field).Error(), ""))
		}
		fields = append(fields, client.IndexedFieldDescription{
			Name:       fieldName,
			Descending: order == desc,
		})
	}

	desc := client.NewIndexRequest{
		Name:   C.GoString(indexName),
		Fields: fields,
		Unique: isUnique != 0,
	}
	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	ident := iIdentity.FromContext(ctx)
	col, err := store.GetCollectionByName(ctx, collectionName,
		defraOpts.WithIdentity(defraOpts.GetCollectionByName(), ident))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	descWithID, err := col.NewIndex(ctx, desc,
		defraOpts.WithIdentity(defraOpts.NewCollectionIndex(), ident))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(descWithID))
}

//export ListIndexes
func ListIndexes(nodePtr C.uintptr_t, options C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
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

	ident := iIdentity.FromContext(ctx)
	collectionName := C.GoString(options.name)
	switch {
	// Get the indices associated with a given collection
	case collectionName != "":
		col, err := store.GetCollectionByName(ctx, collectionName,
			defraOpts.WithIdentity(defraOpts.GetCollectionByName(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		indices, err := col.ListIndexes(ctx,
			defraOpts.WithIdentity(defraOpts.ListCollectionIndexes(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(marshalJSONToGoCResult(indices))
	// Get all of the indices, because no collection was specified
	default:
		indices, err := store.ListIndexes(ctx,
			defraOpts.WithIdentity(defraOpts.ListIndexes(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(marshalJSONToGoCResult(indices))
	}
}

//export DeleteIndex
func DeleteIndex(nodePtr C.uintptr_t,
	indexName *C.char,
	options C.CollectionOptions,
	identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	collectionName := C.GoString(options.name)

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	ident := iIdentity.FromContext(ctx)
	col, err := store.GetCollectionByName(ctx, collectionName,
		defraOpts.WithIdentity(defraOpts.GetCollectionByName(), ident))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = col.DeleteIndex(ctx, C.GoString(indexName),
		defraOpts.WithIdentity(defraOpts.DeleteCollectionIndex(), ident))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
