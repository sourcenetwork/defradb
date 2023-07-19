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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeSchemaAddCommand(cfg *config.Config) *cobra.Command {
	var schemaFile string
	var cmd = &cobra.Command{
		Use:   "add [schema]",
		Short: "Add a new schema type to DefraDB",
		Long: `Add a new schema type to DefraDB.

Example: add from an argument string:
  defradb client schema add 'type Foo { ... }'

Example: add from file:
  defradb client schema add -f schema.graphql

Example: add from stdin:
  cat schema.graphql | defradb client schema add -

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var schema string
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

			if schemaFile != "" {
				buf, err := os.ReadFile(schemaFile)
				if err != nil {
					return errors.Wrap("failed to read schema file", err)
				}
				schema = string(buf)
			} else if isFileInfoPipe(fi) && (len(args) == 0 || args[0] != "-") {
				log.FeedbackInfo(
					cmd.Context(),
					"Run 'defradb client schema add -' to read from stdin."+
						" Example: 'cat schema.graphql | defradb client schema add -').",
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
					return errors.New("no schema in stdin provided")
				} else {
					schema = stdin
				}
			} else {
				schema = args[0]
			}

			if schema == "" {
				return errors.New("empty schema provided")
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.SchemaPath)
			if err != nil {
				return errors.Wrap("join paths failed", err)
			}

			res, err := http.Post(endpoint.String(), "text", strings.NewReader(schema))
			if err != nil {
				return errors.Wrap("failed to post schema", err)
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

			stdout, err := os.Stdout.Stat()
			if err != nil {
				return errors.Wrap("failed to stat stdout", err)
			}
			if isFileInfoPipe(stdout) {
				cmd.Println(string(response))
			} else {
				graphlErr, err := hasGraphQLErrors(response)
				if err != nil {
					return errors.Wrap("failed to handle GraphQL errors", err)
				}
				if graphlErr {
					indentedResult, err := indentJSON(response)
					if err != nil {
						return errors.Wrap("failed to pretty print result", err)
					}
					log.FeedbackError(cmd.Context(), indentedResult)
				} else {
					type schemaResponse struct {
						Data struct {
							Result      string `json:"result"`
							Collections []struct {
								Name string `json:"name"`
								ID   string `json:"id"`
							} `json:"collections"`
						} `json:"data"`
					}
					r := schemaResponse{}
					err = json.Unmarshal(response, &r)
					if err != nil {
						return errors.Wrap("failed to unmarshal response", err)
					}
					if r.Data.Result == "success" {
						log.FeedbackInfo(cmd.Context(), "Successfully added schema.", logging.NewKV("Collections", r.Data.Collections))
					}
					log.FeedbackInfo(cmd.Context(), r.Data.Result)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&schemaFile, "file", "f", "", "File to load a schema from")
	return cmd
}
