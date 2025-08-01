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

func MakeNodeACPCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "node",
		Short: "Interact with the node access control system of a DefraDB node",
		Long: `Interact with the node access control system of a DefraDB node
Note:
- Clean state means, if the access control was never configured, or the entire state was purged.
- The check the node acp status use the 'client acp node status' command
- To start node acp for the first time, use the start command with '--acp-enable true'.
- Specifying an identity is a MUST, when starting first time (from a clean state), this identity
will become the node owner identity.
- To temporarily disable node acp, use the 'client acp node disable' command.
- To re-enable node acp when it is temporarily disabled, use the 'client acp node re-enable' command.
- To give node access to other users use the 'client acp node relationship add' command.
- To revoke node access from other users use the 'client acp node relationship delete' command.
- To reset/purge acp state into a clean state, use the 'client acp node purge' command.
- Purge command(s) require the user to be in dev-mode (Warning: all state will be lost).

For quick help: 'defradb client acp node --help'

Learn more about the DefraDB [ACP System](/acp/README.md)

		`,
	}

	return cmd
}
