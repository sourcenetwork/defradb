// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mobile

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// PeerInfo returns the node's listen addresses as a JSON array of strings.
func (db *DB) PeerInfo() (string, error) {
	return peerInfo(db.node.DB, db.context(), db.identity)
}

func peerInfo(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.PeerInfo()
	setOptIdentity(opt, identity)
	addrs, err := store.PeerInfo(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(addrs)
}

// ActivePeers returns addresses of currently connected peers as a JSON array.
func (db *DB) ActivePeers() (string, error) {
	return activePeers(db.node.DB, db.context(), db.identity)
}

func activePeers(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.ActivePeers()
	setOptIdentity(opt, identity)
	addrs, err := store.ActivePeers(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(addrs)
}

// Connect connects to the peers at the given addresses.
// addressesJSON is a JSON array of multiaddr strings.
func (db *DB) Connect(addressesJSON string) error {
	return connect(db.node.DB, db.context(), db.identity, addressesJSON)
}

func connect(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	addressesJSON string,
) error {
	var addresses []string
	if err := json.Unmarshal([]byte(addressesJSON), &addresses); err != nil {
		return err
	}
	opt := options.Connect()
	setOptIdentity(opt, identity)
	return store.Connect(ctx, addresses, opt)
}

// AddReplicator adds a replicator for the given peer addresses.
// addressesJSON is a JSON array of multiaddr strings.
// collectionNamesJSON is a JSON array of collection names (may be empty array).
func (db *DB) AddReplicator(addressesJSON string, collectionNamesJSON string) error {
	return addReplicator(db.node.DB, db.context(), db.identity, addressesJSON, collectionNamesJSON)
}

func addReplicator(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	addressesJSON string,
	collectionNamesJSON string,
) error {
	var addresses []string
	if err := json.Unmarshal([]byte(addressesJSON), &addresses); err != nil {
		return err
	}
	opt := options.AddReplicator()
	setOptIdentity(opt, identity)
	if collectionNamesJSON != "" && collectionNamesJSON != "[]" {
		var names []string
		if err := json.Unmarshal([]byte(collectionNamesJSON), &names); err != nil {
			return err
		}
		if len(names) > 0 {
			opt.SetCollectionNames(names)
		}
	}
	return store.AddReplicator(ctx, addresses, opt)
}

// DeleteReplicator removes a replicator by its ID.
// collectionNamesJSON is a JSON array of collection names to stop replicating (may be empty array).
func (db *DB) DeleteReplicator(id string, collectionNamesJSON string) error {
	return deleteReplicator(db.node.DB, db.context(), db.identity, id, collectionNamesJSON)
}

func deleteReplicator(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	id string,
	collectionNamesJSON string,
) error {
	opt := options.DeleteReplicator()
	setOptIdentity(opt, identity)
	if collectionNamesJSON != "" && collectionNamesJSON != "[]" {
		var names []string
		if err := json.Unmarshal([]byte(collectionNamesJSON), &names); err != nil {
			return err
		}
		if len(names) > 0 {
			opt.SetCollectionNames(names)
		}
	}
	return store.DeleteReplicator(ctx, id, opt)
}

// ListReplicators returns all replicators as a JSON array.
func (db *DB) ListReplicators() (string, error) {
	return listReplicators(db.node.DB, db.context(), db.identity)
}

func listReplicators(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.ListReplicators()
	setOptIdentity(opt, identity)
	reps, err := store.ListReplicators(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(reps)
}

// AddP2PCollections subscribes to the given collection topics.
// collectionNamesJSON is a JSON array of collection name strings.
func (db *DB) AddP2PCollections(collectionNamesJSON string) error {
	return addP2PCollections(db.node.DB, db.context(), db.identity, collectionNamesJSON)
}

func addP2PCollections(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionNamesJSON string,
) error {
	var names []string
	if err := json.Unmarshal([]byte(collectionNamesJSON), &names); err != nil {
		return err
	}
	opt := options.AddP2PCollections()
	setOptIdentity(opt, identity)
	return store.AddP2PCollections(ctx, names, opt)
}

// DeleteP2PCollections unsubscribes from the given collection topics.
// collectionNamesJSON is a JSON array of collection name strings.
func (db *DB) DeleteP2PCollections(collectionNamesJSON string) error {
	return deleteP2PCollections(db.node.DB, db.context(), db.identity, collectionNamesJSON)
}

func deleteP2PCollections(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionNamesJSON string,
) error {
	var names []string
	if err := json.Unmarshal([]byte(collectionNamesJSON), &names); err != nil {
		return err
	}
	opt := options.DeleteP2PCollections()
	setOptIdentity(opt, identity)
	return store.DeleteP2PCollections(ctx, names, opt)
}

// ListP2PCollections returns the list of P2P collection names as a JSON array.
func (db *DB) ListP2PCollections() (string, error) {
	return listP2PCollections(db.node.DB, db.context(), db.identity)
}

func listP2PCollections(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.ListP2PCollections()
	setOptIdentity(opt, identity)
	names, err := store.ListP2PCollections(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(names)
}

// AddP2PDocuments subscribes to the given document ID topics.
// docIDsJSON is a JSON array of docID strings.
func (db *DB) AddP2PDocuments(docIDsJSON string) error {
	return addP2PDocuments(db.node.DB, db.context(), db.identity, docIDsJSON)
}

func addP2PDocuments(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	docIDsJSON string,
) error {
	var docIDs []string
	if err := json.Unmarshal([]byte(docIDsJSON), &docIDs); err != nil {
		return err
	}
	opt := options.AddP2PDocuments()
	setOptIdentity(opt, identity)
	return store.AddP2PDocuments(ctx, docIDs, opt)
}

// DeleteP2PDocuments unsubscribes from the given document ID topics.
// docIDsJSON is a JSON array of docID strings.
func (db *DB) DeleteP2PDocuments(docIDsJSON string) error {
	return deleteP2PDocuments(db.node.DB, db.context(), db.identity, docIDsJSON)
}

func deleteP2PDocuments(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	docIDsJSON string,
) error {
	var docIDs []string
	if err := json.Unmarshal([]byte(docIDsJSON), &docIDs); err != nil {
		return err
	}
	opt := options.DeleteP2PDocuments()
	setOptIdentity(opt, identity)
	return store.DeleteP2PDocuments(ctx, docIDs, opt)
}

// ListP2PDocuments returns the list of subscribed document IDs as a JSON array.
func (db *DB) ListP2PDocuments() (string, error) {
	return listP2PDocuments(db.node.DB, db.context(), db.identity)
}

func listP2PDocuments(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.ListP2PDocuments()
	setOptIdentity(opt, identity)
	ids, err := store.ListP2PDocuments(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(ids)
}

// SyncDocuments fetches the latest versions of the given documents from the network.
// docIDsJSON is a JSON array of docID strings.
// timeoutMs > 0 sets a timeout in milliseconds.
func (db *DB) SyncDocuments(collectionName string, docIDsJSON string, timeoutMs int) error {
	return syncDocuments(db.node.DB, db.context(), db.identity, collectionName, docIDsJSON, timeoutMs)
}

func syncDocuments(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionName string,
	docIDsJSON string,
	timeoutMs int,
) error {
	var docIDs []string
	if err := json.Unmarshal([]byte(docIDsJSON), &docIDs); err != nil {
		return err
	}
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	opt := options.SyncDocuments()
	setOptIdentity(opt, identity)
	return store.SyncDocuments(ctx, collectionName, docIDs, opt)
}

// SyncCollectionVersions synchronizes the given collection versions from the network.
// versionIDsJSON is a JSON array of version ID strings.
// timeoutMs > 0 sets a timeout in milliseconds.
func (db *DB) SyncCollectionVersions(versionIDsJSON string, timeoutMs int) error {
	return syncCollectionVersions(db.node.DB, db.context(), db.identity, versionIDsJSON, timeoutMs)
}

func syncCollectionVersions(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	versionIDsJSON string,
	timeoutMs int,
) error {
	var versionIDs []string
	if err := json.Unmarshal([]byte(versionIDsJSON), &versionIDs); err != nil {
		return err
	}
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	opt := options.SyncCollectionVersions()
	setOptIdentity(opt, identity)
	return store.SyncCollectionVersions(ctx, versionIDs, opt)
}

// SyncBranchableCollection syncs a branchable collection's DAG from the network.
// timeoutMs > 0 sets a timeout in milliseconds.
func (db *DB) SyncBranchableCollection(collectionID string, timeoutMs int) error {
	return syncBranchableCollection(db.node.DB, db.context(), db.identity, collectionID, timeoutMs)
}

func syncBranchableCollection(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionID string,
	timeoutMs int,
) error {
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	opt := options.SyncBranchableCollection()
	setOptIdentity(opt, identity)
	return store.SyncBranchableCollection(ctx, collectionID, opt)
}
