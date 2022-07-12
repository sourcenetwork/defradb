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
	"fmt"
	"io"
	"net/http"
	"os"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [CID]",
	Short: "Get a block by its CID from the blockstore.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("get requires a CID argument")
		}
		cid := args[0]

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.BlocksPath, cid)
		if err != nil {
			return fmt.Errorf("failed to join endpoint: %w", err)
		}

		res, err := http.Get(endpoint.String())
		if err != nil {
			return fmt.Errorf("failed to send get request: %w", err)
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

		stdout, err := os.Stdout.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat stdout: %w", err)
		}
		if isFileInfoPipe(stdout) {
			cmd.Println(string(response))
		} else {
			graphlErr, err := hasGraphQLErrors(response)
			if err != nil {
				return fmt.Errorf("failed to handle GraphQL errors: %w", err)
			}
			indentedResult, err := indentJSON(response)
			if err != nil {
				return fmt.Errorf("failed to pretty print response: %w", err)
			}
			if graphlErr {
				log.FeedbackError(cmd.Context(), indentedResult)
			} else {
				log.FeedbackInfo(cmd.Context(), indentedResult)
			}
		}
		return nil
	},
}

func init() {
	blocksCmd.AddCommand(getCmd)
}
