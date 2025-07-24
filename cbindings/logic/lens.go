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
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

func LensSet(
	srcSchemaVersionID string,
	dstSchemaVersionID string,
	lensCfgJson string,
	txnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	// Parse the lens config string into a client.LensConfig
	decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnGoC(1, fmt.Sprintf(errInvalidLensConfig, err), "")
	}
	migrationCfg := client.LensConfig{
		SourceSchemaVersionID:      srcSchemaVersionID,
		DestinationSchemaVersionID: dstSchemaVersionID,
		Lens:                       lensCfg,
	}
	err = globalNode.DB.SetMigration(ctx, migrationCfg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func LensDown(collectionID string, documents string, txnID uint64) GoCResult {
	ctx := context.Background()
	srcData := []byte(documents)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnGoC(1, err.Error(), "")
	}

	out, err := globalNode.DB.LensRegistry().MigrateDown(ctx, enumerable.New(src), collectionID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(value)
}

func LensUp(collectionID string, documents string, txnID uint64) GoCResult {
	ctx := context.Background()
	srcData := []byte(documents)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnGoC(1, err.Error(), "")
	}

	out, err := globalNode.DB.LensRegistry().MigrateUp(ctx, enumerable.New(src), collectionID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(value)
}

func LensReload(txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.DB.LensRegistry().ReloadLenses(ctx)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func LensSetRegistry(collectionID string, lensCfgJSON string, txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	decoder := json.NewDecoder(strings.NewReader(lensCfgJSON))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnGoC(1, fmt.Sprintf(errInvalidLensConfig, err), "")
	}

	err = globalNode.DB.LensRegistry().SetMigration(ctx, collectionID, lensCfg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}
