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
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

//export indexCreate
func indexCreate(
	cCollectionName *C.char,
	cIndexName *C.char,
	cFields *C.char,
	cIsUnique C.int,
	cTxnID C.ulonglong,
) *C.Result {
	ctx := context.Background()
	collectionName := C.GoString(cCollectionName)
	indexName := C.GoString(cIndexName)
	isUnique := cIsUnique != 0
	fieldsArg := splitCommaSeparatedCString(cFields)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

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
				return returnC(1, cerrInvalidAscensionOrder, "")
			}
		} else if len(parts) > 2 {
			return returnC(1, cerrInvalidIndexFieldDescription, "")
		}
		fields = append(fields, client.IndexedFieldDescription{
			Name:       fieldName,
			Descending: order == desc,
		})
	}

	// Create a request object, and try to create the index
	desc := client.IndexCreateRequest{
		Name:   indexName,
		Fields: fields,
		Unique: isUnique,
	}
	col, err := globalNode.DB.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	descWithID, err := col.CreateIndex(ctx, desc)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(descWithID)
}

//export indexList
func indexList(cCollectionName *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	collectionName := C.GoString(cCollectionName)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	switch {
	// Get the indices associated with a given collection
	case collectionName != "":
		col, err := globalNode.DB.GetCollectionByName(ctx, collectionName)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		indices, err := col.GetIndexes(ctx)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		return marshalJSONToCResult(indices)
	// Get all of the indices, because no collection was specified
	default:
		indices, err := globalNode.DB.GetAllIndexes(ctx)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		return marshalJSONToCResult(indices)
	}
}

//export indexDrop
func indexDrop(cCollectionName *C.char, cIndexName *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	collectionName := C.GoString(cCollectionName)
	indexName := C.GoString(cIndexName)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Get collection and return
	col, err := globalNode.DB.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	err = col.DropIndex(ctx, indexName)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}
