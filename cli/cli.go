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
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("cli")

// NewDefraCommand returns the root command instanciated with its tree of subcommands.
func NewDefraCommand(cfg *config.Config) *cobra.Command {
	p2p_collection := MakeP2PCollectionCommand()
	p2p_collection.AddCommand(
		MakeP2PCollectionAddCommand(cfg),
		MakeP2PCollectionRemoveCommand(cfg),
		MakeP2PCollectionGetallCommand(cfg),
	)

	p2p_replicator := MakeP2PReplicatorCommand()
	p2p_replicator.AddCommand(
		MakeP2PReplicatorGetallCommand(cfg),
		MakeP2PReplicatorSetCommand(cfg),
		MakeP2PReplicatorDeleteCommand(cfg),
	)

	p2p := MakeP2PCommand()
	p2p.AddCommand(
		p2p_replicator,
		p2p_collection,
	)

	schema_migrate := MakeSchemaMigrationCommand()
	schema_migrate.AddCommand(
		MakeSchemaMigrationSetCommand(cfg),
		MakeSchemaMigrationGetCommand(cfg),
	)

	schema := MakeSchemaCommand()
	schema.AddCommand(
		MakeSchemaAddCommand(cfg),
		MakeSchemaPatchCommand(cfg),
		schema_migrate,
	)

	index := MakeIndexCommand()
	index.AddCommand(
		MakeIndexCreateCommand(cfg),
		MakeIndexDropCommand(cfg),
		MakeIndexListCommand(cfg),
	)

	backup := MakeBackupCommand()
	backup.AddCommand(
		MakeBackupExportCommand(cfg),
		MakeBackupImportCommand(cfg),
	)

	tx := MakeTxCommand()
	tx.AddCommand(
		MakeTxCreateCommand(cfg),
		MakeTxCommitCommand(cfg),
		MakeTxDiscardCommand(cfg),
	)

	client := MakeClientCommand(cfg)
	client.AddCommand(
		MakeDumpCommand(cfg),
		MakeRequestCommand(cfg),
		schema,
		index,
		p2p,
		backup,
		tx,
	)

	root := MakeRootCommand(cfg)
	root.AddCommand(
		client,
		MakeStartCommand(cfg),
		MakeServerDumpCmd(cfg),
		MakeVersionCommand(),
		MakeInitCommand(cfg),
	)

	return root
}
