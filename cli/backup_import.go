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

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeBackupImportCommand(cfg *config.Config) *cobra.Command {
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
			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.ImportPath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			inputPath := args[0]
			data := map[string]string{
				"filepath": inputPath,
			}

			b, err := json.Marshal(data)
			if err != nil {
				return err
			}

			res, err := http.Post(endpoint.String(), "application/json", bytes.NewBuffer(b))
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}

			defer func() {
				if e := res.Body.Close(); e != nil {
					err = NewErrFailedToCloseResponseBody(e, err)
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
