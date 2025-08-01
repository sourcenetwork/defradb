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
	"context"

	"github.com/sourcenetwork/corelog"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

var log = corelog.NewLogger("cli")

type CLI interface {
	client.TxnStore
	client.P2P
	Purge(ctx context.Context) error
}

// NewDefraCommand returns the root command instanciated with its tree of subcommands.
func NewDefraCommand() *cobra.Command {
	p2p_collection := MakeP2PCollectionCommand()
	p2p_collection.AddCommand(
		MakeP2PCollectionAddCommand(),
		MakeP2PCollectionRemoveCommand(),
		MakeP2PCollectionGetAllCommand(),
	)

	p2p_document := MakeP2PDocumentCommand()
	p2p_document.AddCommand(
		MakeP2PDocumentAddCommand(),
		MakeP2PDocumentRemoveCommand(),
		MakeP2PDocumentGetAllCommand(),
		MakeP2PDocumentSyncCommand(),
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
		p2p_document,
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

	acp_node_relationship := MakeNodeACPRelationshipCommand()
	acp_node_relationship.AddCommand(
		MakeNodeACPRelationshipAddCommand(),
		MakeNodeACPRelationshipDeleteCommand(),
	)

	nac := MakeNodeACPCommand()
	nac.AddCommand(
		acp_node_relationship,
		MakeNodeACPReEnableCommand(),
		MakeNodeACPDisableCommand(),
		MakeNodeACPStatusCommand(),
	)

	acp_dac_policy := MakeDocumentACPPolicyCommand()
	acp_dac_policy.AddCommand(
		MakeDocumentACPPolicyAddCommand(),
	)

	acp_dac_relationship := MakeDocumentACPRelationshipCommand()
	acp_dac_relationship.AddCommand(
		MakeDocumentACPRelationshipAddCommand(),
		MakeDocumentACPRelationshipDeleteCommand(),
	)

	dac := MakeDocumentACPCommand()
	dac.AddCommand(
		acp_dac_policy,
		acp_dac_relationship,
	)

	acp := MakeACPCommand()
	acp.AddCommand(
		nac,
		dac,
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
