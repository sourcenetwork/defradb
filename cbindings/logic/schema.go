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

func AddSchema(n int, newSchema string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	collectionVersions, err := GlobalNodes[n].DB.AddSchema(ctx, newSchema)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errAddingSchema, err), "")
	}
	return marshalJSONToGoCResult(collectionVersions)
}

func DescribeSchema(n int, name string, root string, version string, txnID uint64) GoCResult {
	ctx := context.Background()
	options := client.SchemaFetchOptions{}

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

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
	schemas, err := GlobalNodes[n].DB.GetSchemas(ctx, options)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errGettingSchema, err), "")
	}
	return marshalJSONToGoCResult(schemas)
}

func PatchSchema(n int, patchString string, lensString string, setActive bool, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	if patchString == "" {
		return returnGoC(1, errEmptyPatch, "")
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

	err = GlobalNodes[n].DB.PatchSchema(ctx, patchString, migration, setActive)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errPatchingSchema, err), "")
	}
	return returnGoC(0, "", "")
}

func SetActiveSchema(n int, version string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = GlobalNodes[n].DB.SetActiveSchemaVersion(ctx, version)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errSetActiveSchema, err), "")
	}
	return returnGoC(0, "", "")
}
