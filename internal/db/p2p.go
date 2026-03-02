// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.P2P = (*DB)(nil)

func (db *DB) sendUpdate(evt event.Update) {
	db.events.Publish(event.NewMessage(event.UpdateName, evt))
	if db.p2p == nil {
		return
	}
	_ = db.p2p.SendUpdate(evt)
}

// PeerInfo returns the p2p host id and listening addresses.
func (db *DB) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeGetP2PPeerInfoPerm); err != nil {
		return nil, err
	}

	if db.p2p == nil {
		return nil, nil
	}
	return db.p2p.PeerInfo()
}

// Connect tries to connect to the peer with the given [PeerInfo].
func (db *DB) Connect(
	ctx context.Context, addresses []string, opts ...options.Enumerable[options.ConnectOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeConnectP2PPeerPerm); err != nil {
		return err
	}

	return db.p2p.Connect(ctx, addresses)
}

// AddReplicator adds a replicator to the persisted list or adds
// collections if the replicator already exists.
func (db *DB) AddReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeAddP2PReplicatorPerm); err != nil {
		return err
	}

	ctx = identity.WithContext(ctx, opt.Identity)

	if db.p2p == nil {
		return ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.p2p.AddReplicator(ctx, addresses, opt.CollectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// DeleteReplicator deletes a replicator from the persisted list
// or specific collections if they are specified.
func (db *DB) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDeleteP2PReplicatorPerm); err != nil {
		return err
	}

	ctx = identity.WithContext(ctx, opt.Identity)

	if db.p2p == nil {
		return ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.p2p.DeleteReplicator(ctx, id, opt.CollectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// ListReplicators returns the full list of replicators with their
// subscribed collections.
func (db *DB) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeListP2PReplicatorPerm); err != nil {
		return nil, err
	}

	if db.p2p == nil {
		return nil, ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()
	return db.p2p.ListReplicators(ctx)
}

func (db *DB) ActivePeers(
	ctx context.Context, opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeGetP2PActivePeersPerm); err != nil {
		return nil, err
	}

	if db.p2p == nil {
		return nil, ErrNoP2P
	}

	return db.p2p.ActivePeers(ctx)
}

// AddP2PCollections adds the given collections to the P2P system and
// subscribes to their topics. It will error if any of the provided
// collection names are invalid.
func (db *DB) AddP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeAddP2PCollectionPerm); err != nil {
		return err
	}

	if db.p2p == nil {
		return ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.p2p.AddP2PCollections(identity.WithContext(ctx, opt.Identity), collectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// DeleteP2PCollections deletes the given collections from the P2P system and
// unsubscribes from their topics. It will error if the provided
// collection names are invalid.
func (db *DB) DeleteP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDeleteP2PCollectionPerm); err != nil {
		return err
	}

	if db.p2p == nil {
		return ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.p2p.DeleteP2PCollections(identity.WithContext(ctx, opt.Identity), collectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// ListP2PCollections returns the list of persisted collection names that
// the P2P system subscribes to.
func (db *DB) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeListP2PCollectionPerm); err != nil {
		return nil, err
	}

	if db.p2p == nil {
		return nil, ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.p2p.ListP2PCollections(identity.WithContext(ctx, opt.Identity))
}

// AddP2PDocuments adds the given docIDs to the P2P system and
// subscribes to their topics. It will error if any of the provided
// docIDs are invalid.
func (db *DB) AddP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeAddP2PDocumentPerm); err != nil {
		return err
	}

	if db.p2p == nil {
		return ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.p2p.AddP2PDocuments(ctx, docIDs...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// DeleteP2PDocuments removes the given docIDs from the P2P system and
// unsubscribes from their topics. It will error if the provided
// docIDs are invalid.
func (db *DB) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDeleteP2PDocumentPerm); err != nil {
		return err
	}

	if db.p2p == nil {
		return ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.p2p.DeleteP2PDocuments(ctx, docIDs...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// ListP2PDocuments returns the list of persisted docIDs that
// the P2P system subscribes to.
func (db *DB) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeListP2PDocumentPerm); err != nil {
		return nil, err
	}

	if db.p2p == nil {
		return nil, ErrNoP2P
	}
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.p2p.ListP2PDocuments(ctx)
}

// SyncDocuments requests the latest versions of specified documents from the network
// and synchronizes their DAGs locally. It doesn't automatically subscribe
// to the documents or their collection for future updates.
// context.WithTimeout can be used to set a timeout for the operation.
//
// WARNING: This function does not respect transactions.
func (db *DB) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
	opts ...options.Enumerable[options.SyncDocumentsOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeSyncP2PDocumentsPerm); err != nil {
		return err
	}

	ctx = identity.WithContext(ctx, opt.Identity)

	if db.p2p == nil {
		return ErrNoP2P
	}
	return db.p2p.SyncDocuments(ctx, collectionName, docIDs)
}

// SyncCollectionVersions synchronizes the given collection versions to the local node.
//
// It will not complete until a version is found, so it is strongly recommended
// to set a timeout using `context.WithTimeout`.
func (db *DB) SyncCollectionVersions(
	ctx context.Context,
	versionIDs []string,
	opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeSyncP2PCollectionVersionsPerm); err != nil {
		return err
	}

	if db.p2p == nil {
		return ErrNoP2P
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.p2p.SyncCollectionVersions(ctx, versionIDs...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// SyncBranchableCollection requests the latest version of the branchable collection's DAG
// from the network and synchronizes it locally. This syncs the collection-level history
// for branchable collections (collections marked with @branchable directive).
// It doesn't automatically subscribe to the collection for future updates.
// context.WithTimeout can be used to set a timeout for the operation.
//
// WARNING: This function does not respect transactions.
func (db *DB) SyncBranchableCollection(
	ctx context.Context,
	collectionID string,
	opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeSyncP2PBranchableCollectionPerm); err != nil {
		return err
	}

	if db.p2p == nil {
		return ErrNoP2P
	}
	return db.p2p.SyncBranchableCollection(ctx, collectionID, opt)
}
