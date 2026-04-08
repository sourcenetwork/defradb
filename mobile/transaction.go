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

	"github.com/sourcenetwork/defradb/client"
)

// Transaction wraps a client.Txn and associates it with a DB for context building.
type Transaction struct {
	txn client.Txn
	db  *DB
}

// ID returns the transaction's unique identifier.
// Gomobile does not support uint64, so the value is cast to int64.
func (t *Transaction) ID() int64 {
	return int64(t.txn.ID()) //nolint:gosec
}

// Commit finalizes the transaction.
func (t *Transaction) Commit() error {
	return t.txn.Commit()
}

// Discard discards the transaction without committing.
func (t *Transaction) Discard() {
	t.txn.Discard()
}

// ---------- Store methods on Transaction ----------

func (t *Transaction) txnContext() context.Context {
	return t.db.contextWithTxn(t.txn)
}

// AddCollection creates new collection types from the given SDL string.
func (t *Transaction) AddCollection(sdl string) (string, error) {
	return addCollection(t.txn, t.txnContext(), t.db.identity, sdl)
}

// PatchCollection applies a JSON patch string to the collection definitions.
func (t *Transaction) PatchCollection(patch string, migrationJSON string) error {
	return patchCollection(t.txn, t.txnContext(), t.db.identity, patch, migrationJSON)
}

// SetActiveCollectionVersion activates the collection version with the given versionID.
func (t *Transaction) SetActiveCollectionVersion(versionID string) error {
	return setActiveCollectionVersion(t.txn, t.txnContext(), t.db.identity, versionID)
}

// AddView creates a new DefraDB view.
func (t *Transaction) AddView(gqlQuery string, sdl string, transformCID string) (string, error) {
	return addView(t.txn, t.txnContext(), t.db.identity, gqlQuery, sdl, transformCID)
}

// RefreshViews refreshes materialized view caches.
func (t *Transaction) RefreshViews(collectionName string, versionID string, collectionID string, getInactive bool) error {
	return refreshViews(t.txn, t.txnContext(), t.db.identity, collectionName, versionID, collectionID, getInactive)
}

// SetMigration sets the migration for the collection versions.
func (t *Transaction) SetMigration(configJSON string) (string, error) {
	return setMigration(t.txn, t.txnContext(), t.db.identity, configJSON)
}

// AddLens stores a lens configuration.
func (t *Transaction) AddLens(lensJSON string) (string, error) {
	return addLens(t.txn, t.txnContext(), t.db.identity, lensJSON)
}

// ListLenses returns all stored lenses as a JSON object.
func (t *Transaction) ListLenses() (string, error) {
	return listLenses(t.txn, t.txnContext(), t.db.identity)
}

// GetCollectionByName returns a Collection wrapper for the named collection.
func (t *Transaction) GetCollectionByName(name string) (*Collection, error) {
	col, err := t.txn.GetCollectionByName(t.txnContext(), name)
	if err != nil {
		return nil, err
	}
	return &Collection{col: col, db: t.db, txn: t.txn}, nil
}

// GetCollections returns a JSON array of collection descriptions.
func (t *Transaction) GetCollections(collectionName string, versionID string, collectionID string, getInactive bool) (string, error) {
	return getCollections(t.txn, t.txnContext(), t.db.identity, collectionName, versionID, collectionID, getInactive)
}

// ListIndexes returns all indexes as a JSON object.
func (t *Transaction) ListIndexes() (string, error) {
	return listIndexes(t.txn, t.txnContext(), t.db.identity)
}

// ListAllEncryptedIndexes returns all encrypted indexes as a JSON object.
func (t *Transaction) ListAllEncryptedIndexes() (string, error) {
	return listAllEncryptedIndexes(t.txn, t.txnContext(), t.db.identity)
}

// ExecRequest executes a GQL request within this transaction.
func (t *Transaction) ExecRequest(request string, operationName string, variablesJSON string) (string, error) {
	return execRequest(t.txn, t.txnContext(), t.db.identity, request, operationName, variablesJSON)
}

// BasicImport imports a JSON dataset.
func (t *Transaction) BasicImport(filepath string) error {
	return basicImport(t.txn, t.txnContext(), filepath)
}

// BasicExport exports the database to a JSON file.
func (t *Transaction) BasicExport(filepath string) error {
	return basicExport(t.txn, t.txnContext(), filepath)
}

// PrintDump logs the entire contents of the database store.
func (t *Transaction) PrintDump() error {
	return printDump(t.txn, t.txnContext())
}

// AddDACPolicy adds a document ACP policy.
func (t *Transaction) AddDACPolicy(policy string) (string, error) {
	return addDACPolicy(t.txn, t.txnContext(), t.db.identity, policy)
}

// AddDACActorRelationship adds a relationship between a document and an actor.
func (t *Transaction) AddDACActorRelationship(collectionName string, docID string, relation string, targetActor string) (string, error) {
	return addDACActorRelationship(t.txn, t.txnContext(), t.db.identity, collectionName, docID, relation, targetActor)
}

// DeleteDACActorRelationship deletes a relationship between a document and an actor.
func (t *Transaction) DeleteDACActorRelationship(collectionName string, docID string, relation string, targetActor string) (string, error) {
	return deleteDACActorRelationship(t.txn, t.txnContext(), t.db.identity, collectionName, docID, relation, targetActor)
}

// AddNACActorRelationship adds a node access relationship.
func (t *Transaction) AddNACActorRelationship(relation string, targetActor string) (string, error) {
	return addNACActorRelationship(t.txn, t.txnContext(), t.db.identity, relation, targetActor)
}

// DeleteNACActorRelationship deletes a node access relationship.
func (t *Transaction) DeleteNACActorRelationship(relation string, targetActor string) (string, error) {
	return deleteNACActorRelationship(t.txn, t.txnContext(), t.db.identity, relation, targetActor)
}

// ReEnableNAC re-enables node access control.
func (t *Transaction) ReEnableNAC() error {
	return reEnableNAC(t.txn, t.txnContext(), t.db.identity)
}

// DisableNAC temporarily disables node access control.
func (t *Transaction) DisableNAC() error {
	return disableNAC(t.txn, t.txnContext(), t.db.identity)
}

// GetNACStatus returns the node access control status as a JSON string.
func (t *Transaction) GetNACStatus() (string, error) {
	return getNACStatus(t.txn, t.txnContext(), t.db.identity)
}

// GetNodeIdentity returns the node's public identity as a JSON string.
func (t *Transaction) GetNodeIdentity() (string, error) {
	return getNodeIdentity(t.txn, t.txnContext())
}

// VerifySignature verifies the signature of a block using the given public key.
func (t *Transaction) VerifySignature(publicKeyHex string, publicKeyType string, blockCID string) error {
	return verifySignature(t.txn, t.txnContext(), t.db.identity, publicKeyHex, publicKeyType, blockCID)
}

// ---------- P2P methods on Transaction ----------

// PeerInfo returns the node's listen addresses as a JSON array.
func (t *Transaction) PeerInfo() (string, error) {
	return peerInfo(t.txn, t.txnContext(), t.db.identity)
}

// ActivePeers returns addresses of currently connected peers as a JSON array.
func (t *Transaction) ActivePeers() (string, error) {
	return activePeers(t.txn, t.txnContext(), t.db.identity)
}

// Connect connects to the peers at the given addresses.
func (t *Transaction) Connect(addressesJSON string) error {
	return connect(t.txn, t.txnContext(), t.db.identity, addressesJSON)
}

// AddReplicator adds a replicator for the given peer addresses.
func (t *Transaction) AddReplicator(addressesJSON string, collectionNamesJSON string) error {
	return addReplicator(t.txn, t.txnContext(), t.db.identity, addressesJSON, collectionNamesJSON)
}

// DeleteReplicator removes a replicator by its ID.
func (t *Transaction) DeleteReplicator(id string, collectionNamesJSON string) error {
	return deleteReplicator(t.txn, t.txnContext(), t.db.identity, id, collectionNamesJSON)
}

// ListReplicators returns all replicators as a JSON array.
func (t *Transaction) ListReplicators() (string, error) {
	return listReplicators(t.txn, t.txnContext(), t.db.identity)
}

// AddP2PCollections subscribes to the given collection topics.
func (t *Transaction) AddP2PCollections(collectionNamesJSON string) error {
	return addP2PCollections(t.txn, t.txnContext(), t.db.identity, collectionNamesJSON)
}

// DeleteP2PCollections unsubscribes from the given collection topics.
func (t *Transaction) DeleteP2PCollections(collectionNamesJSON string) error {
	return deleteP2PCollections(t.txn, t.txnContext(), t.db.identity, collectionNamesJSON)
}

// ListP2PCollections returns the list of P2P collection names as a JSON array.
func (t *Transaction) ListP2PCollections() (string, error) {
	return listP2PCollections(t.txn, t.txnContext(), t.db.identity)
}

// AddP2PDocuments subscribes to the given document ID topics.
func (t *Transaction) AddP2PDocuments(docIDsJSON string) error {
	return addP2PDocuments(t.txn, t.txnContext(), t.db.identity, docIDsJSON)
}

// DeleteP2PDocuments unsubscribes from the given document ID topics.
func (t *Transaction) DeleteP2PDocuments(docIDsJSON string) error {
	return deleteP2PDocuments(t.txn, t.txnContext(), t.db.identity, docIDsJSON)
}

// ListP2PDocuments returns the list of subscribed document IDs as a JSON array.
func (t *Transaction) ListP2PDocuments() (string, error) {
	return listP2PDocuments(t.txn, t.txnContext(), t.db.identity)
}

// SyncDocuments fetches the latest versions of the given documents from the network.
func (t *Transaction) SyncDocuments(collectionName string, docIDsJSON string, timeoutMs int) error {
	return syncDocuments(t.txn, t.txnContext(), t.db.identity, collectionName, docIDsJSON, timeoutMs)
}

// SyncCollectionVersions synchronizes the given collection versions from the network.
func (t *Transaction) SyncCollectionVersions(versionIDsJSON string, timeoutMs int) error {
	return syncCollectionVersions(t.txn, t.txnContext(), t.db.identity, versionIDsJSON, timeoutMs)
}

// SyncBranchableCollection syncs a branchable collection's DAG from the network.
func (t *Transaction) SyncBranchableCollection(collectionID string, timeoutMs int) error {
	return syncBranchableCollection(t.txn, t.txnContext(), t.db.identity, collectionID, timeoutMs)
}
