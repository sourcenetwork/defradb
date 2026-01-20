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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
)

var log = corelog.NewLogger("cli")

type CLI interface {
	client.TxnStore
	client.P2P
	Purge(ctx context.Context) error
}

// NewDefraCommand returns the root command instanciated with its tree of subcommands.
func NewDefraCommand(ctx context.Context) *cobra.Command {
	p2p_collection := MakeP2PCollectionCommand(ctx)
	p2p_collection.AddCommand(
		MakeP2PCollectionAddCommand(ctx),
		MakeP2PCollectionRemoveCommand(ctx),
		MakeP2PCollectionGetAllCommand(ctx),
		MakeP2PCollectionSyncVersionsCommand(ctx),
		MakeP2PCollectionSyncBranchableCommand(ctx),
	)

	p2p_document := MakeP2PDocumentCommand(ctx)
	p2p_document.AddCommand(
		MakeP2PDocumentAddCommand(ctx),
		MakeP2PDocumentRemoveCommand(ctx),
		MakeP2PDocumentGetAllCommand(ctx),
		MakeP2PDocumentSyncCommand(ctx),
	)

	p2p_replicator := MakeP2PReplicatorCommand(ctx)
	p2p_replicator.AddCommand(
		MakeP2PReplicatorGetAllCommand(ctx),
		MakeP2PReplicatorSetCommand(ctx),
		MakeP2PReplicatorDeleteCommand(ctx),
	)

	p2p := MakeP2PCommand(ctx)
	p2p.AddCommand(
		p2p_replicator,
		p2p_collection,
		p2p_document,
		MakeP2PInfoCommand(ctx),
		MakeP2PActivePeersCommand(ctx),
		MakeP2PConnectCommand(ctx),
	)

	lens := MakeLensCommand(ctx)
	lens.AddCommand(
		MakeLensSetCommand(ctx),
		MakeLensAddCommand(ctx),
		MakeLensListCommand(ctx),
	)

	schema := MakeSchemaCommand(ctx)
	schema.AddCommand(
		MakeSchemaAddCommand(ctx),
	)

	acp_node_relationship := MakeNodeACPRelationshipCommand(ctx)
	acp_node_relationship.AddCommand(
		MakeNodeACPRelationshipAddCommand(ctx),
		MakeNodeACPRelationshipDeleteCommand(ctx),
	)

	nac := MakeNodeACPCommand(ctx)
	nac.AddCommand(
		acp_node_relationship,
		MakeNodeACPReEnableCommand(ctx),
		MakeNodeACPDisableCommand(ctx),
		MakeNodeACPStatusCommand(ctx),
	)

	acp_document_policy := MakeDocumentACPPolicyCommand(ctx)
	acp_document_policy.AddCommand(
		MakeDocumentACPPolicyAddCommand(ctx),
	)

	acp_document_relationship := MakeDocumentACPRelationshipCommand(ctx)
	acp_document_relationship.AddCommand(
		MakeDocumentACPRelationshipAddCommand(ctx),
		MakeDocumentACPRelationshipDeleteCommand(ctx),
	)

	dac := MakeDocumentACPCommand(ctx)
	dac.AddCommand(
		acp_document_policy,
		acp_document_relationship,
	)

	acp := MakeACPCommand(ctx)
	acp.AddCommand(
		nac,
		dac,
	)

	view := MakeViewCommand(ctx)
	view.AddCommand(
		MakeViewAddCommand(ctx),
		MakeViewRefreshCommand(ctx),
	)

	index := MakeIndexCommand(ctx)
	index.AddCommand(
		MakeIndexCreateCommand(ctx),
		MakeIndexDropCommand(ctx),
		MakeIndexListCommand(ctx),
	)

	encrypted_index := MakeEncryptedIndexCommand(ctx)
	encrypted_index.AddCommand(
		MakeEncryptedIndexCreateCommand(ctx),
		MakeEncryptedIndexDeleteCommand(ctx),
		MakeEncryptedIndexListCommand(ctx),
	)

	backup := MakeBackupCommand(ctx)
	backup.AddCommand(
		MakeBackupExportCommand(ctx),
		MakeBackupImportCommand(ctx),
	)

	tx := MakeTxCommand(ctx)
	tx.AddCommand(
		MakeTxCreateCommand(ctx),
		MakeTxCommitCommand(ctx),
		MakeTxDiscardCommand(ctx),
	)

	collection := MakeCollectionCommand(ctx)
	collection.AddCommand(
		MakeCollectionGetCommand(ctx),
		MakeCollectionListDocIDsCommand(ctx),
		MakeCollectionDeleteCommand(ctx),
		MakeCollectionUpdateCommand(ctx),
		MakeCollectionCreateCommand(ctx),
		MakeCollectionDescribeCommand(ctx),
		MakeCollectionPatchCommand(ctx),
		MakeCollectionSetActiveCommand(ctx),
		MakeCollectionTruncateCommand(ctx),
	)

	block := MakeBlockCommand(ctx)
	block.AddCommand(
		MakeBlockVerifySignatureCommand(ctx),
	)

	client := MakeClientCommand(ctx)
	client.AddCommand(
		MakePurgeCommand(ctx),
		MakeDumpCommand(ctx),
		MakeRequestCommand(ctx),
		MakeNodeIdentityCommand(ctx),
		schema,
		acp,
		view,
		index,
		encrypted_index,
		p2p,
		backup,
		tx,
		collection,
		lens,
		block,
	)

	keyring := MakeKeyringCommand(ctx)
	keyring.AddCommand(
		MakeKeyringGenerateCommand(ctx),
		MakeKeyringImportCommand(ctx),
		MakeKeyringExportCommand(ctx),
		MakeKeyringListCommand(ctx),
	)

	identity := MakeIdentityCommand(ctx)
	identity.AddCommand(
		MakeIdentityNewCommand(ctx),
	)

	root := MakeRootCommand(ctx)
	root.AddCommand(
		client,
		keyring,
		identity,
		MakeStartCommand(ctx),
		MakeServerDumpCmd(),
		MakeVersionCommand(ctx),
		MakeWizardCommand(),
	)

	return root
}
