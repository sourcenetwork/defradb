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

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
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
