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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
)

// ---------- Store methods on DB ----------

// AddCollection creates new collection types from the given SDL string.
// Returns a JSON array of CollectionVersion objects.
func (db *DB) AddCollection(sdl string) (string, error) {
	return addCollection(db.node.DB, db.context(), db.identity, sdl)
}

func addCollection(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	sdl string,
) (string, error) {
	opt := options.AddCollection()
	setOptIdentity(opt, identity)
	cols, err := store.AddCollection(ctx, sdl, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(cols)
}

// PatchCollection applies a JSON patch string to the collection definitions.
// migrationJSON may be an empty string (no migration) or a JSON-encoded model.Lens.
func (db *DB) PatchCollection(patch string, migrationJSON string) error {
	return patchCollection(db.node.DB, db.context(), db.identity, patch, migrationJSON)
}

func patchCollection(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	patch string,
	migrationJSON string,
) error {
	var migration immutable.Option[model.Lens]
	if migrationJSON != "" {
		var lens model.Lens
		if err := json.Unmarshal([]byte(migrationJSON), &lens); err != nil {
			return err
		}
		migration = immutable.Some(lens)
	}
	opt := options.PatchCollection()
	setOptIdentity(opt, identity)
	return store.PatchCollection(ctx, patch, migration, opt)
}

// SetActiveCollectionVersion activates the collection version with the given versionID.
func (db *DB) SetActiveCollectionVersion(versionID string) error {
	return setActiveCollectionVersion(db.node.DB, db.context(), db.identity, versionID)
}

func setActiveCollectionVersion(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	versionID string,
) error {
	opt := options.SetActiveCollectionVersion()
	setOptIdentity(opt, identity)
	return store.SetActiveCollectionVersion(ctx, versionID, opt)
}

// AddView creates a new DefraDB view from the given GQL query and SDL.
// transformCID may be empty (no transform). Returns a JSON array of CollectionVersion.
func (db *DB) AddView(gqlQuery string, sdl string, transformCID string) (string, error) {
	return addView(db.node.DB, db.context(), db.identity, gqlQuery, sdl, transformCID)
}

func addView(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	gqlQuery string,
	sdl string,
	transformCID string,
) (string, error) {
	opt := options.AddView()
	setOptIdentity(opt, identity)
	if transformCID != "" {
		opt.SetTransformCID(transformCID)
	}
	cols, err := store.AddView(ctx, gqlQuery, sdl, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(cols)
}

// RefreshViews refreshes materialized view caches matching the given options.
// Empty strings are treated as "not set" (no filter applied).
func (db *DB) RefreshViews(collectionName string, versionID string, collectionID string, getInactive bool) error {
	return refreshViews(db.node.DB, db.context(), db.identity, collectionName, versionID, collectionID, getInactive)
}

func refreshViews(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionName string,
	versionID string,
	collectionID string,
	getInactive bool,
) error {
	opt := options.RefreshViews()
	setOptIdentity(opt, identity)
	if collectionName != "" {
		opt.SetCollectionName(collectionName)
	}
	if versionID != "" {
		opt.SetVersionID(versionID)
	}
	if collectionID != "" {
		opt.SetCollectionID(collectionID)
	}
	if getInactive {
		opt.SetGetInactive(true)
	}
	return store.RefreshViews(ctx, opt)
}

// SetMigration sets the migration for the collection versions described in configJSON.
// configJSON is a JSON-encoded client.LensConfig. Returns the lens ID string.
func (db *DB) SetMigration(configJSON string) (string, error) {
	return setMigration(db.node.DB, db.context(), db.identity, configJSON)
}

func setMigration(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	configJSON string,
) (string, error) {
	var config client.LensConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return "", err
	}
	opt := options.SetMigration()
	setOptIdentity(opt, identity)
	return store.SetMigration(ctx, config, opt)
}

// AddLens stores a lens configuration and returns its CID.
// lensJSON is a JSON-encoded model.Lens.
func (db *DB) AddLens(lensJSON string) (string, error) {
	return addLens(db.node.DB, db.context(), db.identity, lensJSON)
}

func addLens(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	lensJSON string,
) (string, error) {
	var lens model.Lens
	if err := json.Unmarshal([]byte(lensJSON), &lens); err != nil {
		return "", err
	}
	opt := options.AddLens()
	setOptIdentity(opt, identity)
	return store.AddLens(ctx, lens, opt)
}

// ListLenses returns all stored lenses as a JSON object mapping CID to model.Lens.
func (db *DB) ListLenses() (string, error) {
	return listLenses(db.node.DB, db.context(), db.identity)
}

func listLenses(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.ListLenses()
	setOptIdentity(opt, identity)
	lenses, err := store.ListLenses(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(lenses)
}

// GetCollectionByName returns a Collection wrapper for the named collection.
func (db *DB) GetCollectionByName(name string) (*Collection, error) {
	opt := options.GetCollectionByName()
	setOptIdentity(opt, db.identity)
	col, err := db.node.DB.GetCollectionByName(db.context(), name, opt)
	if err != nil {
		return nil, err
	}
	return &Collection{col: col, db: db}, nil
}

// GetCollections returns a JSON array of collection descriptions matching the given options.
// Empty strings are treated as "not set".
func (db *DB) GetCollections(collectionName string, versionID string, collectionID string, getInactive bool) (string, error) {
	return getCollections(db.node.DB, db.context(), db.identity, collectionName, versionID, collectionID, getInactive)
}

func getCollections(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionName string,
	versionID string,
	collectionID string,
	getInactive bool,
) (string, error) {
	opt := options.GetCollections()
	setOptIdentity(opt, identity)
	if collectionName != "" {
		opt.SetCollectionName(collectionName)
	}
	if versionID != "" {
		opt.SetVersionID(versionID)
	}
	if collectionID != "" {
		opt.SetCollectionID(collectionID)
	}
	if getInactive {
		opt.SetGetInactive(true)
	}
	cols, err := store.GetCollections(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(cols)
}

// ListIndexes returns all indexes as a JSON object mapping collection name to index descriptions.
func (db *DB) ListIndexes() (string, error) {
	return listIndexes(db.node.DB, db.context(), db.identity)
}

func listIndexes(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.ListIndexes()
	setOptIdentity(opt, identity)
	indexes, err := store.ListIndexes(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(indexes)
}

// ListAllEncryptedIndexes returns all encrypted indexes as a JSON object.
func (db *DB) ListAllEncryptedIndexes() (string, error) {
	return listAllEncryptedIndexes(db.node.DB, db.context(), db.identity)
}

func listAllEncryptedIndexes(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.ListAllEncryptedIndexes()
	setOptIdentity(opt, identity)
	indexes, err := store.ListAllEncryptedIndexes(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(indexes)
}

// ExecRequest executes a GQL request and returns a JSON-encoded result.
// operationName and variablesJSON may be empty.
// variablesJSON should be a JSON object of variable name to value mappings.
func (db *DB) ExecRequest(request string, operationName string, variablesJSON string) (string, error) {
	return execRequest(db.node.DB, db.context(), db.identity, request, operationName, variablesJSON)
}

func execRequest(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	request string,
	operationName string,
	variablesJSON string,
) (string, error) {
	opt := options.ExecRequest()
	setOptIdentity(opt, identity)
	if operationName != "" {
		opt.SetOperationName(operationName)
	}
	if variablesJSON != "" {
		var variables map[string]any
		if err := json.Unmarshal([]byte(variablesJSON), &variables); err != nil {
			return "", err
		}
		opt.SetVariables(variables)
	}
	res := store.ExecRequest(ctx, request, opt)
	if res.Subscription != nil {
		// Drain channel to prevent goroutine leak.
		go func() {
			for range res.Subscription {
			}
		}()
		return "", ErrSubscriptionsNotSupported
	}
	return marshalJSON(res.GQL)
}

// BasicImport imports a JSON dataset from the given filepath.
func (db *DB) BasicImport(filepath string) error {
	return basicImport(db.node.DB, db.context(), filepath)
}

func basicImport(store client.Store, ctx context.Context, filepath string) error {
	return store.BasicImport(ctx, filepath)
}

// BasicExport exports the database to a JSON file at the given filepath.
func (db *DB) BasicExport(filepath string) error {
	return basicExport(db.node.DB, db.context(), filepath)
}

func basicExport(store client.Store, ctx context.Context, filepath string) error {
	return store.BasicExport(ctx, filepath)
}

// PrintDump logs the entire contents of the database store.
func (db *DB) PrintDump() error {
	return printDump(db.node.DB, db.context())
}

func printDump(store client.Store, ctx context.Context) error {
	return store.PrintDump(ctx)
}

// AddDACPolicy adds a document ACP policy. Returns a JSON-encoded AddPolicyResult.
func (db *DB) AddDACPolicy(policy string) (string, error) {
	return addDACPolicy(db.node.DB, db.context(), db.identity, policy)
}

func addDACPolicy(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	policy string,
) (string, error) {
	opt := options.AddDACPolicy()
	setOptIdentity(opt, identity)
	res, err := store.AddDACPolicy(ctx, policy, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(res)
}

// AddDACActorRelationship adds a relationship between a document and an actor.
// Returns a JSON-encoded AddActorRelationshipResult.
func (db *DB) AddDACActorRelationship(collectionName string, docID string, relation string, targetActor string) (string, error) {
	return addDACActorRelationship(db.node.DB, db.context(), db.identity, collectionName, docID, relation, targetActor)
}

func addDACActorRelationship(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (string, error) {
	opt := options.AddDACActorRelationship()
	setOptIdentity(opt, identity)
	res, err := store.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(res)
}

// DeleteDACActorRelationship deletes a relationship between a document and an actor.
// Returns a JSON-encoded DeleteActorRelationshipResult.
func (db *DB) DeleteDACActorRelationship(collectionName string, docID string, relation string, targetActor string) (string, error) {
	return deleteDACActorRelationship(db.node.DB, db.context(), db.identity, collectionName, docID, relation, targetActor)
}

func deleteDACActorRelationship(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (string, error) {
	opt := options.DeleteDACActorRelationship()
	setOptIdentity(opt, identity)
	res, err := store.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(res)
}

// AddNACActorRelationship adds a node access relationship for the given actor.
// Returns a JSON-encoded AddActorRelationshipResult.
func (db *DB) AddNACActorRelationship(relation string, targetActor string) (string, error) {
	return addNACActorRelationship(db.node.DB, db.context(), db.identity, relation, targetActor)
}

func addNACActorRelationship(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	relation string,
	targetActor string,
) (string, error) {
	opt := options.AddNACActorRelationship()
	setOptIdentity(opt, identity)
	res, err := store.AddNACActorRelationship(ctx, relation, targetActor, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(res)
}

// DeleteNACActorRelationship deletes a node access relationship for the given actor.
// Returns a JSON-encoded DeleteActorRelationshipResult.
func (db *DB) DeleteNACActorRelationship(relation string, targetActor string) (string, error) {
	return deleteNACActorRelationship(db.node.DB, db.context(), db.identity, relation, targetActor)
}

func deleteNACActorRelationship(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	relation string,
	targetActor string,
) (string, error) {
	opt := options.DeleteNACActorRelationship()
	setOptIdentity(opt, identity)
	res, err := store.DeleteNACActorRelationship(ctx, relation, targetActor, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(res)
}

// ReEnableNAC re-enables node access control.
func (db *DB) ReEnableNAC() error {
	return reEnableNAC(db.node.DB, db.context(), db.identity)
}

func reEnableNAC(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) error {
	opt := options.ReEnableNAC()
	setOptIdentity(opt, identity)
	return store.ReEnableNAC(ctx, opt)
}

// DisableNAC temporarily disables node access control.
func (db *DB) DisableNAC() error {
	return disableNAC(db.node.DB, db.context(), db.identity)
}

func disableNAC(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) error {
	opt := options.DisableNAC()
	setOptIdentity(opt, identity)
	return store.DisableNAC(ctx, opt)
}

// GetNACStatus returns the node access control status as a JSON string.
func (db *DB) GetNACStatus() (string, error) {
	return getNACStatus(db.node.DB, db.context(), db.identity)
}

func getNACStatus(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
) (string, error) {
	opt := options.GetNACStatus()
	setOptIdentity(opt, identity)
	res, err := store.GetNACStatus(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(res)
}

// GetNodeIdentity returns the node's public identity as a JSON string,
// or an empty string if no identity is configured.
func (db *DB) GetNodeIdentity() (string, error) {
	return getNodeIdentity(db.node.DB, db.context())
}

func getNodeIdentity(store client.Store, ctx context.Context) (string, error) {
	res, err := store.GetNodeIdentity(ctx)
	if err != nil {
		return "", err
	}
	if !res.HasValue() {
		return "", nil
	}
	return marshalJSON(res.Value())
}

// VerifySignature verifies the signature of a block using the given public key.
// publicKeyType defaults to "secp256k1" if empty.
func (db *DB) VerifySignature(publicKeyHex string, publicKeyType string, blockCID string) error {
	return verifySignature(db.node.DB, db.context(), db.identity, publicKeyHex, publicKeyType, blockCID)
}

func verifySignature(
	store client.Store,
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	publicKeyHex string,
	publicKeyType string,
	blockCID string,
) error {
	if publicKeyType == "" {
		publicKeyType = string(crypto.KeyTypeSecp256k1)
	}
	pubKey, err := crypto.PublicKeyFromString(crypto.KeyType(publicKeyType), publicKeyHex)
	if err != nil {
		return err
	}
	opt := options.VerifySignature()
	setOptIdentity(opt, identity)
	return store.VerifySignature(ctx, blockCID, pubKey, opt)
}

// NewTxn creates a new (non-concurrent) transaction.
func (db *DB) NewTxn(readOnly bool) (*Transaction, error) {
	txn, err := db.node.DB.NewTxn(readOnly)
	if err != nil {
		return nil, err
	}
	return &Transaction{txn: txn, db: db}, nil
}

// NewConcurrentTxn creates a new concurrent (thread-safe) transaction.
func (db *DB) NewConcurrentTxn(readOnly bool) (*Transaction, error) {
	txn, err := db.node.DB.NewConcurrentTxn(readOnly)
	if err != nil {
		return nil, err
	}
	return &Transaction{txn: txn, db: db}, nil
}
