// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
)

var _ client.LensRegistry = (*LensRegistry)(nil)

type LensRegistry struct {
	cmd *cliWrapper
}

func (w *LensRegistry) Init(txnSource client.TxnSource) {}

func (w *LensRegistry) SetMigration(ctx context.Context, collectionID string, config model.Lens) error {
	args := []string{"client", "lens", "set-registry"}

	lenses, err := json.Marshal(config)
	if err != nil {
		return err
	}
	args = append(args, collectionID)
	args = append(args, string(lenses))

	_, err = w.cmd.execute(ctx, args)
	return err
}

func (w *LensRegistry) ReloadLenses(ctx context.Context) error {
	args := []string{"client", "lens", "reload"}

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *LensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	args := []string{"client", "lens", "up"}
	args = append(args, "--collection", fmt.Sprint(collectionID))

	var srcData []map[string]any
	err := enumerable.ForEach(src, func(item map[string]any) {
		srcData = append(srcData, item)
	})
	if err != nil {
		return nil, err
	}
	srcJSON, err := json.Marshal(srcData)
	if err != nil {
		return nil, err
	}
	args = append(args, string(srcJSON))

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var out enumerable.Enumerable[map[string]any]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *LensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	args := []string{"client", "lens", "down"}
	args = append(args, "--collection", fmt.Sprint(collectionID))

	var srcData []map[string]any
	err := enumerable.ForEach(src, func(item map[string]any) {
		srcData = append(srcData, item)
	})
	if err != nil {
		return nil, err
	}
	srcJSON, err := json.Marshal(srcData)
	if err != nil {
		return nil, err
	}
	args = append(args, string(srcJSON))

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var out enumerable.Enumerable[map[string]any]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
