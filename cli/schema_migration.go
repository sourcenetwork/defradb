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
	"github.com/spf13/cobra"
)

func MakeSchemaMigrationCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "migration",
		Short: "Interact with the schema migration system of a running DefraDB instance",
		Long:  `Make set or look for existing schema migrations on a DefraDB node.`,
	}

	return cmd
}
