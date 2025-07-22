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
	"fmt"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

func AddSchema(newSchema string, txnID uint64) GoCResult {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	collectionVersions, err := globalNode.DB.AddSchema(ctx, newSchema)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(cerrAddingSchema, err), "")
	}
	return marshalJSONToGoCResult(collectionVersions)
}

func DescribeSchema(name string, root string, version string, txnID uint64) GoCResult {
	ctx := context.Background()
	options := client.SchemaFetchOptions{}

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the configuration options from the passed in parameters
	if version != "" {
		options.ID = immutable.Some(version)
	}
	if root != "" {
		options.Root = immutable.Some(root)
	}
	if name != "" {
		options.Name = immutable.Some(name)
	}

	// Get the schema, and try to convert it to JSON for return
	schemas, err := globalNode.DB.GetSchemas(ctx, options)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(cerrGettingSchema, err), "")
	}
	return marshalJSONToGoCResult(schemas)
}

func PatchSchema(patchString string, lensString string, setActive bool, txnID uint64) GoCResult {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	if patchString == "" {
		return returnGoC(1, cerrEmptyPatch, "")
	}

	var migration immutable.Option[model.Lens] = immutable.None[model.Lens]()
	if lensString != "" {
		var lensCfg model.Lens
		decoder := json.NewDecoder(strings.NewReader(lensString))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnGoC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
		}

		// Only assign Some if Lense length is greater than 0 (and therefore it is not nil)
		if len(lensCfg.Lenses) > 0 {
			migration = immutable.Some(lensCfg)
		}
	}

	err = globalNode.DB.PatchSchema(ctx, patchString, migration, setActive)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(cerrPatchingSchema, err), "")
	}
	return returnGoC(0, "", "")
}

func SetActiveSchema(version string, txnID uint64) GoCResult {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	err = globalNode.DB.SetActiveSchemaVersion(ctx, version)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(cerrSetActiveSchema, err), "")
	}
	return returnGoC(0, "", "")
}
