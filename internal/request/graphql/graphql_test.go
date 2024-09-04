// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const testSchema = `type Book {
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

func TestGenerateSchema(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	collections, err := parser.ParseSDL(context.Background(), testSchema)
	require.NoError(t, err)

	source, err := GenerateSchema(collections)
	require.NoError(t, err)

	fmt.Printf("%s", source)
}
