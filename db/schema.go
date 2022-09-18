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

	"github.com/graphql-go/graphql/language/ast"
	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/core"
)

// LoadSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, query types, etc.
func (db *db) AddSchema(ctx context.Context, schema string) error {
	// @todo: create collection after generating query types
	types, astdoc, err := db.schema.Generator.FromSDL(ctx, schema)
	if err != nil {
		return err
	}
	colDesc, err := db.schema.Generator.CreateDescriptions(types)
	if err != nil {
		return err
	}
	for _, desc := range colDesc {
		if _, err := db.CreateCollection(ctx, desc); err != nil {
			return err
		}
	}

	return db.saveSchema(ctx, astdoc)
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

	_, _, err = db.schema.Generator.FromSDL(ctx, sdl)
	return err
}

func (db *db) saveSchema(ctx context.Context, astdoc *ast.Document) error {
	// save each type individually
	for _, def := range astdoc.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			body := defType.Loc.Source.Body[defType.Loc.Start:defType.Loc.End]
			key := core.NewSchemaKey(defType.Name.Value)
			if err := db.systemstore().Put(ctx, key.ToDS(), body); err != nil {
				return err
			}
		}
	}
	return nil
}
