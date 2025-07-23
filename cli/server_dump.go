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

	"github.com/sourcenetwork/defradb/acp/dac"
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
			ctx := cmd.Context()
			log.InfoContext(ctx, "Dumping DB state...")

			if cfg.GetString("datastore.store") != configStoreBadger {
				return errors.New("server-side dump is only supported for the Badger datastore")
			}
			badgerPath := cfg.GetString("datastore.badger.path")
			storeOpts := []node.StoreOpt{
				node.WithStorePath(badgerPath),
			}
			rootstore, err := node.NewStore(ctx, storeOpts...)
			if err != nil {
				return err
			}
			adminInfo, err := db.NewAdminInfo(ctx, badgerPath, false)
			if err != nil {
				return err
			}
			db, err := db.NewDB(
				ctx,
				rootstore,
				adminInfo,
				dac.NoDocumentACP,
				nil,
			)
			if err != nil {
				return errors.Wrap("failed to initialize database", err)
			}
			defer db.Close()

			return db.PrintDump(ctx)
		},
	}
	return cmd
}
