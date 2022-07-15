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

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping defradb to test an API connection",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		stdout, err := os.Stdout.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat stdout: %w", err)
		}
		if !isFileInfoPipe(stdout) {
			log.FeedbackInfo(cmd.Context(), "Sending ping...")
		}

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.PingPath)
		if err != nil {
			return fmt.Errorf("failed to join endpoint: %w", err)
		}

		res, err := http.Get(endpoint.String())
		if err != nil {
			return fmt.Errorf("failed to send ping: %w", err)
		}

		defer func() {
			if e := res.Body.Close(); e != nil {
				err = fmt.Errorf("failed to read response body: %v: %w", e.Error(), err)
			}
		}()

		response, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
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
				return fmt.Errorf("parsing of response failed: %w", err)
			}
			log.FeedbackInfo(cmd.Context(), r.Data.Response)
		}
		return err
	},
}

func init() {
	clientCmd.AddCommand(pingCmd)
}
