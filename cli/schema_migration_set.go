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
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

func MakeSchemaMigrationSetCommand(cfg *config.Config) *cobra.Command {
	var lensFile string
	var cmd = &cobra.Command{
		Use:   "set [src] [dst] [cfg]",
		Short: "Set a schema migration within DefraDB",
		Long: `Set a migration between two schema versions within the local DefraDB node.

Example: set from an argument string:
  defradb client schema migration set bae123 bae456 '{"lenses": [...'

Example: set from file:
  defradb client schema migration set bae123 bae456 -f schema_migration.lens

Example: add from stdin:
  cat schema_migration.lens | defradb client schema migration set bae123 bae456 -

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(2)(cmd, args); err != nil {
				return errors.New("must specify src and dst schema versions, as well as a lens cfg")
			}
			if err := cobra.MaximumNArgs(3)(cmd, args); err != nil {
				return errors.New("must specify src and dst schema versions, as well as a lens cfg")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var lensCfgJson string
			var srcSchemaVersionID string
			var dstSchemaVersionID string
			fi, err := os.Stdin.Stat()
			if err != nil {
				return err
			}

			if lensFile != "" {
				buf, err := os.ReadFile(lensFile)
				if err != nil {
					return errors.Wrap("failed to read schema file", err)
				}
				lensCfgJson = string(buf)
			} else if len(args) == 2 {
				// If the lensFile flag has not been provided then it must be provided as an arg
				// and thus len(args) cannot be 2
				return errors.Wrap("must provide a lens cfg", err)
			} else if isFileInfoPipe(fi) && args[2] != "-" {
				log.FeedbackInfo(
					cmd.Context(),
					"Run 'defradb client schema migration set -' to read from stdin."+
						" Example: 'cat schema_migration.lens | defradb client schema migration set -').",
				)
				return nil
			} else if args[2] == "-" {
				stdin, err := readStdin()
				if err != nil {
					return errors.Wrap("failed to read stdin", err)
				}
				if len(stdin) == 0 {
					return errors.New("no lens cfg in stdin provided")
				} else {
					lensCfgJson = stdin
				}
			} else {
				lensCfgJson = args[2]
			}

			srcSchemaVersionID = args[0]
			dstSchemaVersionID = args[1]

			if lensCfgJson == "" {
				return errors.New("empty lens configuration provided")
			}
			if srcSchemaVersionID == "" {
				return errors.New("no source schema version id provided")
			}
			if dstSchemaVersionID == "" {
				return errors.New("no destination schema version id provided")
			}

			var lensCfg model.Lens
			err = json.Unmarshal([]byte(lensCfgJson), &lensCfg)
			if err != nil {
				return errors.Wrap("invalid lens configuration", err)
			}

			migrationCfg := client.LensConfig{
				SourceSchemaVersionID:      srcSchemaVersionID,
				DestinationSchemaVersionID: dstSchemaVersionID,
				Lens:                       lensCfg,
			}

			migrationCfgJson, err := json.Marshal(migrationCfg)
			if err != nil {
				return errors.Wrap("failed to marshal cfg", err)
			}

			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.SchemaMigrationPath)
			if err != nil {
				return errors.Wrap("join paths failed", err)
			}

			res, err := http.Post(endpoint.String(), "application/json", strings.NewReader(string(migrationCfgJson)))
			if err != nil {
				return errors.Wrap("failed to post schema migration", err)
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

			stdout, err := os.Stdout.Stat()
			if err != nil {
				return errors.Wrap("failed to stat stdout", err)
			}
			if isFileInfoPipe(stdout) {
				cmd.Println(string(response))
			} else {
				type migrationSetResponse struct {
					Errors []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				r := migrationSetResponse{}
				err = json.Unmarshal(response, &r)
				if err != nil {
					return NewErrFailedToUnmarshalResponse(err)
				}
				if len(r.Errors) > 0 {
					log.FeedbackError(cmd.Context(), "Failed to set schema migration",
						logging.NewKV("Errors", r.Errors))
				} else {
					log.FeedbackInfo(cmd.Context(), "Successfully set schema migration")
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&lensFile, "file", "f", "", "Lens configuration file")
	return cmd
}
