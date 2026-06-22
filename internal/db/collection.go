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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/lens"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.Collection = (*collection)(nil)

// collection stores data records at Documents, which are gathered
// together under a collection name. This is analogous to SQL Tables.
type collection struct {
	db      *DB
	def     client.CollectionVersion
	indexes []CollectionIndex
	// indexBuildStates holds the build (backfill) state of the collection's indexes that have one,
	// keyed by index ID: an index is present while it is building or failed, and absent once ready.
	// Presence therefore means "not yet queryable". Populated at construction time from the index
	// state store; a rebuild's concurrent drop is not a build record and never appears here.
	indexBuildStates map[uint32]indexState
	fetcherFactory   func() fetcher.Fetcher
	txn              immutable.Option[datastore.Txn]
}

// @todo: Move the base Descriptions to an internal API within the db/ package.
// @body: Currently, the New/Create Collection APIs accept CollectionVersions
// as params. We want these Descriptions objects to be low level descriptions, and
// to be auto generated based on a more controllable and user friendly
// CollectionOptions object.

// newCollection returns a pointer to a newly instantiated DB Collection.
//
// Index instances are only constructed for indexes whose status is building or ready;
// failed and dropping indexes are excluded from the write path.
func (db *DB) newCollection(
	ctx context.Context,
	desc client.CollectionVersion,
	txn immutable.Option[datastore.Txn],
) (*collection, error) {
	col := &collection{
		db:               db,
		def:              desc,
		txn:              txn,
		indexBuildStates: make(map[uint32]indexState),
	}

	if len(desc.Indexes) > 0 {
		// Build a read context that has a txn set so getIndexBuildStates can call CtxMustGetTxn.
		stateCtx := ctx
		if txn.HasValue() {
			stateCtx = datastore.CtxSetTxn(ctx, txn.Value())
		}

		states, err := getIndexBuildStates(stateCtx, desc.CollectionID)
		if err != nil {
			return nil, err
		}
		col.indexBuildStates = states

		for _, index := range desc.Indexes {
			state := states[index.ID]

			// A failed index is abandoned: it is not maintained by writes, so it is left out of
			// c.indexes. A building index is kept, so its new epoch keeps receiving concurrent
			// writes while the backfill runs.
			if state.isFailed() {
				continue
			}

			colIndex, err := NewCollectionIndex(stateCtx, col, index, state.isBuilding())
			if err != nil {
				return nil, err
			}

			col.indexes = append(col.indexes, colIndex)
		}
	}

	return col, nil
}

// QueryableIndexes returns the indexes that are safe for query planning. An index is excluded
// while it has a build record (building or failed), since its entries may be incomplete.
// c.indexBuildStates holds only build records, so an index collecting a superseded epoch after a
// rebuild is absent here and stays queryable — it is already complete on its new epoch.
func (c *collection) QueryableIndexes() []client.IndexDescription {
	all := c.Version().Indexes
	result := make([]client.IndexDescription, 0, len(all))
	for _, idx := range all {
		if _, ok := c.indexBuildStates[idx.ID]; !ok {
			result = append(result, idx)
		}
	}
	return result
}

// newFetcher returns a new fetcher instance for this collection.
// If a fetcherFactory is set, it will be used to create the fetcher.
// It's a very simple factory, but it allows us to inject a mock fetcher
// for testing.
func (c *collection) newFetcher(ctx context.Context) fetcher.Fetcher {
	var innerFetcher fetcher.Fetcher
	if c.fetcherFactory != nil {
		innerFetcher = c.fetcherFactory()
	} else {
		innerFetcher = fetcher.NewDocumentFetcher()
	}

	return lens.NewFetcher(innerFetcher, c.db.getLensStore(ctx), c.db.collectionRepository)
}

// getCollectionByName returns an existing collection within the database.
func (db *DB) getCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	if name == "" {
		return nil, ErrCollectionNameEmpty
	}

	cols, err := db.getCollections(ctx, utils.NewOptions(options.GetCollections().SetCollectionName(name)), true)
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, client.ErrCollectionNotFound
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

// getCollections returns all collections and their descriptions matching the given options
// that currently exist within this [Store].
//
// Inactive collections are not returned by default unless a specific collection version ID
// is provided.
//
// txnIsEphemeral indicates whether or not the txn should be attached to the collection
func (db *DB) getCollections(
	ctx context.Context,
	opts *options.GetCollectionsOptions,
	txnIsEphemeral bool,
) ([]client.Collection, error) {
	if opts == nil {
		opts = &options.GetCollectionsOptions{}
	}

	var cols []client.CollectionVersion
	switch {
	case opts.CollectionName.HasValue() && !opts.GetInactive.Value():
		col, err := description.GetCollectionByName(ctx, db.collectionRepository, opts.CollectionName.Value())
		if err != nil && !errors.Is(err, client.ErrCollectionNotFound) {
			return nil, err
		}
		cols = append(cols, col)

	case opts.VersionID.HasValue():
		col, err := description.GetCollectionByID(ctx, db.collectionRepository, opts.VersionID.Value())
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)

	case opts.CollectionID.HasValue():
		if opts.GetInactive.HasValue() && opts.GetInactive.Value() {
			var err error
			cols, err = description.GetCollectionsByCollectionID(ctx, db.collectionRepository, opts.CollectionID.Value())
			if err != nil {
				return nil, err
			}
		} else {
			// GetActiveCollectionByCollectionID is quite a lot more efficient than GetCollectionsByCollectionID
			// so we use it when we can.
			col, err := description.GetActiveCollectionByCollectionID(ctx, db.collectionRepository, opts.CollectionID.Value())
			if err != nil && !errors.Is(err, client.ErrCollectionNotFound) {
				return nil, err
			}
			cols = append(cols, col)
		}

	// Multi-collection self-referencing relations are the only time the collection set id option
	// will be provided by internal code - it is expected that very few user collections will result
	// in this being called, and so for now we tolerate a full scan plus filter instead of maintaining
	// and index. The commented out case below, highlights its omission - if we want to index it in the
	// future it should be uncommented and handled.
	// case opts.CollectionSetID.HasValue():

	default:
		if opts.GetInactive.HasValue() && opts.GetInactive.Value() {
			var err error
			cols, err = description.GetCollections(ctx, db.collectionRepository)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			cols, err = description.GetActiveCollections(ctx, db.collectionRepository)
			if err != nil {
				return nil, err
			}
		}
	}

	collections := []client.Collection{}
	for _, col := range cols {
		if opts.VersionID.HasValue() {
			if col.VersionID != opts.VersionID.Value() {
				continue
			}
		}

		if opts.CollectionName.HasValue() {
			if col.Name != opts.CollectionName.Value() {
				continue
			}
		}

		// By default, we don't return inactive collections unless a specific version is requested.
		if !opts.GetInactive.Value() && !col.IsActive && !opts.VersionID.HasValue() {
			continue
		}

		if opts.CollectionSetID.HasValue() {
			if !col.CollectionSet.HasValue() {
				continue
			}

			if col.CollectionSet.Value().CollectionSetID != opts.CollectionSetID.Value() {
				continue
			}
		}

		// In the case that the txn was ephemeral, we will not save a reference to it
		// attached to the collection.
		var txnOpt immutable.Option[datastore.Txn]
		if txnIsEphemeral {
			txnOpt = immutable.None[datastore.Txn]()
		} else {
			txnOpt = datastore.CtxTryGetTxnOption(ctx)
		}
		collection, err := db.newCollection(ctx, col, txnOpt)
		if err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

// addCollection takes the provided SDL, and applies it to the database,
// adding the necessary collections, request types, etc.
func (db *DB) addCollection(
	ctx context.Context,
	sdl string,
) ([]client.CollectionVersion, error) {
	newDefinitions, err := db.parser.ParseSDL(ctx, sdl)
	if err != nil {
		return nil, err
	}

	result, err := db.addCollections(ctx, newDefinitions)
	if err != nil {
		return nil, err
	}

	err = db.loadCollectionDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *DB) loadCollectionDefinitions(ctx context.Context) error {
	definitions, err := description.GetActiveCollections(ctx, db.collectionRepository)
	if err != nil {
		return err
	}

	return db.parser.SetSchema(ctx, definitions)
}

// getTxnAndSetCtxForCollection is a helper function that checks if a transaction is attached to the context
// or the collection, and if so, attaches it to the context. It also returns a boolean indicating if a
// transaction was found.
func getTxnAndSetCtxForCollection(ctx context.Context, c *collection) (context.Context, datastore.Txn, bool) {
	txn, hadTxn := datastore.CtxTryGetTxn(ctx)
	if !hadTxn && c.txn.HasValue() {
		hadTxn = true
		txn = c.txn.Value()
		ctx = datastore.CtxSetTxn(ctx, txn)
	}
	return ctx, txn, hadTxn
}

// Version returns the client.CollectionVersion.
func (c *collection) Version() client.CollectionVersion {
	return c.def
}

// Name returns the collection name.
func (c *collection) Name() string {
	return c.Version().Name
}

// VersionID returns the VersionID of the collection.
func (c *collection) VersionID() string {
	return c.Version().VersionID
}

func (c *collection) CollectionID() string {
	return c.Version().CollectionID
}
