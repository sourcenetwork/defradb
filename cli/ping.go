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

func MakePingCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping to test connection with a node",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			stdout, err := os.Stdout.Stat()
			if err != nil {
				return errors.Wrap("failed to stat stdout", err)
			}
			if !isFileInfoPipe(stdout) {
				log.FeedbackInfo(cmd.Context(), "Sending ping...")
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.PingPath)
			if err != nil {
				return errors.Wrap("failed to join endpoint", err)
			}

			res, err := http.Get(endpoint.String())
			if err != nil {
				return errors.Wrap("failed to send ping", err)
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
				type pingResponse struct {
					Data struct {
						Response string `json:"response"`
					} `json:"data"`
				}
				r := pingResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return errors.Wrap("parsing of response failed", err)
				}
				log.FeedbackInfo(cmd.Context(), r.Data.Response)
			}
			return nil
		},
	}
	return cmd
}
