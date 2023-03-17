// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

var _ client.DB = (*implicitTxnDB)(nil)
var _ client.DB = (*explicitTxnDB)(nil)
var _ client.Store = (*implicitTxnDB)(nil)
var _ client.Store = (*explicitTxnDB)(nil)

type implicitTxnDB struct {
	*db
}

type explicitTxnDB struct {
	*db
	txn datastore.Txn
}

func (db *implicitTxnDB) CreateCollection(
	ctx context.Context,
	desc client.CollectionDescription,
) (client.Collection, error) {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	col, err := db.createCollection(ctx, txn, desc)
	if err != nil {
		return nil, err
	}

	err = txn.Commit(ctx)
	return col, err
}

func (db *explicitTxnDB) CreateCollection(
	ctx context.Context,
	desc client.CollectionDescription,
) (client.Collection, error) {
	return db.createCollection(ctx, db.txn, desc)
}

// ExecRequest executes a request against the database.
func (db *implicitTxnDB) ExecRequest(ctx context.Context, request string) *client.RequestResult {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		res := &client.RequestResult{}
		res.GQL.Errors = []any{err.Error()}
		return res
	}
	defer txn.Discard(ctx)

	res := db.execRequest(ctx, request, txn)

	if err := txn.Commit(ctx); err != nil {
		res.GQL.Errors = []any{err.Error()}
		return res
	}

	return res
}

// ExecRequest executes a transaction request against the database.
func (db *explicitTxnDB) ExecRequest(
	ctx context.Context,
	request string,
) *client.RequestResult {
	return db.execRequest(ctx, request, db.txn)
}

// GetCollectionByName returns an existing collection within the database.
func (db *implicitTxnDB) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getCollectionByName(ctx, txn, name)
}

// GetCollectionByName returns an existing collection within the database.
func (db *explicitTxnDB) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	return db.getCollectionByName(ctx, db.txn, name)
}

// GetCollectionBySchemaID returns an existing collection using the schema hash ID.
func (db *implicitTxnDB) GetCollectionBySchemaID(
	ctx context.Context,
	schemaID string,
) (client.Collection, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getCollectionBySchemaID(ctx, txn, schemaID)
}

// GetCollectionBySchemaID returns an existing collection using the schema hash ID.
func (db *explicitTxnDB) GetCollectionBySchemaID(
	ctx context.Context,
	schemaID string,
) (client.Collection, error) {
	return db.getCollectionBySchemaID(ctx, db.txn, schemaID)
}

// AddP2PCollection adds the given collection ID that the P2P system
// subscribes to to the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *implicitTxnDB) AddP2PCollection(ctx context.Context, collectionID string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.addP2PCollection(ctx, txn, collectionID)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// AddP2PCollection adds the given collection ID that the P2P system
// subscribes to to the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *explicitTxnDB) AddP2PCollection(ctx context.Context, collectionID string) error {
	return db.addP2PCollection(ctx, db.txn, collectionID)
}

// RemoveP2PCollection removes the given collection ID that the P2P system
// subscribes to from the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *implicitTxnDB) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.removeP2PCollection(ctx, txn, collectionID)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// RemoveP2PCollection removes the given collection ID that the P2P system
// subscribes to from the the persisted list. It will error if the provided
// collection ID is invalid.
func (db *explicitTxnDB) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	return db.removeP2PCollection(ctx, db.txn, collectionID)
}

// GetAllCollections gets all the currently defined collections.
func (db *implicitTxnDB) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getAllCollections(ctx, txn)
}

// GetAllCollections gets all the currently defined collections.
func (db *explicitTxnDB) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	return db.getAllCollections(ctx, db.txn)
}

// AddSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *implicitTxnDB) AddSchema(ctx context.Context, schemaString string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.addSchema(ctx, txn, schemaString)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// AddSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *explicitTxnDB) AddSchema(ctx context.Context, schemaString string) error {
	return db.addSchema(ctx, db.txn, schemaString)
}

// PatchSchema takes the given JSON patch string and applies it to the set of CollectionDescriptions
// present in the database.
//
// It will also update the GQL types used by the query system. It will error and not apply any of the
// requested, valid updates should the net result of the patch result in an invalid state.  The
// individual operations defined in the patch do not need to result in a valid state, only the net result
// of the full patch.
//
// The collections (including the schema version ID) will only be updated if any changes have actually
// been made, if the net result of the patch matches the current persisted description then no changes
// will be applied.
func (db *implicitTxnDB) PatchSchema(ctx context.Context, patchString string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.patchSchema(ctx, txn, patchString)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// PatchSchema takes the given JSON patch string and applies it to the set of CollectionDescriptions
// present in the database.
//
// It will also update the GQL types used by the query system. It will error and not apply any of the
// requested, valid updates should the net result of the patch result in an invalid state.  The
// individual operations defined in the patch do not need to result in a valid state, only the net result
// of the full patch.
//
// The collections (including the schema version ID) will only be updated if any changes have actually
// been made, if the net result of the patch matches the current persisted description then no changes
// will be applied.
func (db *explicitTxnDB) PatchSchema(ctx context.Context, patchString string) error {
	return db.patchSchema(ctx, db.txn, patchString)
}

// UpdateCollection updates the persisted collection description matching the name of the given
// description, to the values in the given description.
//
// It will validate the given description using [ValidateUpdateCollection] before updating it.
//
// The collection (including the schema version ID) will only be updated if any changes have actually
// been made, if the given description matches the current persisted description then no changes will be
// applied.
func (db *implicitTxnDB) UpdateCollection(
	ctx context.Context,
	desc client.CollectionDescription,
) (client.Collection, error) {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	col, err := db.updateCollection(ctx, txn, desc)
	if err != nil {
		return nil, err
	}

	return col, txn.Commit(ctx)
}

// UpdateCollection updates the persisted collection description matching the name of the given
// description, to the values in the given description.
//
// It will validate the given description using [ValidateUpdateCollection] before updating it.
//
// The collection (including the schema version ID) will only be updated if any changes have actually
// been made, if the given description matches the current persisted description then no changes will be
// applied.
func (db *explicitTxnDB) UpdateCollection(
	ctx context.Context,
	desc client.CollectionDescription,
) (client.Collection, error) {
	return db.updateCollection(ctx, db.txn, desc)
}

// ValidateUpdateCollection validates that the given collection description is a valid update.
//
// Will return true if the given desctiption differs from the current persisted state of the
// collection. Will return an error if it fails validation.
func (db *implicitTxnDB) ValidateUpdateCollection(
	ctx context.Context,
	proposedDesc client.CollectionDescription,
) (bool, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return false, err
	}
	defer txn.Discard(ctx)

	return db.validateUpdateCollection(ctx, txn, proposedDesc)
}

// ValidateUpdateCollection validates that the given collection description is a valid update.
//
// Will return true if the given desctiption differs from the current persisted state of the
// collection. Will return an error if it fails validation.
func (db *explicitTxnDB) ValidateUpdateCollection(
	ctx context.Context,
	proposedDesc client.CollectionDescription,
) (bool, error) {
	return db.validateUpdateCollection(ctx, db.txn, proposedDesc)
}

// SetReplicator adds a new replicator to the database.
func (db *implicitTxnDB) SetReplicator(ctx context.Context, rep client.Replicator) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.setReplicator(ctx, txn, rep)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// SetReplicator adds a new replicator to the database.
func (db *explicitTxnDB) SetReplicator(ctx context.Context, rep client.Replicator) error {
	return db.setReplicator(ctx, db.txn, rep)
}

// DeleteReplicator removes a replicator from the database.
func (db *implicitTxnDB) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.deleteReplicator(ctx, txn, rep)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// DeleteReplicator removes a replicator from the database.
func (db *explicitTxnDB) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	return db.deleteReplicator(ctx, db.txn, rep)
}

// GetAllReplicators returns all replicators of the database.
func (db *implicitTxnDB) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getAllReplicators(ctx, txn)
}

// GetAllReplicators returns all replicators of the database.
func (db *explicitTxnDB) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	return db.getAllReplicators(ctx, db.txn)
}

// GetAllP2PCollections returns the list of persisted collection IDs that
// the P2P system subscribes to.
func (db *implicitTxnDB) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getAllP2PCollections(ctx, txn)
}

// GetAllP2PCollections returns the list of persisted collection IDs that
// the P2P system subscribes to.
func (db *explicitTxnDB) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	return db.getAllP2PCollections(ctx, db.txn)
}
