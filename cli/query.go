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
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
)

var queryCmd = &cobra.Command{
	Use:   "query [query]",
	Short: "Send a DefraDB GraphQL query",
	Long: `Send a DefraDB GraphQL query to the database.

A query can be sent as a single argument. Example command:
  defradb client query 'query { ... }'

Or it can be sent via stdin by using the '-' special syntax. Example command:
  cat query.graphql | defradb client query -

A GraphQL client such as GraphiQL (https://github.com/graphql/graphiql) can be used to interact
with the database more conveniently.

To learn more about the DefraDB GraphQL Query Language, refer to https://docs.source.network.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var query string
		inputIsPipe := stdinIsPipe()

		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}

		if inputIsPipe && (len(args) == 0 || args[0] != "-") {
			log.FeedbackInfo(
				cmd.Context(),
				"Run 'defradb client query -' to read from stdin. Example: 'cat my.graphql | defradb client query -').",
			)
			return nil
		} else if len(args) == 0 {
			err := cmd.Help()
			if err != nil {
				return fmt.Errorf("failed to print help: %w", err)
			}
			return nil
		} else if args[0] == "-" {
			stdin, err := readStdin()
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			if len(stdin) == 0 {
				return fmt.Errorf("no query in stdin provided")
			} else {
				query = stdin
			}
		} else {
			query = args[0]
		}

		if query == "" {
			return fmt.Errorf("query cannot be empty")
		}

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.GraphQLPath)
		if err != nil {
			return fmt.Errorf("joining paths failed: %w", err)
		}

		p := url.Values{}
		p.Add("query", query)
		endpoint.RawQuery = p.Encode()

		res, err := http.Get(endpoint.String())
		if err != nil {
			return fmt.Errorf("failed to send query: %w", err)
		}

		defer func() {
			err = res.Body.Close()
			if err != nil {
				log.ErrorE(cmd.Context(), "response body closing failed: ", err)
			}
		}()

		buf, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		indentedResult, err := indentJSON(buf)
		if err != nil {
			return fmt.Errorf("failed to pretty print result: %w", err)
		}
		log.FeedbackInfo(cmd.Context(), indentedResult)
		return nil
	},
}

func init() {
	clientCmd.AddCommand(queryCmd)
}
