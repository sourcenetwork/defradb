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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeSchemaMigrationGetCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Gets the schema migrations within DefraDB",
		Long: `Gets the schema migrations within the local DefraDB node.

Example:
  defradb client schema migration get'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.NoArgs(cmd, args); err != nil {
				return errors.New("this command take no arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.SchemaMigrationPath)
			if err != nil {
				return errors.Wrap("join paths failed", err)
			}

			res, err := http.Get(endpoint.String())
			if err != nil {
				return errors.Wrap("failed to get schema migrations", err)
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
				type migrationGetResponse struct {
					Data struct {
						Configuration []client.LensConfig `json:"configuration"`
					} `json:"data"`
					Errors []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				r := migrationGetResponse{}
				err = json.Unmarshal(response, &r)
				log.FeedbackInfo(cmd.Context(), string(response))
				if err != nil {
					return NewErrFailedToUnmarshalResponse(err)
				}
				if len(r.Errors) > 0 {
					log.FeedbackError(cmd.Context(), "Failed to get schema migrations",
						logging.NewKV("Errors", r.Errors))
				} else {
					log.FeedbackInfo(cmd.Context(), "Successfully got schema migrations",
						logging.NewKV("Configuration", r.Data.Configuration))
				}
			}

			return nil
		},
	}
	return cmd
}
