// Copyright 2020 Source Inc.
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
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/schema"

	"github.com/graphql-go/graphql/language/ast"
)

// LoadSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, query types, etc.
func (db *DB) AddSchema(schema string) error {
	// @todo: create collection after generating query types
	types, astdoc, err := db.schema.Generator.FromSDL(schema)
	if err != nil {
		return err
	}
	colDesc, err := db.schema.Generator.CreateDescriptions(types)
	if err != nil {
		return err
	}
	for _, desc := range colDesc {
		// fmt.Println(desc)
		if _, err := db.CreateCollection(desc); err != nil {
			return err
		}
	}

	return db.saveSchema(astdoc)
}

func (db *DB) loadSchema() error {
	var sdl string
	q := dsq.Query{
		Prefix: "/schema",
	}
	res, err := db.systemstore.Query(q)
	if err != nil {
		return err
	}

	for kv := range res.Next() {
		buf := kv.Value[:]
		sdl += "\n" + string(buf)
	}

	_, _, err = db.schema.Generator.FromSDL(sdl)
	return err
}

func (db *DB) saveSchema(astdoc *ast.Document) error {
	// save each type individually
	for _, def := range astdoc.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			body := defType.Loc.Source.Body[defType.Loc.Start:defType.Loc.End]
			key := core.NewSchemaKey(defType.Name.Value)
			if err := db.systemstore.Put(key.ToDS(), body); err != nil {
				return err
			}
		}
	}
	return nil
}

// func (db *DB) LoadSchemaIfNotExists(schema string) error { return nil }

func (db *DB) SchemaManager() *schema.SchemaManager {
	return db.schema
}
