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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encryption"

	"github.com/sourcenetwork/immutable"
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

// setTransactionOfCollectionCommand is a helper function that takes a context, and returns a new
// one with a transaction attached
func setTransactionOfCollectionCommand(ctx context.Context, goCOptions GoCOptions) (context.Context, error) {
	ctx2, err := contextWithTransaction(ctx, goCOptions.TxID)
	if err != nil {
		return ctx, err
	}
	return ctx2, nil
}

// setIdentityOfCollectionCommand is a helper function that takes a context, and returns a new one
// with an identity attached (or an unmodified context, if the identity is blank)
func setIdentityOfCollectionCommand(ctx context.Context, goCOptions GoCOptions) (context.Context, error) {
	if goCOptions.Identity != "" {
		newctx, err := contextWithIdentity(ctx, goCOptions.Identity)
		if err != nil {
			return ctx, err
		}
		return newctx, nil
	}
	return ctx, nil
}

// getCollectionForCollectionCommand is a helper function wrapping DB.GetCollections, and ensuring
// that only one collection matches the criteria
func getCollectionForCollectionCommand(
	ctx context.Context,
	options client.CollectionFetchOptions,
) (client.Collection, error) {
	cols, err := globalNode.DB.GetCollections(ctx, options)
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

func DocumentCreate(
	jsonString string,
	isEncrypted bool,
	encryptFieldsStr string,
	gocOptions GoCOptions,
) GoCResult {
	jsonBytes := []byte(jsonString)
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := setIdentityOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = setTransactionOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(ctx, options)
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

func DocumentDelete(docID string, filter string, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := setIdentityOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = setTransactionOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(ctx, options)
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

func CollectionDescribe(gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := setIdentityOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = setTransactionOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	cols, err := globalNode.DB.GetCollections(ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	colDesc := make([]client.CollectionDefinition, len(cols))
	for i, col := range cols {
		colDesc[i] = col.Definition()
	}

	return marshalJSONToGoCResult(colDesc)
}

func CollectionListDocIDs(gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := setIdentityOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = setTransactionOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(ctx, options)
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

func DocumentGet(docIDinput string, showDeleted bool, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := setIdentityOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = setTransactionOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(ctx, options)
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

func CollectionPatch(patch string, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()

	ctx, err := setIdentityOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = setTransactionOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.DB.PatchCollection(ctx, patch)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func DocumentUpdate(docID string, filter string, updater string, gocOptions GoCOptions) GoCResult {
	ctx := context.Background()
	options := parseCollectionOptions(gocOptions)

	ctx, err := setIdentityOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = setTransactionOfCollectionCommand(ctx, gocOptions)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	col, err := getCollectionForCollectionCommand(ctx, options)
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
