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

	"github.com/spf13/cobra"
)

func MakeIndexCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "index",
		Short: "Manage collections' indexes of a running DefraDB instance",
		Long:  `Manage (new, delete, or list) collection indexes on a DefraDB node.`,
	}

	return cmd
}
