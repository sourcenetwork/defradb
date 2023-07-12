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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeDBImportCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "import <input_path>",
		Short: "Import a JSON data file to the database",
		Long: `Import a JSON data file to the database.

Example: import data to the database:
  defradb client import user_data.json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return NewErrInvalidArgumentLength(err, 1)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			contentType := "application/json"
			inputPath := args[0]
			if strings.HasSuffix(inputPath, "cbor") {
				contentType = "application/octet-stream"
			}

			b, err := os.ReadFile(inputPath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.ImportPath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			res, err := http.Post(endpoint.String(), contentType, bytes.NewBuffer(b))
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}

			response, err := io.ReadAll(res.Body)
			if err != nil {
				return NewErrFailedToReadResponseBody(err)
			}
			if err := res.Body.Close(); err != nil {
				return NewErrFailedToCloseResponseBody(err)
			}

			stdout, err := os.Stdout.Stat()
			if err != nil {
				return err
			}

			if isFileInfoPipe(stdout) {
				cmd.Println(string(response))
			} else {
				r := indexCreateResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return NewErrFailedToUnmarshalResponse(err)
				}
				if len(r.Errors) > 0 {
					log.FeedbackError(cmd.Context(), "Failed to import data",
						logging.NewKV("Errors", r.Errors))
				} else {
					log.FeedbackInfo(cmd.Context(), "Successfully imported data from file",
						logging.NewKV("File", inputPath))
				}
			}
			return nil
		},
	}
	return cmd
}
