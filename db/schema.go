// Copyright 2022 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/request/graphql/schema"
)

// AddSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *db) AddSchema(ctx context.Context, schemaString string) error {
	collectionDescriptions, err := schema.FromString(ctx, schemaString)
	if err != nil {
		return err
	}

	err = db.parser.AddSchema(ctx, collectionDescriptions)
	if err != nil {
		return err
	}

	for _, desc := range collectionDescriptions {
		if _, err := db.CreateCollection(ctx, desc); err != nil {
			return err
		}
	}

	return nil
}

func (db *db) loadSchema(ctx context.Context) error {
	collections, err := db.GetAllCollections(ctx)
	if err != nil {
		return err
	}

	descriptions := make([]client.CollectionDescription, len(collections))
	for i, collection := range collections {
		descriptions[i] = collection.Description()
	}

	return db.parser.AddSchema(ctx, descriptions)
}
