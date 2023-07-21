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

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
)

func MakePeerIDCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "peerid",
		Short: "Get the PeerID of the node",
		Long:  `Get the PeerID of the node.`,
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			stdout, err := os.Stdout.Stat()
			if err != nil {
				return errors.Wrap("failed to stat stdout", err)
			}
			if !isFileInfoPipe(stdout) {
				log.FeedbackInfo(cmd.Context(), "Requesting PeerID...")
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.PeerIDPath)
			if err != nil {
				return errors.Wrap("failed to join endpoint", err)
			}

			res, err := http.Get(endpoint.String())
			if err != nil {
				return errors.Wrap("failed to request PeerID", err)
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

			if res.StatusCode == http.StatusNotFound {
				r := httpapi.ErrorResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return errors.Wrap("parsing of response failed", err)
				}
				if len(r.Errors) > 0 {
					if isFileInfoPipe(stdout) {
						b, err := json.Marshal(r.Errors[0])
						if err != nil {
							return errors.Wrap("mashalling error response failed", err)
						}
						cmd.Println(string(b))
					} else {
						log.FeedbackInfo(cmd.Context(), r.Errors[0].Message)
					}
					return nil
				}
				return errors.New("no PeerID available. P2P might be disabled")
			}

			r := httpapi.DataResponse{}
			err = json.Unmarshal(response, &r)
			if err != nil {
				return errors.Wrap("parsing of response failed", err)
			}
			if isFileInfoPipe(stdout) {
				b, err := json.Marshal(r.Data)
				if err != nil {
					return errors.Wrap("mashalling data response failed", err)
				}
				cmd.Println(string(b))
			} else if data, ok := r.Data.(map[string]any); ok {
				log.FeedbackInfo(cmd.Context(), data["peerID"].(string))
			}

			return nil
		},
	}
	return cmd
}
