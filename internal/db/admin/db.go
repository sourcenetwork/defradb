// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package admin

import (
	"context"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/event"
)

func (adminDB *AdminDB) Close() {
	adminDB.db.Close()

	if adminDB.adminACP.HasValue() {
		adminACPValue := adminDB.adminACP.Value()
		if err := adminACPValue.Close(); err != nil {
			log.ErrorE("Failure closing admin acp", err)
		}
	}
}

func (adminDB *AdminDB) PurgeACPState(ctx context.Context) error {
	err := adminDB.db.PurgeACPState(ctx)
	if err != nil {
		return err
	}

	// Purge admin acp state and keep it closed.
	if adminDB.adminACP.HasValue() {
		adminACP := adminDB.adminACP.Value()
		err = adminACP.ResetState(ctx)
		if err != nil {
			return err
		}

		// follow up close call on admin ACP is required since the node.Start function starts
		// admin ACP again anyways so we need to gracefully close before starting again.
		err = adminACP.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (adminDB *AdminDB) NewTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	// checkAdminAccess(ctx, adminDB.adminACP.Value(), acpTypes.AdminDACPerm)
	return adminDB.db.NewTxn(ctx, readonly)
}

func (adminDB *AdminDB) NewConcurrentTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	return adminDB.db.NewConcurrentTxn(ctx, readonly)
}

func (adminDB *AdminDB) Rootstore() datastore.Rootstore {
	return adminDB.db.Rootstore()
}

func (adminDB *AdminDB) Blockstore() datastore.Blockstore {
	return adminDB.db.Blockstore()
}

func (adminDB *AdminDB) Encstore() datastore.Blockstore {
	return adminDB.db.Encstore()
}

func (adminDB *AdminDB) Peerstore() datastore.DSReaderWriter {
	return adminDB.db.Peerstore()
}

func (adminDB *AdminDB) Headstore() corekv.Reader {
	return adminDB.db.Headstore()
}

func (adminDB *AdminDB) Events() *event.Bus {
	return adminDB.db.Events()
}

func (adminDB *AdminDB) MaxTxnRetries() int {
	return adminDB.db.MaxTxnRetries()
}

func (adminDB *AdminDB) PrintDump(ctx context.Context) error {
	return adminDB.db.PrintDump(ctx)
}

func (adminDB *AdminDB) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	return adminDB.db.VerifySignature(ctx, blockCid, pubKey)
}

func (adminDB *AdminDB) GetNodeIdentity(_ context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	return adminDB.db.GetNodeIdentity(nil)
}

func (adminDB *AdminDB) GetNodeIdentityToken(_ context.Context, audience immutable.Option[string]) ([]byte, error) {
	return adminDB.db.GetNodeIdentityToken(nil, audience)
}

func (adminDB *AdminDB) AddPolicy(ctx context.Context, policy string) (client.AddPolicyResult, error) {
	return adminDB.db.AddPolicy(ctx, policy)
}

func (adminDB *AdminDB) AddDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddDocActorRelationshipResult, error) {
	return adminDB.db.AddDocActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (adminDB *AdminDB) DeleteDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteDocActorRelationshipResult, error) {
	return adminDB.db.DeleteDocActorRelationship(ctx, collectionName, docID, relation, targetActor)
}
