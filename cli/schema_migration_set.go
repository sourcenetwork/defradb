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
	"os"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

func MakeSchemaMigrationSetCommand() *cobra.Command {
	var lensFile string
	var cmd = &cobra.Command{
		Use:   "set [src] [dst] [cfg]",
		Short: "Set a schema migration within DefraDB",
		Long: `Set a migration from a source schema version to a destination schema version for
all collections that are on the given source schema version within the local DefraDB node.

Example: set from an argument string:
  defradb client schema migration set bae123 bae456 '{"lenses": [...'

Example: set from file:
  defradb client schema migration set bae123 bae456 -f schema_migration.lens

Example: add from stdin:
  cat schema_migration.lens | defradb client schema migration set bae123 bae456 -

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			db := mustGetContextDB(cmd)

			var lensCfgJson string
			switch {
			case lensFile != "":
				data, err := os.ReadFile(lensFile)
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			case len(args) == 3 && args[2] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			case len(args) == 3:
				lensCfgJson = args[2]
			default:
				return ErrNoLensConfig
			}

			srcSchemaVersionID := args[0]
			dstSchemaVersionID := args[1]

			decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
			decoder.DisallowUnknownFields()

			var lensCfg model.Lens
			if err := decoder.Decode(&lensCfg); err != nil {
				return NewErrInvalidLensConfig(err)
			}

			migrationCfg := client.LensConfig{
				SourceSchemaVersionID:      srcSchemaVersionID,
				DestinationSchemaVersionID: dstSchemaVersionID,
				Lens:                       lensCfg,
			}

			return db.SetMigration(cmd.Context(), migrationCfg)
		},
	}
	cmd.Flags().StringVarP(&lensFile, "file", "f", "", "Lens configuration file")
	return cmd
}
