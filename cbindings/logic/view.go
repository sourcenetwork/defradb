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

import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

func ViewAdd(query string, sdl string, lensCfgJson string, txnID uint64) GoCResult {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Decode the lens configuration into a lens
	var transform immutable.Option[model.Lens]
	if lensCfgJson != "" {
		decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
		decoder.DisallowUnknownFields()
		var lensCfg model.Lens
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnGoC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
		}
		transform = immutable.Some(lensCfg)
	}

	// Add the view
	defs, err := globalNode.DB.AddView(ctx, query, sdl, transform)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(defs)
}

func ViewRefresh(
	name string,
	collectionID string,
	versionID string,
	getInactive bool,
	txnID uint64,
) GoCResult {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Parse the input parameters into a CollectionFetchOptions object
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

	// Refresh the views and return
	err = globalNode.DB.RefreshViews(ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}
