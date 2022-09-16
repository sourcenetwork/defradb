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
	"os"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/errors"
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var query string

		fi, err := os.Stdin.Stat()
		if err != nil {
			return err
		}

		if len(args) > 1 {
			if err = cmd.Usage(); err != nil {
				return err
			}
			return errors.New("too many arguments")
		}

		if isFileInfoPipe(fi) && (len(args) == 0 || args[0] != "-") {
			log.FeedbackInfo(
				cmd.Context(),
				"Run 'defradb client query -' to read from stdin. Example: 'cat my.graphql | defradb client query -').",
			)
			return nil
		} else if len(args) == 0 {
			err := cmd.Help()
			if err != nil {
				return errors.Wrap("failed to print help", err)
			}
			return nil
		} else if args[0] == "-" {
			stdin, err := readStdin()
			if err != nil {
				return errors.Wrap("failed to read stdin", err)
			}
			if len(stdin) == 0 {
				return errors.New("no query in stdin provided")
			} else {
				query = stdin
			}
		} else {
			query = args[0]
		}

		if query == "" {
			return errors.New("query cannot be empty")
		}

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.GraphQLPath)
		if err != nil {
			return errors.Wrap("joining paths failed", err)
		}

		p := url.Values{}
		p.Add("query", query)
		endpoint.RawQuery = p.Encode()

		res, err := http.Get(endpoint.String())
		if err != nil {
			return errors.Wrap("failed query", err)
		}

		defer func() {
			if e := res.Body.Close(); e != nil {
				err = errors.Wrap(fmt.Sprintf("failed to read response body: %v", e.Error()), err)
			}
		}()

		response, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap("failed to read response body", err)
		}

		fi, err = os.Stdout.Stat()
		if err != nil {
			return errors.Wrap("failed to stat stdout", err)
		}

		if isFileInfoPipe(fi) {
			cmd.Println(string(response))
		} else {
			graphlErr, err := hasGraphQLErrors(response)
			if err != nil {
				return errors.Wrap("failed to handle GraphQL errors", err)
			}
			indentedResult, err := indentJSON(response)
			if err != nil {
				return errors.Wrap("failed to pretty print result", err)
			}
			if graphlErr {
				log.FeedbackError(cmd.Context(), indentedResult)
			} else {
				log.FeedbackInfo(cmd.Context(), indentedResult)
			}
		}
		return nil
	},
}

func init() {
	clientCmd.AddCommand(queryCmd)
}
