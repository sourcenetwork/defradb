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
	"fmt"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/action"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (db *DB) addView(
	ctx context.Context,
	inputQuery string,
	sdl string,
	transformCID immutable.Option[string],
) ([]client.CollectionVersion, error) {
	// Wrap the given query as part of the GQL query object - this simplifies the syntax for users
	// and ensures that we can't be given mutations.  In the future this line should disappear along
	// with the all calls to the parser appart from `ParseSDL` when we implement the DQL stuff.
	query := fmt.Sprintf(`query { %s }`, inputQuery)

	parseResults, err := db.parser.ParseSDL(ctx, sdl)
	if err != nil {
		return nil, err
	}

	ast, err := db.parser.BuildRequestAST(ctx, query)
	if err != nil {
		return nil, err
	}

	req, errs := db.parser.Parse(ctx, ast, &client.GQLOptions{})
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	if len(req.Queries) == 0 || len(req.Queries[0].Selections) == 0 {
		return nil, NewErrInvalidViewQueryMissingQuery()
	}

	baseQuery, ok := req.Queries[0].Selections[0].(*request.Select)
	if !ok {
		return nil, NewErrInvalidViewQueryCastFailed(inputQuery)
	}

	for i := range parseResults {
		var lensID immutable.Option[string]
		if transformCID.HasValue() {
			exists, err := db.lensCIDExists(ctx, transformCID.Value())
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, NewErrLensCIDNotFound(transformCID.Value())
			}
			lensID = transformCID
		}

		source := client.QuerySource{
			Query:     *baseQuery,
			Transform: lensID,
		}
		parseResults[i].Definition.Query = immutable.Some(source)
	}

	returnDescriptions, err := db.addCollections(ctx, parseResults)
	if err != nil {
		return nil, err
	}

	err = db.loadCollectionDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	// The materialized view caches are refreshed by the caller (AddView) after the collection
	// metadata has been committed, so the refresh - which writes txn-free - never runs while a
	// transaction is open. See https://github.com/sourcenetwork/defradb/issues/4959.
	return returnDescriptions, nil
}

// refreshViewsInTxn refreshes views using the explicit transaction already on the context. The
// action-execution markers are written txn-free (corekv issue #107) while that transaction is open,
// so this path is unsupported on leveldb - the txn-free write blocks on the open transaction (see
// issue #4959). It preserves the original behaviour for the other stores.
func (db *DB) refreshViewsInTxn(ctx context.Context, opts *options.GetCollectionsOptions) error {
	// For now, we only support user-cache management of views, not all collections
	cols, err := db.getViews(ctx, opts)
	if err != nil {
		return err
	}

	txn := datastore.CtxMustGetTxn(ctx)

	// Clear the transaction on the context used to write the action execution information, otherwise
	// corekv will pick it up again, writing using the transaction.
	// https://github.com/sourcenetwork/corekv/issues/107
	writeCtx := datastore.CtxSetTxn(ctx, nil)

	for _, col := range cols {
		if !col.IsMaterialized {
			// We only care about materialized views here, so skip any that aren't
			continue
		}

		shortID, err := id.GetShortCollectionID(ctx, col.CollectionID)
		if err != nil {
			return err
		}
		db.lockSet.CollectionLock(txn, shortID)

		colObject, err := db.newCollection(col, immutable.Some(txn))
		if err != nil {
			return err
		}

		multistore := datastore.NewMultistore(db.rootstore, db.lockSet, db.blockStoreChunkSize)

		err = action.Register(writeCtx, multistore, db.events, col.CollectionID, client.RefreshDatastoreAction)
		if err != nil {
			return err
		}

		// Clearing and then constructing is a bit inefficient, but it should do for now.
		// Long term we probably want to update inline as much as possible to avoid unnessecarily
		// moving/adding/deleting keys in storage
		err = colObject.truncate(ctx)
		if err != nil {
			errErr := action.Set(
				writeCtx,
				multistore,
				db.events,
				col.CollectionID,
				client.TruncateAction,
				client.ErroredActionStatus,
			)
			return errors.Join(errErr, err)
		}

		err = db.buildViewCache(ctx, col)
		if err != nil {
			errErr := action.Set(
				writeCtx,
				multistore,
				db.events,
				col.CollectionID,
				client.TruncateAction,
				client.ErroredActionStatus,
			)
			return errors.Join(errErr, err)
		}

		err = action.Complete(writeCtx, multistore, db.events, col.CollectionID, client.RefreshDatastoreAction)
		if err != nil {
			return err
		}
	}

	return nil
}

// refreshViewsImplicit refreshes views without an explicit transaction. Each view's action-execution
// markers are written with no transaction open, and the truncate-and-rebuild runs inside its own
// transaction. This keeps txn-free writes and an open transaction from ever overlapping, which is
// what leveldb requires (issue #4959), while still giving the rebuild planner a real transaction.
func (db *DB) refreshViewsImplicit(ctx context.Context, opts *options.GetCollectionsOptions) error {
	cols, err := db.listViewsToRefresh(ctx, opts)
	if err != nil {
		return err
	}

	for _, col := range cols {
		if !col.IsMaterialized {
			// We only care about materialized views here, so skip any that aren't
			continue
		}
		if err := db.refreshViewImplicit(ctx, col); err != nil {
			return err
		}
	}

	return nil
}

// listViewsToRefresh reads the set of views to refresh under a short-lived read transaction that is
// closed before any refresh begins, so no transaction is open when the action markers are written.
func (db *DB) listViewsToRefresh(
	ctx context.Context,
	opts *options.GetCollectionsOptions,
) ([]client.CollectionVersion, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.getViews(ctx, opts)
}

// refreshViewImplicit registers the refresh action (txn-free, with no transaction open), rebuilds the
// view inside its own transaction, then records completion (also txn-free, after the transaction has
// closed).
func (db *DB) refreshViewImplicit(ctx context.Context, col client.CollectionVersion) (err error) {
	multistore := datastore.NewMultistore(db.rootstore, db.lockSet, db.blockStoreChunkSize)

	err = action.Register(ctx, multistore, db.events, col.CollectionID, client.RefreshDatastoreAction)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			errErr := action.Set(
				ctx,
				multistore,
				db.events,
				col.CollectionID,
				client.TruncateAction,
				client.ErroredActionStatus,
			)
			err = errors.Join(errErr, err)
			return
		}
		err = action.Complete(ctx, multistore, db.events, col.CollectionID, client.RefreshDatastoreAction)
	}()

	return db.rebuildMaterializedView(ctx, col)
}

// rebuildMaterializedView clears and rebuilds a materialized view's cache inside its own transaction.
// The planner used by buildViewCache re-enters public APIs that require a real transaction, so this
// must run under one - which is fine, as it performs no txn-free writes.
func (db *DB) rebuildMaterializedView(ctx context.Context, col client.CollectionVersion) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	shortID, err := id.GetShortCollectionID(ctx, col.CollectionID)
	if err != nil {
		return err
	}
	db.lockSet.CollectionLock(txn, shortID)

	colObject, err := db.newCollection(col, immutable.Some[datastore.Txn](txn))
	if err != nil {
		return err
	}

	// Clearing and then constructing is a bit inefficient, but it should do for now.
	// Long term we probably want to update inline as much as possible to avoid unnessecarily
	// moving/adding/deleting keys in storage
	if err := colObject.truncate(ctx); err != nil {
		return err
	}

	if err := db.buildViewCache(ctx, col); err != nil {
		return err
	}

	return txn.Commit()
}

// refreshNewMaterializedViews refreshes the caches of the materialized views among defs, after
// AddView has created (and, in the implicit case, committed) their metadata.
func (db *DB) refreshNewMaterializedViews(
	ctx context.Context,
	defs []client.CollectionVersion,
	implicit bool,
) error {
	for _, view := range defs {
		if !view.Query.HasValue() || !view.IsMaterialized {
			continue
		}

		opts := utils.NewOptions(options.GetCollections().SetVersionID(view.VersionID))

		var err error
		if implicit {
			err = db.refreshViewsImplicit(ctx, opts)
		} else {
			err = db.refreshViewsInTxn(ctx, opts)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) getViews(ctx context.Context, opts *options.GetCollectionsOptions) ([]client.CollectionVersion, error) {
	cols, err := db.getCollections(ctx, opts, true)
	if err != nil {
		return nil, err
	}

	var views []client.CollectionVersion
	for _, col := range cols {
		if !col.Version().Query.HasValue() {
			continue
		}

		views = append(views, col.Version())
	}

	return views, nil
}

func (db *DB) buildViewCache(ctx context.Context, col client.CollectionVersion) (err error) {
	p := planner.New(
		ctx,
		identity.FromContext(ctx),
		db.nodeACP,
		db.documentACP,
		db,
		db.p2p,
		db.getLensStore(ctx),
		db.collectionRepository,
	)

	// temporarily disable the cache in order to query without using it
	col.IsMaterialized = false
	err = description.SaveCollection(ctx, db.collectionRepository, col)
	if err != nil {
		return err
	}
	defer func() {
		var defErr error
		col.IsMaterialized = true
		defErr = description.SaveCollection(ctx, db.collectionRepository, col)
		if err == nil {
			// Do not overwrite the original error if there is one, defErr is probably an artifact of the original
			// failue and can be discarded.
			err = defErr
		}
	}()

	request, err := db.generateMaximalSelectFromCollection(ctx, col, immutable.None[string](), map[string]struct{}{})
	if err != nil {
		return err
	}

	source, err := p.MakeSelectionPlan(request)
	if err != nil {
		return err
	}

	err = source.Init()
	if err != nil {
		return err
	}
	defer func() {
		defErr := source.Close()
		if err == nil {
			// Do not overwrite the original error if there is one, defErr is probably an artifact of the original
			// failue and can be discarded.
			err = defErr
		}
	}()

	err = source.Start()
	if err != nil {
		return err
	}

	hasValue, err := source.Next()
	if err != nil {
		return err
	}

	ds := datastore.NewMultistore(db.rootstore, db.lockSet, db.blockStoreChunkSize).Datastore()

	// View items are currently keyed by their index, starting at 1.
	// The order in which results are returned must be consistent with the results of the
	// underlying query/transform.
	var itemID uint
	for itemID = 1; hasValue; itemID++ {
		doc := source.Value()

		serializedItem, err := core.MarshalViewItem(doc)
		if err != nil {
			return err
		}

		shortID, err := id.GetShortCollectionID(ctx, col.CollectionID)
		if err != nil {
			return err
		}

		itemKey := keys.NewViewCacheKey(shortID, itemID)
		err = ds.Set(ctx, itemKey, serializedItem)
		if err != nil {
			return NewErrStoreViewCacheItem(err)
		}

		hasValue, err = source.Next()
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) generateMaximalSelectFromCollection(
	ctx context.Context,
	col client.CollectionVersion,
	fieldName immutable.Option[string],
	typesHit map[string]struct{},
) (*request.Select, error) {
	// `__-` is an impossible field name prefix, so we can safely concat using it as a separator without risk
	// of collision.
	identifier := col.Name + "__-" + fieldName.Value()
	if _, ok := typesHit[identifier]; ok {
		// If this identifier is already in the set, the collection type must be circular and we should return
		return nil, nil
	}
	typesHit[identifier] = struct{}{}

	childRequests := []request.Selection{}
	for _, field := range col.Fields {
		if field.RelationName.HasValue() && field.Kind.IsObject() {
			relatedCol, _, err := description.GetRelatedCollection(ctx, db.collectionRepository, col, field.Kind)
			if err != nil {
				return nil, err
			}

			innerSelect, err := db.generateMaximalSelectFromCollection(
				ctx,
				relatedCol,
				immutable.Some(field.Name),
				typesHit,
			)
			if err != nil {
				return nil, err
			}

			if innerSelect != nil {
				// innerSelect may be nil if a circular relationship is defined in the collection type and we have already
				// added this field
				childRequests = append(childRequests, innerSelect)
			}
		}
	}

	var name string
	if fieldName.HasValue() {
		name = fieldName.Value()
	} else {
		name = col.Name
	}

	return &request.Select{
		Field: request.Field{
			Name: name,
		},
		ChildSelect: request.ChildSelect{
			Fields: childRequests,
		},
	}, nil
}

// lensCIDExists checks if a lens with the given CID exists in the lens store.
func (db *DB) lensCIDExists(ctx context.Context, cidStr string) (bool, error) {
	_, err := cid.Decode(cidStr)
	if err != nil {
		return false, err
	}

	lenses, err := db.getLensStore(ctx).List(ctx)
	if err != nil {
		return false, err
	}

	for storedCID := range lenses {
		if storedCID == cidStr {
			return true, nil
		}
	}
	return false, nil
}
