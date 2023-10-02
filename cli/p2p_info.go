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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/http"
)

func MakeP2PInfoCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Get peer info from a DefraDB node",
		Long:  `Get peer info from a DefraDB node`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db := cmd.Context().Value(dbContextKey).(*http.Client)

			res, err := db.PeerInfo(cmd.Context())
			if err != nil {
				return err
			}
			return writeJSON(cmd, res)
		},
	}
	return cmd
}
