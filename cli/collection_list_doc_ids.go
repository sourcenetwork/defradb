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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/http"
)

func MakeCollectionListDocIDsCommand() *cobra.Command {
	const identityFlagLongRequired string = "identity"
	const identityFlagShortRequired string = "i"

	var identityValue string

	var cmd = &cobra.Command{
		Use:   "docIDs [-i --identity]",
		Short: "List all document IDs (docIDs).",
		Long: `List all document IDs (docIDs).
		
Example: list all docID(s):
  defradb client collection docIDs --name User

Example: list all docID(s), with an identity:
  defradb client collection docIDs -i cosmos1f2djr7dl9vhrk3twt3xwqp09nhtzec9mdkf70j --name User 
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO-ACP: `https://github.com/sourcenetwork/defradb/issues/2358` do the validation here.
			identity := acpIdentity.NewIdentity(identityValue)

			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			docCh, err := col.GetAllDocIDs(cmd.Context(), identity)
			if err != nil {
				return err
			}
			for docIDResult := range docCh {
				results := &http.DocIDResult{
					DocID: docIDResult.ID.String(),
				}
				if docIDResult.Err != nil {
					results.Error = docIDResult.Err.Error()
				}
				if err := writeJSON(cmd, results); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(
		&identityValue,
		identityFlagLongRequired,
		identityFlagShortRequired,
		"",
		"Identity of the actor",
	)
	return cmd
}
