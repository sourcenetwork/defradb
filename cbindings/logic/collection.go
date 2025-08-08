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
	"encoding/json"
	"errors"
	"fmt"
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

// parseCollectionOptions is a helper function that converts a GoCOptions struct into
// a client.CollectionFetchOptions struct disregarding the transaction and identity
// inside the GoCOptions struct
func parseCollectionOptions(gocOptions GoCOptions) client.CollectionFetchOptions {
	versionID := gocOptions.Version
	collectionID := gocOptions.CollectionID
	name := gocOptions.Name
	getInactive := gocOptions.GetInactive != 0
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

// getCollectionForCollectionCommand is a helper function wrapping DB.GetCollections, and ensuring
// that only one collection matches the criteria
func getCollectionForCollectionCommand(
	n int,
	ctx context.Context,
	options client.CollectionFetchOptions,
) (client.Collection, error) {
	cols, err := GetNode(n).DB.GetCollections(ctx, options)
	if err != nil {
		return nil, fmt.Errorf(errGettingCollection, err)
	}

	// Only one collection should match the criteria
	if len(cols) == 0 {
		return nil, errors.New(errNoMatchingCollection)
	}
	if len(cols) > 1 {
		return nil, errors.New(errAmbiguousCollection)
	}
	return cols[0], nil
}

func CollectionCreate(
	n int,
	jsonString string,
	isEncrypted bool,
	encryptFieldsStr string,
	gocOptions GoCOptions,
) GoCResult {
	jsonBytes := []byte(jsonString)
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := contextWithIdentity(ctx, gocOptions.Identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, gocOptions.TxID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(n, ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var encryptFields []string
	if encryptFieldsStr != "" {
		for _, f := range strings.Split(encryptFieldsStr, ",") {
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
			return returnGoC(1, err.Error(), "")
		}
		err = col.CreateMany(ctx, docs)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
	} else {
		// Single document
		doc, err := client.NewDocFromJSON(jsonBytes, col.Definition())
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		err = col.Create(ctx, doc)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
	}
	return returnGoC(0, "", "")
}

func CollectionDelete(n int, docID string, filter string, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := contextWithIdentity(ctx, gocOptions.Identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, gocOptions.TxID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(n, ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	switch {
	case docID != "":
		ID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		_, err = col.Delete(ctx, ID)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		return returnGoC(0, "", "")
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnGoC(1, err.Error(), "")
		}
		res, err := col.DeleteWithFilter(ctx, filterValue)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		return returnGoC(0, "", string(jsonBytes))
	default:
		return returnGoC(1, errNoDocIDOrFilter, "")
	}
}

func CollectionDescribe(n int, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := contextWithIdentity(ctx, gocOptions.Identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, gocOptions.TxID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	cols, err := GetNode(n).DB.GetCollections(ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	colDesc := make([]client.CollectionDefinition, len(cols))
	for i, col := range cols {
		colDesc[i] = col.Definition()
	}

	return marshalJSONToGoCResult(colDesc)
}

func CollectionListDocIDs(n int, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := contextWithIdentity(ctx, gocOptions.Identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, gocOptions.TxID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(n, ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	// Get and return the Doc IDs as a JSON list
	// Note: This is different from the format returned by the CLI, which contains error fields
	docCh, err := col.GetAllDocIDs(ctx)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var results []docIDResult
	for doc := range docCh {
		result := docIDResult{
			DocID: doc.ID.String(),
		}
		if doc.Err != nil {
			// Return immediately upon error
			return returnGoC(1, doc.Err.Error(), "")
		}
		results = append(results, result)
	}

	data, err := json.Marshal(results)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", string(data))
}

func CollectionGet(n int, docIDinput string, showDeleted bool, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := contextWithIdentity(ctx, gocOptions.Identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, gocOptions.TxID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(n, ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	docID, err := client.NewDocIDFromString(docIDinput)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	doc, err := col.Get(ctx, docID, showDeleted)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	docMap, err := doc.ToMap()
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(docMap)
}

func CollectionPatch(n int, patch string, lensString string, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, gocOptions.Identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, gocOptions.TxID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var migration immutable.Option[model.Lens] = immutable.None[model.Lens]()
	if lensString != "" {
		var lensCfg model.Lens
		decoder := json.NewDecoder(strings.NewReader(lensString))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnGoC(1, fmt.Sprintf(errInvalidLensConfig, err), "")
		}

		// Length being greater than 0 also means it is not nil, so no need to check
		if len(lensCfg.Lenses) > 0 {
			migration = immutable.Some(lensCfg)
		}
	}

	err = GetNode(n).DB.PatchCollection(ctx, patch, migration)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func CollectionUpdate(n int, docID string, filter string, updater string, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := contextWithIdentity(ctx, gocOptions.Identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, gocOptions.TxID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(n, ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	switch {
	// Update by filter
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnGoC(1, err.Error(), "")
		}
		res, err := col.UpdateWithFilter(ctx, filterValue, updater)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		return returnGoC(0, "", string(jsonBytes))

	// Update by docID
	case docID != "":
		newDocID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		doc, err := col.Get(ctx, newDocID, true)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		if err := doc.SetWithJSON([]byte(updater)); err != nil {
			return returnGoC(1, err.Error(), "")
		}
		err = col.Update(ctx, doc)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		return returnGoC(0, "", "")
	default:
		return returnGoC(1, errNoDocIDOrFilter, "")
	}
}

func SetActiveCollection(n int, version string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = GetNode(n).DB.SetActiveCollectionVersion(ctx, version)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errSetActiveSchema, err), "")
	}
	return returnGoC(0, "", "")
}
