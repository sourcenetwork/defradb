// Copyright 2025 Democratized Data Foundation
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
)

func MakeP2PDocumentGetAllCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "getall",
		Short: "Get all P2P documents",
		Long: `Get all P2P documents in the pubsub topics.
This is the list of documents of the node that are synchronized on the pubsub network.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			cols, err := cliClient.GetAllP2PDocuments(cmd.Context())
			if err != nil {
				return err
			}
			return writeJSON(cmd, cols)
		},
	}
	return cmd
}
