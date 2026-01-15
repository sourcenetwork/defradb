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
	"context"

	"github.com/spf13/cobra"
)

func MakeCollectionTruncateCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "truncate",
		Short: "Truncate the given collection",
		Long: `Truncate the given collection, removing all document data within it from the local node.
 Does not propagate the deletion to other Defra nodes in the peer network.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			return col.Truncate(cmd.Context())
		},
	}
	return cmd
}
