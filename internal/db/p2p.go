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
	"github.com/sourcenetwork/defradb/event"
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
func (db *DB) PeerInfo() ([]string, error) {
	if db.p2p == nil {
		return nil, nil
	}
	return db.p2p.PeerInfo()
}

// Connect tries to connect to the peer with the given [PeerInfo].
func (db *DB) Connect(ctx context.Context, addresses []string) error {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PPeerConnectPerm); err != nil {
		return err
	}

	return db.p2p.Connect(ctx, addresses)
}

// CreateReplicator adds a replicator to the persisted list or adds
// schemas if the replicator already exists.
func (db *DB) CreateReplicator(ctx context.Context, addresses []string, collectionNames ...string) error {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PReplicatorCreatePerm); err != nil {
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

	err = db.p2p.CreateReplicator(ctx, addresses, collectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// DeleteReplicator deletes a replicator from the persisted list
// or specific schemas if they are specified.
func (db *DB) DeleteReplicator(ctx context.Context, id string, collectionNames ...string) error {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PReplicatorDeletePerm); err != nil {
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

	err = db.p2p.DeleteReplicator(ctx, id, collectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// ListReplicators returns the full list of replicators with their
// subscribed schemas.
func (db *DB) ListReplicators(ctx context.Context) ([]client.Replicator, error) {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PReplicatorListPerm); err != nil {
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

func (db *DB) ActivePeers(ctx context.Context) ([]string, error) {
	if db.p2p == nil {
		return nil, ErrNoP2P
	}

	return db.p2p.ActivePeers(ctx)
}

// CreateP2PCollections creates the given collections to the P2P system and
// subscribes to their topics. It will error if any of the provided
// collection names are invalid.
func (db *DB) CreateP2PCollections(ctx context.Context, collectionNames ...string) error {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PCollectionCreatePerm); err != nil {
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

	err = db.p2p.CreateP2PCollections(ctx, collectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// DeleteP2PCollections deletes the given collections from the P2P system and
// unsubscribes from their topics. It will error if the provided
// collection names are invalid.
func (db *DB) DeleteP2PCollections(ctx context.Context, collectionNames ...string) error {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PCollectionDeletePerm); err != nil {
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

	err = db.p2p.DeleteP2PCollections(ctx, collectionNames...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// ListP2PCollections returns the list of persisted collection names that
// the P2P system subscribes to.
func (db *DB) ListP2PCollections(ctx context.Context) ([]string, error) {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PCollectionListPerm); err != nil {
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

	return db.p2p.ListP2PCollections(ctx)
}

// CreateP2PDocuments adds the given docIDs to the P2P system and
// subscribes to their topics. It will error if any of the provided
// docIDs are invalid.
func (db *DB) CreateP2PDocuments(ctx context.Context, docIDs ...string) error {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PDocumentCreatePerm); err != nil {
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

	err = db.p2p.CreateP2PDocuments(ctx, docIDs...)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// DeleteP2PDocuments removes the given docIDs from the P2P system and
// unsubscribes from their topics. It will error if the provided
// docIDs are invalid.
func (db *DB) DeleteP2PDocuments(ctx context.Context, docIDs ...string) error {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PDocumentDeletePerm); err != nil {
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
func (db *DB) ListP2PDocuments(ctx context.Context) ([]string, error) {
	if err := db.checkNodeAccess(ctx, acpTypes.NodeP2PDocumentListPerm); err != nil {
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
func (db *DB) SyncDocuments(ctx context.Context, collectionName string, docIDs []string) error {
	if db.p2p == nil {
		return ErrNoP2P
	}
	return db.p2p.SyncDocuments(ctx, collectionName, docIDs)
}

// SyncCollectionVersions synchronizes the given collection versions to the local node.
//
// It will not complete until a version is found, so it is strongly recommended
// to set a timeout using `context.WithTimeout`.
func (db *DB) SyncCollectionVersions(ctx context.Context, versionIDs ...string) error {
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
func (db *DB) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	if db.p2p == nil {
		return ErrNoP2P
	}
	return db.p2p.SyncBranchableCollection(ctx, collectionID)
}
