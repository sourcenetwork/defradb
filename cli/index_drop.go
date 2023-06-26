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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeIndexDropCommand(cfg *config.Config) *cobra.Command {
	var collectionArg string
	var nameArg string
	var cmd = &cobra.Command{
		Use:   "drop -c --collection <collection> -n --name <name>",
		Short: "Drop a collection's secondary index",
		Long: `Drop a collection's secondary index.
		
Example: drop the index 'UsersByName' for 'Users' collection:
  defradb client index create --collection Users --name UsersByName`,
		ValidArgs: []string{"collection", "name"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.IndexDropPath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			values := url.Values{
				"collection": {collectionArg},
				"name":     {nameArg},
			}
			res, err := http.PostForm(endpoint.String(), values)
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}

			defer res.Body.Close()
			response, err := io.ReadAll(res.Body)
			if err != nil {
				return NewErrFailedToReadResponseBody(err)
			}

			stdout, err := os.Stdout.Stat()
			if err != nil {
				return err
			}

			if !isFileInfoPipe(stdout) {
				type schemaResponse struct {
					Data struct {
						Result bool `json:"result"`
					} `json:"data"`
					Errors []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				r := schemaResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return NewErrFailedToUnmarshalResponse(err)
				}
				if len(r.Errors) > 0 {
					log.FeedbackError(cmd.Context(), "Failed to create index.",
						logging.NewKV("Errors", r.Errors))
					log.FeedbackInfo(cmd.Context(), "success")
				} else {
					log.FeedbackInfo(cmd.Context(), "Successfully created index.",
						logging.NewKV("Result", r.Data.Result))
					log.FeedbackInfo(cmd.Context(), "failure")
				}
			}
			cmd.Println(string(response))
			return nil
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVarP(&nameArg, "name", "n", "", "Index name")

	err := cmd.MarkFlagRequired("collection")
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not mark 'collection' as required argument", err)
	}
	err = cmd.MarkFlagRequired("name")
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not mark 'fields' as required argument", err)
	}

	return cmd
}
