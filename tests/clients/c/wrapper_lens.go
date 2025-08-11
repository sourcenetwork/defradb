// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cwrap

import (
	"context"
	"encoding/json"
	"errors"

	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"
	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

var _ client.LensRegistry = (*LensRegistry)(nil)

type LensRegistry struct {
	*CWrapper
}

func (w *LensRegistry) Init(txnSource client.TxnSource) {}

func (w *LensRegistry) SetMigration(ctx context.Context, collectionID string, config model.Lens) error {
	txnID := txnIDFromContext(ctx)
	cfgBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	lens := string(cfgBytes)

	result := cbindings.LensSetRegistry(w.nodeNum, collectionID, lens, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *LensRegistry) ReloadLenses(ctx context.Context) error {
	txnID := txnIDFromContext(ctx)

	result := cbindings.LensReload(w.nodeNum, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *LensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	txnID := txnIDFromContext(ctx)
	docs, err := collectEnumerable(src)
	if err != nil {
		return nil, err
	}
	docBytes, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}
	docStr := string(docBytes)

	result := cbindings.LensUp(w.nodeNum, collectionID, docStr, txnID)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	var out []map[string]any
	if err := json.Unmarshal([]byte(result.Value), &out); err != nil {
		return nil, err
	}
	return enumerable.New(out), nil
}

func (w *LensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	txnID := txnIDFromContext(ctx)
	docs, err := collectEnumerable(src)
	if err != nil {
		return nil, err
	}

	docBytes, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}

	docStr := string(docBytes)

	result := cbindings.LensDown(w.nodeNum, collectionID, docStr, txnID)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	var out []map[string]any
	if err := json.Unmarshal([]byte(result.Value), &out); err != nil {
		return nil, err
	}
	return enumerable.New(out), nil
}
