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
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	REQ_RESULTS_HEADER = "------ Request Results ------\n"
	SUB_RESULTS_HEADER = "------ Subscription Results ------\n"
)

func MakeRequestCommand() *cobra.Command {
	var filePath string
	var shouldEncrypt bool
	var cmd = &cobra.Command{
		Use:   "query [-i --identity] [request]",
		Short: "Send a DefraDB GraphQL query request",
		Long: `Send a DefraDB GraphQL query request to the database.

Options:
    -i, --identity 
        Marks the document as private and set the identity as the owner. The access to the document
		and permissions are controlled by ACP (Access Control Policy).

	-e, --encrypt
		Encrypt flag specified if the document needs to be encrypted. If set, DefraDB will generate a
		symmetric key for encryption using AES-GCM.

A query request can be sent as a single argument. Example command:
  defradb client query 'query { ... }'

Do a query request from a file by using the '-f' flag. Example command:
  defradb client query -f request.graphql

Do a query request from a file and with an identity. Example command:
  defradb client query -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f -f request.graphql

Or it can be sent via stdin by using the '-' special syntax. Example command:
  cat request.graphql | defradb client query -

A GraphQL client such as GraphiQL (https://github.com/graphql/graphiql) can be used to interact
with the database more conveniently.

To learn more about the DefraDB GraphQL Query Language, refer to https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var request string
			switch {
			case filePath != "":
				data, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}
				request = string(data)
			case len(args) > 0 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				request = string(data)
			case len(args) > 0:
				request = string(args[0])
			}

			if request == "" {
				return errors.New("request cannot be empty")
			}

			store := mustGetContextStore(cmd)
			setContextDocEncryption(cmd, shouldEncrypt, nil)
			result := store.ExecRequest(cmd.Context(), request)

			var errors []string
			for _, err := range result.GQL.Errors {
				errors = append(errors, err.Error())
			}
			if result.Subscription == nil {
				cmd.Print(REQ_RESULTS_HEADER)
				return writeJSON(cmd, map[string]any{"data": result.GQL.Data, "errors": errors})
			}
			cmd.Print(SUB_RESULTS_HEADER)
			for item := range result.Subscription {
				writeJSON(cmd, item) //nolint:errcheck
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&shouldEncrypt, "encrypt", "e", false,
		"Flag to enable encryption of the document")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File containing the query request")
	return cmd
}
