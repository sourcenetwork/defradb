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
	"strconv"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/spf13/cobra"
)

func MakeSchemaMigrationSetRegistryCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "set-registry [collectionID] [cfg]",
		Short: "Set a schema migration within the DefraDB LensRegistry",
		Long: `Set a migration to a collection within the LensRegistry of the local DefraDB node.
Does not persist the migration after restart.

Example: set from an argument string:
  defradb client schema migration set-registry 2 '{"lenses": [...'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)

			decoder := json.NewDecoder(strings.NewReader(args[1]))
			decoder.DisallowUnknownFields()

			var lensCfg model.Lens
			if err := decoder.Decode(&lensCfg); err != nil {
				return NewErrInvalidLensConfig(err)
			}

			collectionID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			return store.LensRegistry().SetMigration(cmd.Context(), uint32(collectionID), lensCfg)
		},
	}
	return cmd
}
