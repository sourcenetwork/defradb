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
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

const cborFileType = "cbor"
const jsonFileType = "json"

func MakeDBExportCommand(cfg *config.Config) *cobra.Command {
	var collections []string
	var pretty bool
	var format string
	var cmd = &cobra.Command{
		Use:   "export  [-c --collections | -p --pretty | -f --format] <output_path>",
		Short: "Export the database to a file",
		Long: `Export the database to a file.
		
If the --collection flag is provided, only the data for that collection will be exported.
Otherwise, all collections in the database will be exported.

If the --pretty flag is provided, the JSON will be pretty printed.

If the --format flag is provided, the supported options are json and cbor.

Example: export data for the 'Users' collection:
  defradb client export --format cbor --collection Users user_data.json`,
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

			if len(collections) != 0 {
				for i := range collections {
					collections[i] = strings.Trim(collections[i], " ")
				}
				values := url.Values{
					"collections": collections,
				}
				endpoint.RawQuery = values.Encode()
			}

			res, err := http.Get(endpoint.String())
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
				r := httpAPIResponse[map[string][]map[string]any]{}
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
				var b []byte

				if strings.ToLower(format) == cborFileType {
					em, err := cbor.CanonicalEncOptions().EncMode()
					if err != nil {
						return NewFailedToMarshalData(err)
					}
					b, err = em.Marshal(r.Data)
					if err != nil {
						return NewFailedToMarshalData(err)
					}
				} else if pretty {
					b, err = json.MarshalIndent(r.Data, "", "  ")
					if err != nil {
						return NewFailedToMarshalData(err)
					}
				} else {
					b, err = json.Marshal(r.Data)
					if err != nil {
						return NewFailedToMarshalData(err)
					}
				}

				err = os.WriteFile(outputPath, b, 0644)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&pretty, "pretty", "p", false, "Set the output JSON to be pretty printed")
	cmd.Flags().StringVarP(&format, "format", "f", jsonFileType,
		"Define the output format. Supported formats: [json, cbor]")
	cmd.Flags().StringSliceVarP(&collections, "collections", "c", []string{}, "List of collections")

	return cmd
}

func isValidExportFormat(format string) bool {
	switch strings.ToLower(format) {
	case cborFileType, jsonFileType:
		return true
	default:
		return false
	}
}
