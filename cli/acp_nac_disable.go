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

func MakeNodeACPDisableCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "disable [-i --identity]",
		Short: "Disable the node access control",
		Long: `Disable the node access control

Example:
  defradb client acp node disable -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Note:
- This command will disable an node acp system that is currently enabled.
- If node acp is already disabled, then it will return an error.
- If node acp is in a clean/non-configured state, then it will return an error.

Learn more about the DefraDB [ACP System](/acp/README.md)

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			if err := cliClient.DisableNAC(cmd.Context()); err != nil {
				return err
			}

			return writeJSON(cmd, client.SuccessResponse{Success: true})
		},
	}
	return cmd
}
