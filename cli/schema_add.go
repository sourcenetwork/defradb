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
	"os"
	"strings"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
)

var schemaFile string

var addCmd = &cobra.Command{
	Use:   "add [schema]",
	Short: "Add a new schema type to DefraDB",
	Long: `Add a new schema type to DefraDB.

Example: add as an argument string:
  defradb client schema add 'type Foo { ... }'

Example: add from file:
  defradb client schema add -f schema.graphql

Example: add from stdin:
  cat schema.graphql | defradb client schema add -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var schema string
		inputIsPipe, err := stdinIsPipe()
		if err != nil {
			return err
		}

		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}

		if schemaFile != "" {
			buf, err := os.ReadFile(schemaFile)
			if err != nil {
				return fmt.Errorf("failed to read schema file: %w", err)
			}
			schema = string(buf)
		} else if inputIsPipe && (len(args) == 0 || args[0] != "-") {
			log.FeedbackInfo(
				cmd.Context(),
				"Run 'defradb client schema add -' to read from stdin."+
					" Example: 'cat schema.graphql | defradb client schema add -').",
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
				return fmt.Errorf("no schema in stdin provided")
			} else {
				schema = stdin
			}
		} else {
			schema = args[0]
		}

		if schema == "" {
			return fmt.Errorf("empty schema provided")
		}

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.SchemaLoadPath)
		if err != nil {
			return fmt.Errorf("join paths failed: %w", err)
		}

		res, err := http.Post(endpoint.String(), "text", strings.NewReader(schema))
		if err != nil {
			return fmt.Errorf("failed to post schema: %w", err)
		}

		defer func() {
			err = res.Body.Close()
			if err != nil {
				log.ErrorE(cmd.Context(), "response body closing failed", err)
			}
		}()

		result, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		indentedResult, err := indentJSON(result)
		if err != nil {
			return fmt.Errorf("failed to pretty print result: %w", err)
		}
		log.FeedbackInfo(cmd.Context(), indentedResult)
		return nil
	},
}

func init() {
	schemaCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&schemaFile, "file", "f", "", "file to load a schema from")
}
