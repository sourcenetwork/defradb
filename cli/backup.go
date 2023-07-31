// Copyright 2022 Democratized Data Foundation
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

func MakeBackupCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "backup",
		Short: "Interact with the backup utility",
		Long: `Export to or Import from a backup file.
Currently only supports JSON format.`,
	}
	return cmd
}
