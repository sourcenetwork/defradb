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
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	REQ_RESULTS_HEADER = "------ Request Results ------\n"
	SUB_RESULTS_HEADER = "------ Subscription Results ------\n"
)

func MakeRequestCommand(ctx context.Context) *cobra.Command {
	var filePath string
	var operationName string
	var variablesJSON string
	var cmd = &cobra.Command{
		Use:   "query [-i --identity] [request]",
		Short: "Send a DefraDB GraphQL query request",
		Long: `Send a DefraDB GraphQL query request to the database.

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
				request = args[0]
			}

			if request == "" {
				return errors.New("request cannot be empty")
			}

			var options []client.RequestOption
			if variablesJSON != "" {
				var variables map[string]any
				err := json.Unmarshal([]byte(variablesJSON), &variables)
				if err != nil {
					return err
				}
				options = append(options, client.WithVariables(variables))
			}
			if operationName != "" {
				options = append(options, client.WithOperationName(operationName))
			}

			cliClient := mustGetContextCLIClient(cmd)
			result := cliClient.ExecRequest(cmd.Context(), request, options...)

			if result.Subscription == nil {
				cmd.Print(REQ_RESULTS_HEADER)
				return writeJSON(cmd, result.GQL)
			}
			cmd.Print(SUB_RESULTS_HEADER)
			for item := range result.Subscription {
				writeJSON(cmd, item) //nolint:errcheck
			}
			return nil
		},
	}

	EmbedCLIExample(ctx, cmd, "Single argument query",
		`defradb client query 'query { ... }'`)

	EmbedCLIExample(ctx, cmd, "Query from file",
		`defradb client query -f request.graphql`)

	EmbedCLIExample(ctx, cmd, "Query from file, with a provided identity",
		`defradb client query -f request.graphql \
	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f `)

	EmbedCLIExample(ctx, cmd, "Read query from stdin",
		`cat request.graphql | defradb client query -`)

	cmd.Flags().StringVarP(&operationName, "operation", "o", "", "Name of the operation to execute in the query")
	cmd.Flags().StringVarP(&variablesJSON, "variables", "v", "", "JSON encoded variables to use in the query")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File containing the query request")
	return cmd
}
