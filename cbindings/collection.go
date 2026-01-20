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
	"encoding/json"
	"strings"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

type docIDResult struct {
	DocID string `json:"docID"`
	Error string `json:"error"`
}

// parseCollectionOptions is a helper function that converts a C.CollectionOptions
// struct into a client.CollectionFetchOptions struct disregarding the identity
func parseCollectionOptions(opts C.CollectionOptions) client.CollectionFetchOptions {
	versionID := C.GoString(opts.version)
	collectionID := C.GoString(opts.collectionID)
	name := C.GoString(opts.name)
	getInactive := opts.getInactive != 0
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

// getCollection is a helper function wrapping DB.GetCollections, and ensuring
// that only one collection matches the criteria
func getCollection(
	store client.Store,
	ctx context.Context,
	options client.CollectionFetchOptions,
) (client.Collection, error) {
	cols, err := store.GetCollections(ctx, options)
	if err != nil {
		return nil, err
	}

	// Only one collection should match the criteria
	if len(cols) == 0 {
		return nil, NewErrNoMatchingCollection()
	}
	if len(cols) > 1 {
		return nil, NewErrAmbiguousCollection()
	}
	return cols[0], nil
}

//export CollectionCreate
func CollectionCreate(
	nodePtr C.uintptr_t,
	json *C.char,
	isEncrypted C.int,
	encryptedFields *C.char,
	options C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	colOptions := parseCollectionOptions(options)

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var encryptFields []string
	encryptFieldsStr := C.GoString(encryptedFields)
	if encryptFieldsStr != "" {
		for _, f := range strings.Split(encryptFieldsStr, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				encryptFields = append(encryptFields, trimmed)
			}
		}
	}
	ctx = encryption.SetContextConfigFromParams(ctx, isEncrypted != 0, encryptFields)

	// Determine if JSON is array or object by looking for the first character being [
	jsonString := strings.TrimSpace(C.GoString(json))
	if strings.HasPrefix(jsonString, "[") {
		// Multiple documents
		docs, err := client.NewDocsFromJSON(ctx, []byte(jsonString), col.Version())
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		err = col.CreateMany(ctx, docs)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
	} else {
		// Single document
		doc, err := client.NewDocFromJSON(ctx, []byte(jsonString), col.Version())
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		err = col.Create(ctx, doc)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
	}
	return returnC(returnGoC(0, "", ""))
}

//export CollectionDelete
func CollectionDelete(nodePtr C.uintptr_t,
	docIDStr *C.char,
	filterStr *C.char,
	options C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	colOptions := parseCollectionOptions(options)

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	docID := C.GoString(docIDStr)
	filter := C.GoString(filterStr)
	switch {
	case docID != "":
		ID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		_, err = col.Delete(ctx, ID)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", ""))
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		res, err := col.DeleteWithFilter(ctx, filterValue)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", string(jsonBytes)))
	default:
		return returnC(returnGoC(1, errNoDocIDOrFilter, ""))
	}
}

//export CollectionDescribe
func CollectionDescribe(nodePtr C.uintptr_t, options C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	colOptions := parseCollectionOptions(options)

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	cols, err := store.GetCollections(ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	colDesc := make([]client.CollectionVersion, len(cols))
	for i, col := range cols {
		colDesc[i] = col.Version()
	}

	return returnC(marshalJSONToGoCResult(colDesc))
}

//export CollectionListDocIDs
func CollectionListDocIDs(nodePtr C.uintptr_t, options C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	colOptions := parseCollectionOptions(options)

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	// Get and return the Doc IDs as a JSON list
	// Note: This is different from the format returned by the CLI, which contains error fields
	docCh, err := col.GetAllDocIDs(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var results []docIDResult
	for doc := range docCh {
		result := docIDResult{
			DocID: doc.ID.String(),
		}
		if doc.Err != nil {
			// Return immediately upon error
			return returnC(returnGoC(1, doc.Err.Error(), ""))
		}
		results = append(results, result)
	}

	data, err := json.Marshal(results)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", string(data)))
}

//export CollectionGet
func CollectionGet(nodePtr C.uintptr_t,
	docIDStr *C.char,
	showDeleted C.int,
	options C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	colOptions := parseCollectionOptions(options)

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	docID, err := client.NewDocIDFromString(C.GoString(docIDStr))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	doc, err := col.Get(ctx, docID, showDeleted != 0)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	docMap, err := doc.ToMap()
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(docMap))
}

//export CollectionPatch
func CollectionPatch(nodePtr C.uintptr_t,
	patch *C.char, lensConfig *C.char,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var migration immutable.Option[model.Lens] = immutable.None[model.Lens]()
	lensString := C.GoString(lensConfig)
	if lensString != "" {
		var lensCfg model.Lens
		decoder := json.NewDecoder(strings.NewReader(lensString))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}

		// Length being greater than 0 also means it is not nil, so no need to check
		if len(lensCfg.Lenses) > 0 {
			migration = immutable.Some(lensCfg)
		}
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = store.PatchCollection(ctx, C.GoString(patch), migration)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export CollectionUpdate
func CollectionUpdate(
	nodePtr C.uintptr_t,
	docIDStr *C.char,
	filterStr *C.char,
	updaterStr *C.char,
	options C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	colOptions := parseCollectionOptions(options)

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	docID := C.GoString(docIDStr)
	filter := C.GoString(filterStr)
	updater := C.GoString(updaterStr)
	switch {
	// Update by filter
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		res, err := col.UpdateWithFilter(ctx, filterValue, updater)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", string(jsonBytes)))

	// Update by docID
	case docID != "":
		newDocID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		doc, err := col.Get(ctx, newDocID, true)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		if err := doc.SetWithJSON(ctx, []byte(updater)); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		err = col.Update(ctx, doc)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", ""))
	default:
		return returnC(returnGoC(1, errNoDocIDOrFilter, ""))
	}
}

//export SetActiveCollection
func SetActiveCollection(nodePtr C.uintptr_t, options C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	versionID := C.GoString(options.version)

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = store.SetActiveCollectionVersion(ctx, versionID)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export CollectionTruncate
func CollectionTruncate(
	nodePtr C.uintptr_t,
	options C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	colOptions := parseCollectionOptions(options)

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = col.Truncate(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(returnGoC(0, "", ""))
}
