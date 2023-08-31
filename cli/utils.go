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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/http"
)

type contextKey string

var (
	txContextKey    = contextKey("tx")
	dbContextKey    = contextKey("db")
	storeContextKey = contextKey("store")
)

// setTransactionContext sets the transaction for the current command context.
func setTransactionContext(cmd *cobra.Command, cfg *config.Config, txId uint64) error {
	if txId == 0 {
		return nil
	}
	tx, err := http.NewTransaction(cfg.API.Address, txId)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	ctx = context.WithValue(ctx, txContextKey, tx)
	cmd.SetContext(ctx)
	return nil
}

// setStoreContext sets the store for the current command context.
func setStoreContext(cmd *cobra.Command, cfg *config.Config) error {
	db, err := http.NewClient(cfg.API.Address)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	ctx = context.WithValue(ctx, dbContextKey, db)
	if tx, ok := ctx.Value(txContextKey).(datastore.Txn); ok {
		ctx = context.WithValue(ctx, storeContextKey, db.WithTxn(tx))
	} else {
		ctx = context.WithValue(ctx, storeContextKey, db)
	}
	cmd.SetContext(ctx)
	return nil
}

// loadConfig loads the rootDir containing the configuration file,
// otherwise warn about it and load a default configuration.
func loadConfig(cfg *config.Config) error {
	if err := cfg.LoadRootDirFromFlagOrDefault(); err != nil {
		return err
	}
	return cfg.LoadWithRootdir(cfg.ConfigFileExists())
}

// createConfig creates the config directories and writes
// the current config to a file.
func createConfig(cfg *config.Config) error {
	if config.FolderExists(cfg.Rootdir) {
		return cfg.WriteConfigFile()
	}
	return cfg.CreateRootDirAndConfigFile()
}

func writeJSON(cmd *cobra.Command, out any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
