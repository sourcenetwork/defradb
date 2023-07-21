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
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
)

func MakeBlocksGetCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get [CID]",
		Short: "Get a block by its CID from the blockstore",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return NewErrMissingArg("CID")
			}
			cid := args[0]

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.BlocksPath, cid)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			res, err := http.Get(endpoint.String())
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}

			defer func() {
				if e := res.Body.Close(); e != nil {
					err = NewErrFailedToReadResponseBody(err)
				}
			}()

			response, err := io.ReadAll(res.Body)
			if err != nil {
				return NewErrFailedToReadResponseBody(err)
			}

			stdout, err := os.Stdout.Stat()
			if err != nil {
				return NewErrFailedToStatStdOut(err)
			}
			if isFileInfoPipe(stdout) {
				cmd.Println(string(response))
			} else {
				graphlErr, err := hasGraphQLErrors(response)
				if err != nil {
					return NewErrFailedToHandleGQLErrors(err)
				}
				indentedResult, err := indentJSON(response)
				if err != nil {
					return NewErrFailedToPrettyPrintResponse(err)
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
	return cmd
}
