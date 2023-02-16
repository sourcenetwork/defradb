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
	"github.com/sourcenetwork/defradb/datastore"
)

// AddSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *db) AddSchema(ctx context.Context, schemaString string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	collectionDescriptions, err := db.parser.ParseSDL(ctx, schemaString)
	if err != nil {
		return err
	}

	err = db.parser.AddSchema(ctx, collectionDescriptions)
	if err != nil {
		return err
	}

	for _, desc := range collectionDescriptions {
		if _, err := db.CreateCollectionTxn(ctx, txn, desc); err != nil {
			return err
		}
	}

	return txn.Commit(ctx)
}

func (db *db) loadSchema(ctx context.Context, txn datastore.Txn) error {
	collections, err := db.GetAllCollectionsTxn(ctx, txn)
	if err != nil {
		return err
	}

	descriptions := make([]client.CollectionDescription, len(collections))
	for i, collection := range collections {
		descriptions[i] = collection.Description()
	}

	return db.parser.AddSchema(ctx, descriptions)
}
