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
	"github.com/sourcenetwork/defradb/errors"
)

var patchFile string

var patchCmd = &cobra.Command{
	Use:   "patch [schema]",
	Short: "Update an existing schema type",
	Long: `Update an existing schema.

Uses JSON PATCH formatting as an update DDL.

Example: update from an argument string:
  defradb client schema patch '[{ "op": "add", "path": "...", "value": {...} }]'

Example: update from file:
  defradb client schema patch -f patch.json

Example: update from stdin:
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
			return errors.New("too many arguments")
		}

		if patchFile != "" {
			buf, err := os.ReadFile(patchFile)
			if err != nil {
				return errors.Wrap("failed to read patch file", err)
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
				return errors.New("no patch in stdin provided")
			} else {
				patch = stdin
			}
		} else {
			patch = args[0]
		}

		if patch == "" {
			return errors.New("empty patch provided")
		}

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.SchemaLoadPath)
		if err != nil {
			return errors.Wrap("join paths failed", err)
		}

		res, err := http.Post(endpoint.String(), "text", strings.NewReader(patch))
		if err != nil {
			return errors.Wrap("failed to post patch", err)
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
						Result string `json:"result"`
					} `json:"data"`
				}
				r := schemaResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return errors.Wrap("failed to unmarshal response", err)
				}
				log.FeedbackInfo(cmd.Context(), r.Data.Result)
			}
		}
		return nil
	},
}

func init() {
	schemaCmd.AddCommand(patchCmd)
	patchCmd.Flags().StringVarP(&patchFile, "file", "f", "", "File to load a patch from")
}
