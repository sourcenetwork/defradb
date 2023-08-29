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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("cli")

// NewDefraCommand returns the root command instanciated with its tree of subcommands.
func NewDefraCommand(cfg *config.Config) (*cobra.Command, error) {
	db, err := http.NewClient("http://" + cfg.API.Address)
	if err != nil {
		return nil, err
	}

	rootCmd := MakeRootCommand(cfg)
	p2pCmd := MakeP2PCommand(cfg)
	schemaCmd := MakeSchemaCommand()
	schemaMigrationCmd := MakeSchemaMigrationCommand()
	indexCmd := MakeIndexCommand()
	clientCmd := MakeClientCommand()
	backupCmd := MakeBackupCommand()
	p2pReplicatorCmd := MakeP2PReplicatorCommand()
	p2pCollectionCmd := MakeP2PCollectionCommand()
	p2pCollectionCmd.AddCommand(
		MakeP2PCollectionAddCommand(cfg, db),
		MakeP2PCollectionRemoveCommand(cfg, db),
		MakeP2PCollectionGetallCommand(cfg, db),
	)
	p2pReplicatorCmd.AddCommand(
		MakeP2PReplicatorGetallCommand(cfg, db),
		MakeP2PReplicatorSetCommand(cfg, db),
		MakeP2PReplicatorDeleteCommand(cfg, db),
	)
	p2pCmd.AddCommand(
		p2pReplicatorCmd,
		p2pCollectionCmd,
	)
	schemaMigrationCmd.AddCommand(
		MakeSchemaMigrationSetCommand(cfg, db),
		MakeSchemaMigrationGetCommand(cfg, db),
	)
	schemaCmd.AddCommand(
		MakeSchemaAddCommand(cfg, db),
		MakeSchemaPatchCommand(cfg, db),
		schemaMigrationCmd,
	)
	indexCmd.AddCommand(
		MakeIndexCreateCommand(cfg, db),
		MakeIndexDropCommand(cfg, db),
		MakeIndexListCommand(cfg, db),
	)
	backupCmd.AddCommand(
		MakeBackupExportCommand(cfg, db),
		MakeBackupImportCommand(cfg, db),
	)
	clientCmd.AddCommand(
		MakeDumpCommand(cfg, db),
		MakeRequestCommand(cfg, db),
		schemaCmd,
		indexCmd,
		p2pCmd,
		backupCmd,
	)
	rootCmd.AddCommand(
		clientCmd,
		MakeStartCommand(cfg),
		MakeServerDumpCmd(cfg, db),
		MakeVersionCommand(),
		MakeInitCommand(cfg),
	)

	return rootCmd, nil
}
