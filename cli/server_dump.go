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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/node"
)

func MakeServerDumpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server-dump",
		Short: "Dumps the state of the entire database",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := mustGetContextConfig(cmd)
			log.InfoContext(cmd.Context(), "Dumping DB state...")

			if cfg.GetString("datastore.store") != configStoreBadger {
				return errors.New("server-side dump is only supported for the Badger datastore")
			}
			storeOpts := []node.StoreOpt{
				node.WithBadgerPath(cfg.GetString("datastore.badger.path")),
			}
			rootstore, err := node.NewStore(cmd.Context(), storeOpts...)
			if err != nil {
				return err
			}
			db, err := db.NewDB(cmd.Context(), rootstore, acp.NoACP, nil)
			if err != nil {
				return errors.Wrap("failed to initialize database", err)
			}
			defer db.Close()

			return db.PrintDump(cmd.Context())
		},
	}
	return cmd
}
