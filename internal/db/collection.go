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
	db             *DB
	def            client.CollectionVersion
	indexes        []CollectionIndex
	fetcherFactory func() fetcher.Fetcher
	txn            immutable.Option[client.Txn]
}

// @todo: Move the base Descriptions to an internal API within the db/ package.
// @body: Currently, the New/Create Collection APIs accept CollectionVersions
// as params. We want these Descriptions objects to be low level descriptions, and
// to be auto generated based on a more controllable and user friendly
// CollectionOptions object.

// newCollection returns a pointer to a newly instantiated DB Collection
func (db *DB) newCollection(desc client.CollectionVersion, txn immutable.Option[client.Txn]) (*collection, error) {
	col := &collection{
		db:  db,
		def: desc,
		txn: txn,
	}
	for _, index := range desc.Indexes {
		colIndex, err := NewCollectionIndex(col, index)
		if err != nil {
			return nil, err
		}
		col.indexes = append(col.indexes, colIndex)
	}
	return col, nil
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

	return lens.NewFetcher(innerFetcher, c.db.getLensStore(ctx))
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
		col, err := description.GetCollectionByName(ctx, opts.CollectionName.Value())
		if err != nil && !errors.Is(err, client.ErrCollectionNotFound) {
			return nil, err
		}
		cols = append(cols, col)

	case opts.VersionID.HasValue():
		col, err := description.GetCollectionByID(ctx, opts.VersionID.Value())
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)

	case opts.CollectionID.HasValue():
		var err error
		cols, err = description.GetCollectionsByCollectionID(ctx, opts.CollectionID.Value())
		if err != nil {
			return nil, err
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
			cols, err = description.GetCollections(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			cols, err = description.GetActiveCollections(ctx)
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
		var txnOpt immutable.Option[client.Txn]
		if txnIsEphemeral {
			txnOpt = immutable.None[client.Txn]()
		} else {
			txnOpt = datastore.CtxTryGetClientTxnOption(ctx)
		}
		collection, err := db.newCollection(col, txnOpt)
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
	definitions, err := description.GetActiveCollections(ctx)
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
		txn = c.txn.Value().(datastore.Txn)
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
