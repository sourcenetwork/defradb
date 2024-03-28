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
	"io"
	"os"

	"github.com/spf13/cobra"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
)

func MakeCollectionCreateCommand() *cobra.Command {
	const identityFlagLongRequired string = "identity"
	const identityFlagShortRequired string = "i"

	var identityValue string
	var file string

	var cmd = &cobra.Command{
		Use:   "create [-i --identity] <document>",
		Short: "Create a new document.",
		Long: `Create a new document.

Example: create from string:
  defradb client collection create --name User '{ "name": "Bob" }'

Example: create from string, with identity:
  defradb client collection create -i cosmos1f2djr7dl9vhrk3twt3xwqp09nhtzec9mdkf70j --name User '{ "name": "Bob" }'

Example: create multiple from string:
  defradb client collection create --name User '[{ "name": "Alice" }, { "name": "Bob" }]'

Example: create from file:
  defradb client collection create --name User -f document.json

Example: create from stdin:
  cat document.json | defradb client collection create --name User -
		`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO-ACP: `https://github.com/sourcenetwork/defradb/issues/2358` do the validation here.
			identity := acpIdentity.NewIdentity(identityValue)

			var docData []byte
			switch {
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				docData = data
			case len(args) == 1 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				docData = data
			case len(args) == 1:
				docData = []byte(args[0])
			default:
				return ErrNoDocOrFile
			}

			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			if client.IsJSONArray(docData) {
				docs, err := client.NewDocsFromJSON(docData, col.Schema())
				if err != nil {
					return err
				}
				return col.CreateMany(cmd.Context(), identity, docs)
			}

			doc, err := client.NewDocFromJSON(docData, col.Schema())
			if err != nil {
				return err
			}
			return col.Create(cmd.Context(), identity, doc)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "File containing document(s)")
	cmd.Flags().StringVarP(
		&identityValue,
		identityFlagLongRequired,
		identityFlagShortRequired,
		"",
		"Identity of the actor",
	)
	return cmd
}
