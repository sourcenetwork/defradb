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

	"github.com/sourcenetwork/corelog"
)

var log = corelog.NewLogger("cli")

// NewDefraCommand returns the root command instanciated with its tree of subcommands.
func NewDefraCommand() *cobra.Command {
	p2p_collection := MakeP2PCollectionCommand()
	p2p_collection.AddCommand(
		MakeP2PCollectionAddCommand(),
		MakeP2PCollectionRemoveCommand(),
		MakeP2PCollectionGetAllCommand(),
	)

	p2p_replicator := MakeP2PReplicatorCommand()
	p2p_replicator.AddCommand(
		MakeP2PReplicatorGetAllCommand(),
		MakeP2PReplicatorSetCommand(),
		MakeP2PReplicatorDeleteCommand(),
	)

	p2p := MakeP2PCommand()
	p2p.AddCommand(
		p2p_replicator,
		p2p_collection,
		MakeP2PInfoCommand(),
	)

	lens := MakeLensCommand()
	lens.AddCommand(
		MakeLensUpCommand(),
		MakeLensDownCommand(),
		MakeLensReloadCommand(),
		MakeLensSetCommand(),
		MakeLensSetRegistryCommand(),
	)

	schema := MakeSchemaCommand()
	schema.AddCommand(
		MakeSchemaAddCommand(),
		MakeSchemaPatchCommand(),
		MakeSchemaSetActiveCommand(),
		MakeSchemaDescribeCommand(),
	)

	acp_policy := MakeACPPolicyCommand()
	acp_policy.AddCommand(
		MakeACPPolicyAddCommand(),
	)

	acp_relationship := MakeACPRelationshipCommand()
	acp_relationship.AddCommand(
		MakeACPRelationshipAddCommand(),
		MakeACPRelationshipDeleteCommand(),
	)

	acp := MakeACPCommand()
	acp.AddCommand(
		acp_policy,
		acp_relationship,
	)

	view := MakeViewCommand()
	view.AddCommand(
		MakeViewAddCommand(),
		MakeViewRefreshCommand(),
	)

	index := MakeIndexCommand()
	index.AddCommand(
		MakeIndexCreateCommand(),
		MakeIndexDropCommand(),
		MakeIndexListCommand(),
	)

	backup := MakeBackupCommand()
	backup.AddCommand(
		MakeBackupExportCommand(),
		MakeBackupImportCommand(),
	)

	tx := MakeTxCommand()
	tx.AddCommand(
		MakeTxCreateCommand(),
		MakeTxCommitCommand(),
		MakeTxDiscardCommand(),
	)

	collection := MakeCollectionCommand()
	collection.AddCommand(
		MakeCollectionGetCommand(),
		MakeCollectionListDocIDsCommand(),
		MakeCollectionDeleteCommand(),
		MakeCollectionUpdateCommand(),
		MakeCollectionCreateCommand(),
		MakeCollectionDescribeCommand(),
		MakeCollectionPatchCommand(),
	)

	block := MakeBlockCommand()
	block.AddCommand(
		MakeBlockVerifySignatureCommand(),
	)

	client := MakeClientCommand()
	client.AddCommand(
		MakePurgeCommand(),
		MakeDumpCommand(),
		MakeRequestCommand(),
		MakeNodeIdentityCommand(),
		schema,
		acp,
		view,
		index,
		p2p,
		backup,
		tx,
		collection,
		lens,
		block,
	)

	keyring := MakeKeyringCommand()
	keyring.AddCommand(
		MakeKeyringGenerateCommand(),
		MakeKeyringImportCommand(),
		MakeKeyringExportCommand(),
		MakeKeyringListCommand(),
	)

	identity := MakeIdentityCommand()
	identity.AddCommand(
		MakeIdentityNewCommand(),
	)

	root := MakeRootCommand()
	root.AddCommand(
		client,
		keyring,
		identity,
		MakeStartCommand(),
		MakeServerDumpCmd(),
		MakeVersionCommand(),
	)

	return root
}
