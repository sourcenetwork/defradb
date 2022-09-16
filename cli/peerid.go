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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/spf13/cobra"
)

var peerIDCmd = &cobra.Command{
	Use:   "peerid",
	Short: "Get the peer ID of the DefraDB node",
	RunE: func(cmd *cobra.Command, _ []string) (err error) {
		stdout, err := os.Stdout.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat stdout: %w", err)
		}
		if !isFileInfoPipe(stdout) {
			log.FeedbackInfo(cmd.Context(), "Requesting peer ID...")
		}

		endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.PeerIDPath)
		if err != nil {
			return fmt.Errorf("failed to join endpoint: %w", err)
		}

		res, err := http.Get(endpoint.String())
		if err != nil {
			return fmt.Errorf("failed to request peer ID: %w", err)
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

		if res.StatusCode == http.StatusNotFound {
			r := httpapi.ErrorResponse{}
			err = json.Unmarshal(response, &r)
			if err != nil {
				return fmt.Errorf("parsing of response failed: %w", err)
			}
			if len(r.Errors) > 0 {
				if isFileInfoPipe(stdout) {
					b, err := json.Marshal(r.Errors[0])
					if err != nil {
						return fmt.Errorf("mashalling error response failed: %w", err)
					}
					cmd.Println(string(b))
				} else {
					log.FeedbackInfo(cmd.Context(), r.Errors[0].Message)
				}
				return nil
			}
			return errors.New("no peer ID available. P2P might be disabled")
		}

		r := httpapi.DataResponse{}
		err = json.Unmarshal(response, &r)
		if err != nil {
			return fmt.Errorf("parsing of response failed: %w", err)
		}
		if isFileInfoPipe(stdout) {
			b, err := json.Marshal(r.Data)
			if err != nil {
				return fmt.Errorf("mashalling data response failed: %w", err)
			}
			cmd.Println(string(b))
		} else if data, ok := r.Data.(map[string]any); ok {
			log.FeedbackInfo(cmd.Context(), data["peerID"].(string))
		}

		return nil
	},
}

func init() {
	clientCmd.AddCommand(peerIDCmd)
}
