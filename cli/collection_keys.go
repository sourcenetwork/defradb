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

	"github.com/sourcenetwork/defradb/http"
)

func MakeCollectionKeysCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "keys",
		Short: "List all document keys.",
		Long: `List all document keys.
		
Example:
  defradb client collection keys --name User
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetCollectionContext(cmd)
			if !ok {
				return cmd.Usage()
			}

			docCh, err := col.GetAllDocKeys(cmd.Context())
			if err != nil {
				return err
			}
			for docKey := range docCh {
				results := &http.DocKeyResult{
					Key: docKey.Key.String(),
				}
				if docKey.Err != nil {
					results.Error = docKey.Err.Error()
				}
				if err := writeJSON(cmd, results); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}
