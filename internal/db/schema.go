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
)

// addSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *DB) addSchema(
	ctx context.Context,
	schemaString string,
) ([]client.CollectionVersion, error) {
	newDefinitions, err := db.parser.ParseSDL(ctx, schemaString)
	if err != nil {
		return nil, err
	}

	returnDefinitions, err := db.createCollections(ctx, newDefinitions)
	if err != nil {
		return nil, err
	}

	returnDescriptions := make([]client.CollectionVersion, len(returnDefinitions))
	for i, def := range returnDefinitions {
		returnDescriptions[i] = def.Version
	}

	err = db.loadSchema(ctx)
	if err != nil {
		return nil, err
	}

	return returnDescriptions, nil
}

func (db *DB) loadSchema(ctx context.Context) error {
	definitions, err := db.getAllActiveDefinitions(ctx)
	if err != nil {
		return err
	}

	return db.parser.SetSchema(ctx, definitions)
}
