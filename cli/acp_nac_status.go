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

func MakeNodeACPStatusCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "status [-i --identity]",
		Short: "Check the node access control status",
		Long: `Check the node access control status

Example:
  defradb client acp node status -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Note:
- If Node ACP is in clean state (not configured) the status has [IsConfigured] == false.
- If Node ACP is temporarily disabled, then [IsConfigured] == true and [IsEnabled] == false.
- If Node ACP is enabled then [IsEnabled] == true.

Learn more about the DefraDB [ACP System](/acp/README.md)

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			status, err := cliClient.GetNACStatus(cmd.Context())
			if err != nil {
				return err
			}
			return writeJSON(cmd, status)
		},
	}
	return cmd
}
