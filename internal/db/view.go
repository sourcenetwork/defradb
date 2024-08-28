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
	"errors"
	"fmt"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/planner"
)

func (db *db) addView(
	ctx context.Context,
	inputQuery string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	// Wrap the given query as part of the GQL query object - this simplifies the syntax for users
	// and ensures that we can't be given mutations.  In the future this line should disappear along
	// with the all calls to the parser appart from `ParseSDL` when we implement the DQL stuff.
	query := fmt.Sprintf(`query { %s }`, inputQuery)

	newDefinitions, err := db.parser.ParseSDL(ctx, sdl)
	if err != nil {
		return nil, err
	}

	ast, err := db.parser.BuildRequestAST(query)
	if err != nil {
		return nil, err
	}

	req, errs := db.parser.Parse(ast, &client.GQLOptions{})
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

	for i := range newDefinitions {
		source := client.QuerySource{
			Query:     *baseQuery,
			Transform: transform,
		}
		newDefinitions[i].Description.Sources = append(newDefinitions[i].Description.Sources, &source)
	}

	returnDescriptions, err := db.createCollections(ctx, newDefinitions)
	if err != nil {
		return nil, err
	}

	for _, definition := range returnDescriptions {
		for _, source := range definition.Description.QuerySources() {
			if source.Transform.HasValue() {
				err = db.LensRegistry().SetMigration(ctx, definition.Description.ID, source.Transform.Value())
				if err != nil {
					return nil, err
				}
			}
		}
	}

	err = db.loadSchema(ctx)
	if err != nil {
		return nil, err
	}

	return returnDescriptions, nil
}

func (db *db) refreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	// For now, we only support user-cache management of views, not all collections
	cols, err := db.getViews(ctx, opts)
	if err != nil {
		return err
	}

	for _, col := range cols {
		if !col.Description.IsMaterialized {
			// We only care about materialized views here, so skip any that aren't
			continue
		}

		// Clearing and then constructing is a bit inefficient, but it should do for now.
		// Long term we probably want to update inline as much as possible to avoid unnessecarily
		// moving/adding/deleting keys in storage
		err := db.clearViewCache(ctx, col)
		if err != nil {
			return err
		}

		err = db.buildViewCache(ctx, col)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) getViews(ctx context.Context, opts client.CollectionFetchOptions) ([]client.CollectionDefinition, error) {
	cols, err := db.getCollections(ctx, opts)
	if err != nil {
		return nil, err
	}

	var views []client.CollectionDefinition
	for _, col := range cols {
		if querySrcs := col.Description().QuerySources(); len(querySrcs) == 0 {
			continue
		}

		views = append(views, col.Definition())
	}

	return views, nil
}

func (db *db) buildViewCache(ctx context.Context, col client.CollectionDefinition) (err error) {
	txn := mustGetContextTxn(ctx)
	identity := GetContextIdentity(ctx)

	p := planner.New(ctx, identity, db.acp, db, txn)

	// temporarily disable the cache in order to query without using it
	col.Description.IsMaterialized = false
	col.Description, err = description.SaveCollection(ctx, txn, col.Description)
	if err != nil {
		return err
	}
	defer func() {
		var defErr error
		col.Description.IsMaterialized = true
		col.Description, defErr = description.SaveCollection(ctx, txn, col.Description)
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

	hasNext, err := source.Next()
	if err != nil {
		return err
	}

	// View items are currently keyed by their index, starting at 1.
	// The order in which results are returned must be consistent with the results of the
	// underlying query/transform.
	var itemID uint
	for itemID = 1; hasNext; itemID++ {
		doc := source.Value()

		serializedItem, err := core.MarshalViewItem(doc)
		if err != nil {
			return err
		}

		itemKey := core.NewViewCacheKey(col.Description.RootID, itemID)
		err = txn.Datastore().Put(ctx, itemKey.ToDS(), serializedItem)
		if err != nil {
			return err
		}

		hasNext, err = source.Next()
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) clearViewCache(ctx context.Context, col client.CollectionDefinition) error {
	txn := mustGetContextTxn(ctx)
	prefix := core.NewViewCacheColPrefix(col.Description.RootID)

	q, err := txn.Datastore().Query(ctx, query.Query{
		Prefix:   prefix.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return err
	}

	for res := range q.Next() {
		if res.Error != nil {
			return errors.Join(res.Error, q.Close())
		}

		err = txn.Datastore().Delete(ctx, ds.NewKey(res.Key))
		if err != nil {
			return errors.Join(err, q.Close())
		}
	}

	return q.Close()
}

func (db *db) generateMaximalSelectFromCollection(
	ctx context.Context,
	col client.CollectionDefinition,
	fieldName immutable.Option[string],
	typesHit map[string]struct{},
) (*request.Select, error) {
	// `__-` is an impossible field name prefix, so we can safely concat using it as a separator without risk
	// of collision.
	identifier := col.GetName() + "__-" + fieldName.Value()
	if _, ok := typesHit[identifier]; ok {
		// If this identifier is already in the set, the schema must be circular and we should return
		return nil, nil
	}
	typesHit[identifier] = struct{}{}

	childRequests := []request.Selection{}
	for _, field := range col.GetFields() {
		if field.IsRelation() && field.Kind.IsObject() {
			relatedCol, _, err := client.GetDefinitionFromStore(ctx, db, col, field.Kind)
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
				// innerSelect may be nil if a circular relationship is defined in the schema and we have already
				// added this field
				childRequests = append(childRequests, innerSelect)
			}
		}
	}

	var name string
	if fieldName.HasValue() {
		name = fieldName.Value()
	} else {
		name = col.GetName()
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
