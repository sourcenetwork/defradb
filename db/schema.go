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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/request/graphql/schema"
)

// AddSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *db) AddSchema(ctx context.Context, schemaString string) error {
	err := db.parser.AddSchema(ctx, schemaString)
	if err != nil {
		return err
	}

	collectionDescriptions, schemaDefinitions, err := createDescriptions(ctx, schemaString)
	if err != nil {
		return err
	}

	for _, desc := range collectionDescriptions {
		if _, err := db.CreateCollection(ctx, desc); err != nil {
			return err
		}
	}

	return db.saveSchema(ctx, schemaDefinitions)
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

	return db.parser.AddSchema(ctx, sdl)
}

func (db *db) saveSchema(ctx context.Context, schemaDefinitions []core.SchemaDefinition) error {
	// save each type individually
	for _, def := range schemaDefinitions {
		key := core.NewSchemaKey(def.Name)
		if err := db.systemstore().Put(ctx, key.ToDS(), def.Body); err != nil {
			return err
		}
	}
	return nil
}

func createDescriptions(
	ctx context.Context,
	schemaString string,
) ([]client.CollectionDescription, []core.SchemaDefinition, error) {
	// We should not be using this package at all here, and should not be doing most of what this package does
	// Rework: https://github.com/sourcenetwork/defradb/issues/923
	schemaManager, err := schema.NewSchemaManager()
	if err != nil {
		return nil, nil, err
	}

	types, astdoc, err := schemaManager.Generator.FromSDL(ctx, schemaString)
	if err != nil {
		return nil, nil, err
	}

	colDesc, err := schemaManager.Generator.CreateDescriptions(types)
	if err != nil {
		return nil, nil, err
	}

	definitions := make([]core.SchemaDefinition, len(astdoc.Definitions))
	for i, astDefinition := range astdoc.Definitions {
		objDef, isObjDef := astDefinition.(*ast.ObjectDefinition)
		if !isObjDef {
			continue
		}

		definitions[i] = core.SchemaDefinition{
			Name: objDef.Name.Value,
			Body: objDef.Loc.Source.Body[objDef.Loc.Start:objDef.Loc.End],
		}
	}

	return colDesc, definitions, nil
}
