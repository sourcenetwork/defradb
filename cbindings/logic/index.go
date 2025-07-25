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

import (
	"context"
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

func IndexCreate(
	collectionName string,
	indexName string,
	fieldsStr string,
	isUnique bool,
	txnID uint64,
) GoCResult {
	ctx := context.Background()
	fieldsArg := splitCommaSeparatedString(fieldsStr)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

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
				return returnGoC(1, errInvalidAscensionOrder, "")
			}
		} else if len(parts) > 2 {
			return returnGoC(1, errInvalidIndexFieldDescription, "")
		}
		fields = append(fields, client.IndexedFieldDescription{
			Name:       fieldName,
			Descending: order == desc,
		})
	}

	desc := client.IndexCreateRequest{
		Name:   indexName,
		Fields: fields,
		Unique: isUnique,
	}
	col, err := globalNode.DB.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	descWithID, err := col.CreateIndex(ctx, desc)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(descWithID)
}

func IndexList(collectionName string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	switch {
	// Get the indices associated with a given collection
	case collectionName != "":
		col, err := globalNode.DB.GetCollectionByName(ctx, collectionName)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		indices, err := col.GetIndexes(ctx)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		return marshalJSONToGoCResult(indices)
	// Get all of the indices, because no collection was specified
	default:
		indices, err := globalNode.DB.GetAllIndexes(ctx)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		return marshalJSONToGoCResult(indices)
	}
}

func IndexDrop(collectionName string, indexName string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := globalNode.DB.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	err = col.DropIndex(ctx, indexName)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}
