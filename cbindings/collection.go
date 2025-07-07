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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/encryption"

	"github.com/sourcenetwork/immutable"
)

// Helper struct
type docIDResult struct {
	DocID string `json:"docID"`
	Error string `json:"error"`
}

// Helper function
// Parse the simple (not txn, not identity) options from the C struct into a CollectionFetchOptions object
func parseCollectionOptions(cOptions C.CollectionOptions) client.CollectionFetchOptions {
	versionID := C.GoString(cOptions.version)
	collectionID := C.GoString(cOptions.collectionID)
	name := C.GoString(cOptions.name)
	getInactive := cOptions.getInactive != 0
	options := client.CollectionFetchOptions{}
	if versionID != "" {
		options.VersionID = immutable.Some(versionID)
	}
	if collectionID != "" {
		options.CollectionID = immutable.Some(collectionID)
	}
	if name != "" {
		options.Name = immutable.Some(name)
	}
	if getInactive {
		options.IncludeInactive = immutable.Some(getInactive)
	}
	return options
}

// Helper function
// Set the transaction associated with a given context from the value inside a C.CollectionOptions struct
func setTransactionOfCollectionCommand(ctx context.Context, cOptions C.CollectionOptions) (context.Context, error) {
	TxnIDu64 := uint64(cOptions.tx)
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

// Helper function
// Set the identity associated with a given context from the value inside a C.CollectionOptions struct
// If the identity string is blank, return the original context, unchanged
func setIdentityOfCollectionCommand(ctx context.Context, cOptions C.CollectionOptions) (context.Context, error) {
	identityStr := C.GoString(cOptions.identity)
	if identityStr != "" {
		newctx, err := contextWithIdentity(ctx, identityStr)
		if err != nil {
			return ctx, err
		}
		return newctx, nil
	}
	return ctx, nil
}

// Helper function
func getCollectionForCollectionCommand(
	ctx context.Context,
	options client.CollectionFetchOptions,
) (client.Collection, error) {
	// Get the collections that match the options criteria
	cols, err := globalNode.DB.GetCollections(ctx, options)
	if err != nil {
		return nil, fmt.Errorf(cerrGettingCollection, err)
	}

	// Make sure only one collection matches the criteria, and select it
	if len(cols) == 0 {
		return nil, errors.New(cerrNoMatchingCollection)
	}
	if len(cols) > 1 {
		return nil, errors.New(cerrAmbiguousCollection)
	}
	return cols[0], nil
}

//export collectionCreate
func collectionCreate(
	cJSON *C.char,
	cIsEncrypted C.int,
	cEncryptedFields *C.char,
	cOptions C.CollectionOptions,
) *C.Result {
	isEncrypted := cIsEncrypted != 0
	jsonString := C.GoString(cJSON)
	jsonBytes := []byte(jsonString)
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Attach the identity to the context
	ctx, err := setIdentityOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the correct collection
	foundcol, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	col := foundcol

	// Set the context's collection to the selected one
	ctx = context.WithValue(ctx, collectionContextKey{}, col)

	// Set the encryption
	raw := C.GoString(cEncryptedFields)
	var encryptFields []string
	if raw != "" {
		for _, f := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				encryptFields = append(encryptFields, trimmed)
			}
		}
	}
	ctx = encryption.SetContextConfigFromParams(ctx, isEncrypted, encryptFields)

	// Determine if JSON is array or object by looking for the first character being [
	jsonString = strings.TrimSpace(jsonString)
	if strings.HasPrefix(jsonString, "[") {
		// Multiple documents
		docs, err := client.NewDocsFromJSON(jsonBytes, col.Definition())
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		err = col.CreateMany(ctx, docs)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
	} else {
		// Single document
		doc, err := client.NewDocFromJSON(jsonBytes, col.Definition())
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		err = col.Create(ctx, doc)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
	}

	return returnC(0, "", "")
}

//export collectionDelete
func collectionDelete(cDocID *C.char, cFilter *C.char, cOptions C.CollectionOptions) *C.Result {
	docID := C.GoString(cDocID)
	filter := C.GoString(cFilter)
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Attach the identity to the context
	ctx, err := setIdentityOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the correct collection
	col, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	switch {
	case docID != "":
		ID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		_, err = col.Delete(ctx, ID)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		return returnC(0, "", "")
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnC(1, err.Error(), "")
		}
		res, err := col.DeleteWithFilter(ctx, filterValue)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		return returnC(0, "", string(jsonBytes))
	default:
		return returnC(1, cerrNoDocIDOrFilter, "")
	}
}

//export collectionDescribe
func collectionDescribe(cOptions C.CollectionOptions) *C.Result {
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Attach the identity to the context
	ctx, err := setIdentityOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Get the collections
	cols, err := globalNode.DB.GetCollections(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Get the descriptions
	colDesc := make([]client.CollectionDefinition, len(cols))
	for i, col := range cols {
		colDesc[i] = col.Definition()
	}

	return marshalJSONToCResult(colDesc)
}

//export collectionListDocIDs
func collectionListDocIDs(cOptions C.CollectionOptions) *C.Result {
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Attach the identity to the context
	ctx, err := setIdentityOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Get the collection
	col, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Get and return the Doc IDs as a JSON list
	// Note: This is different from the format returned by the CLI, which contains error fields
	docCh, err := col.GetAllDocIDs(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Range over the chan, filling out an array of docIDResults (ID/error pairs)
	var results []docIDResult
	for doc := range docCh {
		result := docIDResult{
			DocID: doc.ID.String(),
		}
		if doc.Err != nil {
			result.Error = doc.Err.Error()
		}
		results = append(results, result)
	}

	// Marshal and return the JSON
	data, err := json.Marshal(results)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", string(data))
}

//export collectionGet
func collectionGet(cDocID *C.char, cShowDeleted C.int, cOptions C.CollectionOptions) *C.Result {
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)
	docIDinput := C.GoString(cDocID)
	showDeleted := cShowDeleted != 0

	// Attach the identity to the context
	ctx, err := setIdentityOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Get the collection
	col, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Get the document by docID
	docID, err := client.NewDocIDFromString(docIDinput)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	doc, err := col.Get(ctx, docID, showDeleted)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	docMap, err := doc.ToMap()
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(docMap)
}

//export collectionPatch
func collectionPatch(cPatch *C.char, cOptions C.CollectionOptions) *C.Result {
	ctx := context.Background()
	patch := C.GoString(cPatch)

	// Attach the identity to the context
	ctx, err := setIdentityOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to patch the collection
	err = globalNode.DB.PatchCollection(ctx, patch)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

//export collectionUpdate
func collectionUpdate(cDocID *C.char, cFilter *C.char, cUpdater *C.char, cOptions C.CollectionOptions) *C.Result {
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)
	docID := C.GoString(cDocID)
	filter := C.GoString(cFilter)
	updater := C.GoString(cUpdater)

	// Attach the identity to the context
	ctx, err := setIdentityOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Get the collection
	col, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the context's collection to the selected one
	ctx = context.WithValue(ctx, collectionContextKey{}, col)
	ctx = context.WithValue(ctx, schemaNameContextKey{}, options.CollectionID.Value())

	switch {
	// Update by filter
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnC(1, err.Error(), "")
		}
		res, err := col.UpdateWithFilter(ctx, filterValue, updater)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		return returnC(0, "", string(jsonBytes))

	// Update by docID
	case docID != "":
		newDocID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		doc, err := col.Get(ctx, newDocID, true)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		if err := doc.SetWithJSON([]byte(updater)); err != nil {
			return returnC(1, err.Error(), "")
		}
		err = col.Update(ctx, doc)
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		return returnC(0, "", "")
	default:
		return returnC(1, cerrNoDocIDOrFilter, "")
	}
}
