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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	ds "github.com/sourcenetwork/defradb/datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeServerDumpCmd(cfg *config.Config) *cobra.Command {
	var datastore string

	cmd := &cobra.Command{
		Use:   "server-dump",
		Short: "Dumps the state of the entire database",
		RunE: func(cmd *cobra.Command, _ []string) error {
			log.FeedbackInfo(cmd.Context(), "Starting DefraDB process...")

			// setup signal handlers
			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt)

			var rootstore ds.RootStore
			var err error
			if datastore == badgerDatastoreName {
				info, err := os.Stat(cfg.Datastore.Badger.Path)
				exists := (err == nil && info.IsDir())
				if !exists {
					return errors.New(fmt.Sprintf(
						"badger store does not exist at %s. Try with an existing directory",
						cfg.Datastore.Badger.Path,
					))
				}
				log.FeedbackInfo(cmd.Context(), "Opening badger store", logging.NewKV("Path", cfg.Datastore.Badger.Path))
				rootstore, err = badgerds.NewDatastore(cfg.Datastore.Badger.Path, cfg.Datastore.Badger.Options)
				if err != nil {
					return errors.Wrap("could not open badger datastore", err)
				}
			} else {
				return errors.New("server-side dump is only supported for the Badger datastore")
			}

			db, err := db.NewDB(cmd.Context(), rootstore)
			if err != nil {
				return errors.Wrap("failed to initialize database", err)
			}

			log.FeedbackInfo(cmd.Context(), "Dumping DB state...")
			return db.PrintDump(cmd.Context())
		},
	}
	cmd.Flags().StringVar(
		&datastore, "store", cfg.Datastore.Store,
		"Datastore to use. Options are badger, memory",
	)
	return cmd
}
