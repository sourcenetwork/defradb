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
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

type indexListResponse struct {
	Data struct {
		Collections map[string][]client.IndexDescription `json:"collections"`
		Indexes     []client.IndexDescription            `json:"indexes"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func MakeIndexListCommand(cfg *config.Config) *cobra.Command {
	var collectionArg string
	var cmd = &cobra.Command{
		Use:   "list [-c --collection <collection>]",
		Short: "Shows the list indexes in the database or for a specific collection",
		Long: `Shows the list indexes in the database or for a specific collection.
		
If the --collection flag is provided, only the indexes for that collection will be shown.
Otherwise, all indexes in the database will be shown.

Example: show all index for 'Users' collection:
  defradb client index list --collection Users`,
		ValidArgs: []string{"collection"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.IndexPath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			if collectionArg != "" {
				values := url.Values{
					"collection": {collectionArg},
				}
				endpoint.RawQuery = values.Encode()
			}

			res, err := http.Get(endpoint.String())
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}

			defer func() {
				if e := res.Body.Close(); e != nil {
					err = NewErrFailedToCloseResponseBody(e, err)
				}
			}()

			response, err := io.ReadAll(res.Body)
			if err != nil {
				return NewErrFailedToReadResponseBody(err)
			}

			stdout, err := os.Stdout.Stat()
			if err != nil {
				return err
			}

			if isFileInfoPipe(stdout) {
				cmd.Println(string(response))
			} else {
				r := indexListResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return NewErrFailedToUnmarshalResponse(err)
				}
				if len(r.Errors) > 0 {
					log.FeedbackError(cmd.Context(), "Failed to list index",
						logging.NewKV("Errors", r.Errors))
				} else if collectionArg != "" {
					log.FeedbackInfo(cmd.Context(), "Fetched indexes for collection "+collectionArg,
						logging.NewKV("Indexes", r.Data.Indexes))
				} else {
					log.FeedbackInfo(cmd.Context(), "Fetched all indexes",
						logging.NewKV("Collections", r.Data.Collections))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")

	return cmd
}
