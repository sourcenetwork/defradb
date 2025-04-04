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

func MakeBlockVerifySignatureCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "verify-signature",
		Short: "Verify the signature of a block",
		Long: `Verify the signature of a block.
		
Notes:
  - The identity must be specified.

Example to verify the signature of a block:
  defradb client block verify-signature -i <identity> <cid>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return NewErrMissingRequiredParameter("cid")
			}

			db := mustGetContextDB(cmd)
			err := db.VerifyBlock(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			_, err = out.Write([]byte("Block's signature verified\n"))
			return err
		},
	}
	return cmd
}
