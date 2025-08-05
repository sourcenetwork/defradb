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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
)

func ViewAdd(n int, query string, sdl string, lensCfgJson string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var transform immutable.Option[model.Lens]
	if lensCfgJson != "" {
		decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
		decoder.DisallowUnknownFields()
		var lensCfg model.Lens
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnGoC(1, fmt.Sprintf(errInvalidLensConfig, err), "")
		}
		transform = immutable.Some(lensCfg)
	}

	defs, err := GetNode(n).DB.AddView(ctx, query, sdl, transform)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(defs)
}

func ViewRefresh(
	n int,
	name string,
	collectionID string,
	versionID string,
	getInactive bool,
	txnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

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

	err = GetNode(n).DB.RefreshViews(ctx, options)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}
