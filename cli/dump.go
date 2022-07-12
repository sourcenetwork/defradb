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

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps the state of the entire database (server-side)",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.FeedbackInfo(cmd.Context(), "Requesting the database to dump its state, server-side...")

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.DumpPath)
		if err != nil {
			return fmt.Errorf("failed to join endpoint: %w", err)
		}

		res, err := http.Get(endpoint.String())
		if err != nil {
			return fmt.Errorf("failed dump request: %w", err)
		}

		defer func() {
			err = res.Body.Close()
			if err != nil {
				log.ErrorE(cmd.Context(), "Response body closing failed", err)
			}
		}()

		response, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		// dumpResponse follows structure of HTTP API's response
		type dumpResponse struct {
			Data struct {
				Response string `json:"response"`
			} `json:"data"`
		}
		r := dumpResponse{}
		err = json.Unmarshal(response, &r)
		if err != nil {
			return fmt.Errorf("failed parsing of response: %w", err)
		}
		log.FeedbackInfo(cmd.Context(), r.Data.Response)
		return nil
	},
}

func init() {
	clientCmd.AddCommand(dumpCmd)
}
