// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package testfixtures holds the SDL test fixtures for the schema generator along with
// the logic to (re)generate them. The committed *.gen.graphql files in this directory are
// produced by the standalone generator in ./gen (run via `make sdl-fixtures`) and verified
// against the current generator output by the package's npx-tagged test.
package testfixtures

import (
	"bytes"
	"context"
	"path/filepath"
	"runtime"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
)

// Fixture pairs an input SDL with the name of its committed, generated fixture file.
type Fixture struct {
	// Name is the file name of the generated fixture within this directory.
	Name string
	// SDL is the input schema the fixture is generated from.
	SDL string
}

// Fixtures is the set of SDL fixtures shared by the generator and the verification test.
// Keeping a single source for both ensures the committed fixtures and the test cannot drift.
var Fixtures = []Fixture{
	{
		Name: "schema.simple.gen.graphql",
		SDL: `
	type User {
		name: String
		age: Int
		verified: Boolean
		points: Int @crdt(type: pncounter)
	}`,
	},
	{
		Name: "schema.relatedone.gen.graphql",
		SDL: `
	type Book {
		name: String
		rating: Float
		author: Author @primary
	}

	type Author {
		name: String
		age: Int
		verified: Boolean
		published: Book
	}`,
	},
	{
		Name: "schema.relatedmany.gen.graphql",
		SDL: `
	type Book {
		name: String
		rating: Float
		author: Author
	}

	type Author {
		name: String
		age: Int
		verified: Boolean
		published: [Book]
	}`,
	},
}

// GenerateSDL runs the schema generator over the given input SDL and returns the
// resulting normalised SDL output.
func GenerateSDL(ctx context.Context, sdl string) ([]byte, error) {
	manager, err := schema.NewSchemaManager(false)
	if err != nil {
		return nil, err
	}

	cols, err := manager.ParseSDL(sdl)
	if err != nil {
		return nil, err
	}

	collections := make([]client.CollectionVersion, len(cols))
	for i, c := range cols {
		collections[i] = c.Definition
	}

	if _, err := manager.Generator.Generate(ctx, collections); err != nil {
		return nil, err
	}

	out := bytes.NewBuffer(nil)
	if err := manager.WriteSDL(out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// Path returns the absolute path of a fixture file within this directory.
func Path(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), name)
}
