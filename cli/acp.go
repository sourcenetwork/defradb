// Copyright 2024 Democratized Data Foundation
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

func MakeACPCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "acp",
		Short: "Interact with the access control system of a DefraDB node",
		Long: `Interact with the access control system of a DefraDB node

Learn more about [ACP](/acp/README.md)

		`,
	}

	return cmd
}
