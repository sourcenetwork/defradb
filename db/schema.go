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

	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/core"
)

// LoadSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, query types, etc.
func (db *db) AddSchema(ctx context.Context, schemaString string) error {
	schema, err := db.parser.AddSchema(ctx, schemaString)
	if err != nil {
		return err
	}

	for _, desc := range schema.Descriptions {
		if _, err := db.CreateCollection(ctx, desc); err != nil {
			return err
		}
	}

	return db.saveSchema(ctx, schema)
}

func (db *db) loadSchema(ctx context.Context) error {
	var sdl string
	q := dsq.Query{
		Prefix: "/schema",
	}
	res, err := db.systemstore().Query(ctx, q)
	if err != nil {
		return err
	}

	for kv := range res.Next() {
		buf := kv.Value[:]
		sdl += "\n" + string(buf)
	}

	_, err = db.parser.AddSchema(ctx, sdl)
	return err
}

func (db *db) saveSchema(ctx context.Context, schema *core.Schema) error {
	// save each type individually
	for _, def := range schema.Definitions {
		key := core.NewSchemaKey(def.Name)
		if err := db.systemstore().Put(ctx, key.ToDS(), def.Body); err != nil {
			return err
		}
	}
	return nil
}
