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

	"github.com/sourcenetwork/defradb/client"
)

func MakeNodeACPReEnableCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "re-enable [-i --identity]",
		Short: "Re-enable the node access control",
		Long: `Re-enable the node access control

Example:
  defradb client acp node re-enable -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Note:
- This command will re-enable an already configured node acp system that is temporarily disabled.
- If node acp is already enabled, then it will return an error.
- If node acp is in a clean/non-configured state, then it will return an error.

Learn more about the DefraDB [ACP System](/acp/README.md)

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			if err := cliClient.ReEnableNAC(cmd.Context()); err != nil {
				return err
			}

			return writeJSON(cmd, client.SuccessResponse{Success: true})
		},
	}
	return cmd
}
