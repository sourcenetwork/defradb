// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package cli provides the command-line interface.
*/
package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("cli")

// NewDefraCommand returns the root command instanciated with its tree of subcommands.
func NewDefraCommand(cfg *config.Config) *cobra.Command {
	rootCmd := MakeRootCommand(cfg)
	p2pCmd := MakeP2PCommand()
	schemaCmd := MakeSchemaCommand()
	schemaMigrationCmd := MakeSchemaMigrationCommand()
	indexCmd := MakeIndexCommand()
	clientCmd := MakeClientCommand()
	backupCmd := MakeBackupCommand()
	p2pReplicatorCmd := MakeP2PReplicatorCommand()
	p2pCollectionCmd := MakeP2PCollectionCommand()
	p2pCollectionCmd.AddCommand(
		MakeP2PCollectionAddCommand(cfg),
		MakeP2PCollectionRemoveCommand(cfg),
		MakeP2PCollectionGetallCommand(cfg),
	)
	p2pReplicatorCmd.AddCommand(
		MakeP2PReplicatorGetallCommand(cfg),
		MakeP2PReplicatorSetCommand(cfg),
		MakeP2PReplicatorDeleteCommand(cfg),
	)
	p2pCmd.AddCommand(
		p2pReplicatorCmd,
		p2pCollectionCmd,
	)
	schemaMigrationCmd.AddCommand(
		MakeSchemaMigrationSetCommand(cfg),
		MakeSchemaMigrationGetCommand(cfg),
	)
	schemaCmd.AddCommand(
		MakeSchemaAddCommand(cfg),
		MakeSchemaPatchCommand(cfg),
		schemaMigrationCmd,
	)
	indexCmd.AddCommand(
		MakeIndexCreateCommand(cfg),
		MakeIndexDropCommand(cfg),
		MakeIndexListCommand(cfg),
	)
	backupCmd.AddCommand(
		MakeBackupExportCommand(cfg),
		MakeBackupImportCommand(cfg),
	)
	clientCmd.AddCommand(
		MakeDumpCommand(cfg),
		MakeRequestCommand(cfg),
		schemaCmd,
		indexCmd,
		p2pCmd,
		backupCmd,
	)
	rootCmd.AddCommand(
		clientCmd,
		MakeStartCommand(cfg),
		MakeServerDumpCmd(cfg),
		MakeVersionCommand(),
		MakeInitCommand(cfg),
	)

	return rootCmd
}

func writeJSON(cmd *cobra.Command, out any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
