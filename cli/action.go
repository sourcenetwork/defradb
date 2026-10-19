// Copyright 2026 Democratized Data Foundation
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

func MakeActionCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "action",
		Short: "Manage DefraDB actions",
		Long: `Manage DefraDB actions.
Manage long running actions such as truncate, RefreshView, and the (re)building of indexes.
`,
	}

	EmbedCLIExample(ctx, cmd, "List information about actions",
		"defradb client action list")

	return cmd
}
