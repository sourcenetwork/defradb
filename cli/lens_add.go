// Copyright 2024 Democratized Data Foundation
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
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

func MakeLensAddCommand(ctx context.Context) *cobra.Command {
	var lensFile string
	var cmd = &cobra.Command{
		Use:   "add [cfg]",
		Short: "Add a lens to the lens store",
		Long: `Add a lens configuration to the lens store and return its CID.

The lens store is content-addressed, so identical lens configurations
will return the same CID without duplicating storage.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			var lensCfgJson string
			switch {
			case lensFile != "":
				data, err := os.ReadFile(lensFile)
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			case len(args) == 1 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			case len(args) == 1:
				lensCfgJson = args[0]
			default:
				return ErrNoLensConfig
			}

			decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
			decoder.DisallowUnknownFields()

			var lensCfg model.Lens
			if err := decoder.Decode(&lensCfg); err != nil {
				return NewErrInvalidLensConfig(err)
			}

			opt := options.WithIdentity(options.AddLens(), identity.FromContext(cmd.Context()))
			lensID, err := cliClient.AddLens(cmd.Context(), lensCfg, opt)
			if err != nil {
				return err
			}

			return writeJSON(cmd, lensID)
		},
	}

	EmbedCLIExample(ctx, cmd, "add from an argument string",
		`defradb client lens add '{"lenses": [...'`)

	EmbedCLIExample(ctx, cmd, "add from file",
		`defradb client lens add -f lens_config.json`)

	EmbedCLIExample(ctx, cmd, "add from stdin",
		`cat lens_config.json | defradb client lens add -`)

	cmd.Flags().StringVarP(&lensFile, "file", "f", "", "Lens configuration file")
	return cmd
}
