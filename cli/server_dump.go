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

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/http"
)

func MakeServerDumpCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server-dump",
		Short: "Dumps the state of the entire database",
		RunE: func(cmd *cobra.Command, _ []string) error {
			db, err := http.NewClient(cfg.API.Address)
			if err != nil {
				return err
			}
			return db.PrintDump(cmd.Context())
		},
	}
	return cmd
}
