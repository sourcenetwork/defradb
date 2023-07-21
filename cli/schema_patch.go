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
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
)

func MakeSchemaPatchCommand(cfg *config.Config) *cobra.Command {
	var patchFile string

	var cmd = &cobra.Command{
		Use:   "patch [schema]",
		Short: "Patch an existing schema type",
		Long: `Patch an existing schema.

Uses JSON Patch to modify schema types.

Example: patch from an argument string:
  defradb client schema patch '[{ "op": "add", "path": "...", "value": {...} }]'

Example: patch from file:
  defradb client schema patch -f patch.json

Example: patch from stdin:
  cat patch.json | defradb client schema patch -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var patch string
			fi, err := os.Stdin.Stat()
			if err != nil {
				return err
			}

			if len(args) > 1 {
				if err = cmd.Usage(); err != nil {
					return err
				}
				return ErrTooManyArgs
			}

			if patchFile != "" {
				buf, err := os.ReadFile(patchFile)
				if err != nil {
					return NewFailedToReadFile(err)
				}
				patch = string(buf)
			} else if isFileInfoPipe(fi) && (len(args) == 0 || args[0] != "-") {
				log.FeedbackInfo(
					cmd.Context(),
					"Run 'defradb client schema patch -' to read from stdin."+
						" Example: 'cat patch.json | defradb client schema patch -').",
				)
				return nil
			} else if len(args) == 0 {
				// ignore error, nothing we can do about it
				// as printing an error about failing to print help
				// is useless
				//nolint:errcheck
				cmd.Help()
				return nil
			} else if args[0] == "-" {
				stdin, err := readStdin()
				if err != nil {
					return NewFailedToReadStdin(err)
				}
				if len(stdin) == 0 {
					return ErrEmptyStdin
				} else {
					patch = stdin
				}
			} else {
				patch = args[0]
			}

			if patch == "" {
				return ErrEmptyFile
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.SchemaPath)
			if err != nil {
				return err
			}

			req, err := http.NewRequest(http.MethodPatch, endpoint.String(), strings.NewReader(patch))
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}

			//nolint:errcheck
			defer res.Body.Close()
			response, err := io.ReadAll(res.Body)
			if err != nil {
				return NewErrFailedToReadResponseBody(err)
			}

			stdout, err := os.Stdout.Stat()
			if err != nil {
				return NewErrFailedToStatStdOut(err)
			}
			if isFileInfoPipe(stdout) {
				cmd.Println(string(response))
			} else {
				graphlErr, err := hasGraphQLErrors(response)
				if err != nil {
					return NewErrFailedToHandleGQLErrors(err)
				}
				if graphlErr {
					indentedResult, err := indentJSON(response)
					if err != nil {
						return NewErrFailedToPrettyPrintResponse(err)
					}
					log.FeedbackError(cmd.Context(), indentedResult)
				} else {
					type schemaResponse struct {
						Data struct {
							Result string `json:"result"`
						} `json:"data"`
					}
					r := schemaResponse{}
					err = json.Unmarshal(response, &r)
					if err != nil {
						return NewErrFailedToUnmarshalResponse(err)
					}
					log.FeedbackInfo(cmd.Context(), r.Data.Result)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&patchFile, "file", "f", "", "File to load a patch from")
	return cmd
}
