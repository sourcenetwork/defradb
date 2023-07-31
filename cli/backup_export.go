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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

const jsonFileType = "json"

func MakeBackupExportCommand(cfg *config.Config) *cobra.Command {
	var collections []string
	var pretty bool
	var format string
	var cmd = &cobra.Command{
		Use:   "export  [-c --collections | -p --pretty | -f --format] <output_path>",
		Short: "Export the database to a file",
		Long: `Export the database to a file. If a file exists at the <output_path> location, it will be overwritten.
		
If the --collection flag is provided, only the data for that collection will be exported.
Otherwise, all collections in the database will be exported.

If the --pretty flag is provided, the JSON will be pretty printed.

Example: export data for the 'Users' collection:
  defradb client export --collection Users user_data.json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return NewErrInvalidArgumentLength(err, 1)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if !isValidExportFormat(format) {
				return ErrInvalidExportFormat
			}
			outputPath := args[0]
			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.ExportPath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			for i := range collections {
				collections[i] = strings.Trim(collections[i], " ")
			}

			data := client.BackupConfig{
				Filepath:    outputPath,
				Format:      format,
				Pretty:      pretty,
				Collections: collections,
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
				type exportResponse struct {
					Errors []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				r := exportResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return NewErrFailedToUnmarshalResponse(err)
				}
				if len(r.Errors) > 0 {
					log.FeedbackError(cmd.Context(), "Failed to export data",
						logging.NewKV("Errors", r.Errors))
				} else if len(collections) == 1 {
					log.FeedbackInfo(cmd.Context(), "Data exported for collection "+collections[0])
				} else if len(collections) > 1 {
					log.FeedbackInfo(cmd.Context(), "Data exported for collections "+strings.Join(collections, ", "))
				} else {
					log.FeedbackInfo(cmd.Context(), "Data exported for all collections")
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&pretty, "pretty", "p", false, "Set the output JSON to be pretty printed")
	cmd.Flags().StringVarP(&format, "format", "f", jsonFileType,
		"Define the output format. Supported formats: [json]")
	cmd.Flags().StringSliceVarP(&collections, "collections", "c", []string{}, "List of collections")

	return cmd
}

func isValidExportFormat(format string) bool {
	switch strings.ToLower(format) {
	case jsonFileType:
		return true
	default:
		return false
	}
}
