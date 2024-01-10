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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/description"
)

func (db *db) addView(
	ctx context.Context,
	txn datastore.Txn,
	inputQuery string,
	sdl string,
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

	req, errs := db.parser.Parse(ast)
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
		newDefinitions[i].Description.BaseQuery = baseQuery
	}

	returnDescriptions := make([]client.CollectionDefinition, len(newDefinitions))
	for i, definition := range newDefinitions {
		if definition.Description.Name == "" {
			schema, err := description.CreateSchemaVersion(ctx, txn, definition.Schema)
			if err != nil {
				return nil, err
			}
			returnDescriptions[i] = client.CollectionDefinition{
				// `Collection` is left as default for embedded types
				Schema: schema,
			}
		} else {
			col, err := db.createCollection(ctx, txn, definition)
			if err != nil {
				return nil, err
			}
			returnDescriptions[i] = col.Definition()
		}
	}

	err = db.loadSchema(ctx, txn)
	if err != nil {
		return nil, err
	}

	return returnDescriptions, nil
}
