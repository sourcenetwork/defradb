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
	"net/url"
	"os"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeIndexCreateCommand(cfg *config.Config) *cobra.Command {
	var collectionArg string
	var nameArg string
	var fieldsArg string
	var cmd = &cobra.Command{
		Use:   "create -c --collection <collection> --fields <fields> [-n --name <name>]",
		Short: "Creates a secondary index on a collection's field(s)",
		Long: `Creates a secondary index on a collection's field(s).
		
The --name flag is optional. If not provided, a name will be generated automatically.

Example: create an index for 'Users' collection on 'Name' field:
  defradb client index create --collection 'Users' --fields 'Name'

Example: create a named index for 'Users' collection on 'Name' field:
  defradb client index create --collection Users --fields Name --name UsersByName`,
		ValidArgs: []string{"collection", "fields", "name"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if collectionArg == "" || fieldsArg == "" {
				if collectionArg == "" {
					return NewErrMissingArg("collection")
				} else {
					return NewErrMissingArg("fields")
				}
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.IndexCreatePath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			values := url.Values{
				"collection": {collectionArg},
				"fields":     {fieldsArg},
			}
			if nameArg != "" {
				values.Add("name", nameArg)
			}

			res, err := http.PostForm(endpoint.String(), values)
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}

			defer func() {
				if e := res.Body.Close(); e != nil {
					err = NewErrFailedToCloseResponseBody(err)
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
				type schemaResponse struct {
					Data struct {
						Index struct {
							Name   string `json:"name"`
							ID     uint32 `json:"id"`
							Fields []struct {
								Name      string `json:"name"`
								Direction string `json:"direction"`
							} `json:"fields"`
						} `json:"index"`
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
					log.FeedbackError(cmd.Context(), "Failed to create index",
						logging.NewKV("Errors", r.Errors))
				} else {
					log.FeedbackInfo(cmd.Context(), "Successfully created index",
						logging.NewKV("Index", r.Data.Index))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVarP(&nameArg, "name", "n", "", "Index name")
	cmd.Flags().StringVar(&fieldsArg, "fields", "", "Fields to index")

	return cmd
}
