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

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
)

func MakeDumpCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump the contents of DefraDB node-side",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			stdout, err := os.Stdout.Stat()
			if err != nil {
				return errors.Wrap("failed to stat stdout", err)
			}
			if !isFileInfoPipe(stdout) {
				log.FeedbackInfo(cmd.Context(), "Requesting the database to dump its state, server-side...")
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.DumpPath)
			if err != nil {
				return errors.Wrap("failed to join endpoint", err)
			}

			res, err := http.Get(endpoint.String())
			if err != nil {
				return errors.Wrap("failed dump request", err)
			}

			defer func() {
				if e := res.Body.Close(); e != nil {
					err = NewErrFailedToCloseResponseBody(e, err)
				}
			}()

			response, err := io.ReadAll(res.Body)
			if err != nil {
				return errors.Wrap("failed to read response body", err)
			}

			if isFileInfoPipe(stdout) {
				cmd.Println(string(response))
			} else {
				// dumpResponse follows structure of HTTP API's response
				type dumpResponse struct {
					Data struct {
						Response string `json:"response"`
					} `json:"data"`
				}
				r := dumpResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return errors.Wrap("failed parsing of response", err)
				}
				log.FeedbackInfo(cmd.Context(), r.Data.Response)
			}
			return nil
		},
	}
	return cmd
}
