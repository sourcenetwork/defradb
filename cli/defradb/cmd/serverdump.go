// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cmd

import (
	"context"
	"os"
	"os/signal"

	ds "github.com/ipfs/go-datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/db"
)

// dumpCmd represents the dump command
var srvDumpCmd = &cobra.Command{
	Use:   "server-dump",
	Short: "Dumps the state of the entire database (server side)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		log.Info(ctx, "Starting DefraDB process...")

		// setup signal handlers
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)

		var rootstore ds.Batching
		var err error
		if cfg.Datastore.Store == databaseStoreName {
			log.Info(
				ctx,
				"opening badger store",
				logging.NewKV("Path", cfg.Datastore.Badger.Path),
			)
			rootstore, err = badgerds.NewDatastore(
				cfg.Datastore.Badger.Path,
				cfg.Datastore.Badger.Options,
			)
		} else {
			log.Fatal(ctx, "Server side dump is only supported for the Badger datastore")
		}
		if err != nil {
			log.FatalE(ctx, "Failed to initiate datastore:", err)
		}

		db, err := db.NewDB(ctx, rootstore)
		if err != nil {
			log.FatalE(ctx, "Failed to initiate database:", err)
		}

		log.Info(ctx, "Dumping DB state:")
		db.PrintDump(ctx)
	},
}

func init() {
	rootCmd.AddCommand(srvDumpCmd)
	srvDumpCmd.Flags().String(
		"store",
		defaultCfg.Datastore.Store,
		"Specify the data store to use (supported: badger, memory)",
	)
}
