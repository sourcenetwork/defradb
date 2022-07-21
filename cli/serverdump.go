// Copyright 2022 Democratized Data Foundation
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
	"fmt"
	"os"
	"os/signal"

	ds "github.com/ipfs/go-datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
)

func MakeServerDumpCmd() *cobra.Command {
	var datastore string
	cmd := &cobra.Command{
		Use:   "server-dump",
		Short: "Dumps the state of the entire database",
		RunE: func(cmd *cobra.Command, _ []string) error {
			log.Info(cmd.Context(), "Starting DefraDB process...")

			// setup signal handlers
			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt)

			var rootstore ds.Batching
			var err error
			if datastore == badgerDatastoreName {
				info, err := os.Stat(cfg.Datastore.Badger.Path)
				exists := (err == nil && info.IsDir())
				if !exists {
					return fmt.Errorf(
						"Badger store does not exist at %s. Try with an existing directory",
						cfg.Datastore.Badger.Path,
					)
				}
				log.Info(cmd.Context(), "Opening badger store", logging.NewKV("Path", cfg.Datastore.Badger.Path))
				rootstore, err = badgerds.NewDatastore(cfg.Datastore.Badger.Path, cfg.Datastore.Badger.Options)
				if err != nil {
					return fmt.Errorf("could not open badger datastore: %w", err)
				}
			} else {
				return fmt.Errorf("server-side dump is only supported for the Badger datastore")
			}

			db, err := db.NewDB(cmd.Context(), rootstore)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}

			log.Info(cmd.Context(), "Dumping DB state...")
			db.PrintDump(cmd.Context())
			return nil
		},
	}

	cmd.Flags().StringVar(
		&datastore, "store", cfg.Datastore.Store,
		"datastore to use. Options are badger, memory",
	)

	return cmd
}
