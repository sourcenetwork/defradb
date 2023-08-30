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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/immutable/enumerable"
)

var _ client.LensRegistry = (*lensWrapper)(nil)

type lensWrapper struct {
	lens client.LensRegistry
	cmd  *cliWrapper
}

func (w *lensWrapper) WithTxn(tx datastore.Txn) client.LensRegistry {
	return &lensWrapper{
		lens: w.lens.WithTxn(tx),
		cmd:  w.cmd.withTxn(tx),
	}
}

func (w *lensWrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	args := []string{"client", "schema", "migration", "set"}
	args = append(args, config.SourceSchemaVersionID)
	args = append(args, config.DestinationSchemaVersionID)

	lensCfg, err := json.Marshal(config.Lens)
	if err != nil {
		return err
	}
	args = append(args, string(lensCfg))

	_, err = w.cmd.execute(ctx, args)
	return err
}

func (w *lensWrapper) ReloadLenses(ctx context.Context) error {
	return w.lens.ReloadLenses(ctx)
}

func (w *lensWrapper) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	return w.lens.MigrateUp(ctx, src, schemaVersionID)
}

func (w *lensWrapper) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	return w.lens.MigrateDown(ctx, src, schemaVersionID)
}

func (w *lensWrapper) Config(ctx context.Context) ([]client.LensConfig, error) {
	args := []string{"client", "schema", "migration", "get"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var cfgs []client.LensConfig
	if err := json.Unmarshal(data, &cfgs); err != nil {
		return nil, err
	}
	return cfgs, nil
}

func (w *lensWrapper) HasMigration(ctx context.Context, schemaVersionID string) (bool, error) {
	return w.lens.HasMigration(ctx, schemaVersionID)
}
