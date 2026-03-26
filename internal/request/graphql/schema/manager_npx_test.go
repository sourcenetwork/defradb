// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build npx
// +build npx

package schema

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/stretchr/testify/require"
)

func TestWriteSDL_Simple_Succeeds(t *testing.T) {
	sdl := `
	type User {
		name: String
		age: Int
		verified: Boolean
		points: Int @crdt(type: pncounter)
	}`

	runWriteSDLTest(t, sdl, "schema.simple.gen.graphql")
}

func TestWriteSDL_RelatedOne_Succeeds(t *testing.T) {
	sdl := `
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
	}`

	runWriteSDLTest(t, sdl, "schema.relatedone.gen.graphql")
}

func TestWriteSDL_RelatedMany_Succeeds(t *testing.T) {
	sdl := `
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
	}`

	runWriteSDLTest(t, sdl, "schema.relatedmany.gen.graphql")
}

func runWriteSDLTest(t *testing.T, sdl string, fixtureName string) {
	ctx := t.Context()
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)

	cols, err := manager.ParseSDL(sdl)
	require.NoError(t, err)

	collections := make([]client.CollectionVersion, len(cols))
	for i, c := range cols {
		collections[i] = c.Definition
	}

	cache := description.NewCollectionCache()
	cache.AddAll(collections)
	ctx = description.ContextWithCollectionCache(ctx, cache)

	_, err = manager.Generator.Generate(ctx, collections)
	require.NoError(t, err)

	outBuf := bytes.NewBuffer(nil)
	err = manager.WriteSDL(outBuf)
	require.NoError(t, err)
	sdlResult := outBuf.Bytes()

	testFixturePath := getFullFixturePath(fixtureName)

	cmd := exec.Command("npx", "-y", "@graphql-inspector/cli@6.0.7",
		"diff",
		string(sdlResult),
		testFixturePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(output))
	}
	require.NoError(t, err)
	require.Contains(t, string(output), "No changes detected")
}

func getFullFixturePath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(thisFile)
	return filepath.Join(baseDir, "testfixtures", name)
}
