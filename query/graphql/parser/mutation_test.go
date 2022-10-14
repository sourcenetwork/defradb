// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package parser

import (
	"testing"

	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
)

func TestMutationParse_Update_Simple_Array(t *testing.T) {
	var query = (`
	mutation {
		update_Book(data: "[{\"a\": 1}]") {
			_key
		}
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	assert.NoError(t, err)

	_, err = ParseQuery(doc)
	assert.NoError(t, err)

	// mut, ok := q.Mutations[0].Selections[0].(*Mutation)
	// assert.True(t, ok)
	// assert.NotNil(t, mut)
	// assert.Equal(t, "update_Book", mut.Name)
	// assert.Equal(t, UpdateObjects, mut.Type)
	// assert.Equal(t, "Book", mut.Schema)
	// assert.Empty(t, mut.Data)
	// assert.Equal(t, []any{
	// 	map[string]any{
	// 		"a": float64(1), // json numbers are always floats
	// 	},
	// }, mut.Data.Array)
	// assert.Len(t, mut.Fields, 1)
}

func TestMutationParse_Update_Simple_UnderscoreName(t *testing.T) {
	var query = (`
	mutation {
		update_my_book(data: "{\"a\": 1}") {
			_key
		}
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	mut, ok := q.Mutations[0].Selections[0].(*Mutation)
	assert.True(t, ok)
	assert.NotNil(t, mut)
	assert.Equal(t, "update_my_book", mut.Name)
	assert.Equal(t, UpdateObjects, mut.Type)
	assert.Equal(t, "my_book", mut.Schema)
	// assert.Equal(t, map[string]any{
	// 	"a": float64(1), // json numbers are always floats
	// }, mut.Data.Object)
	assert.Len(t, mut.Fields, 1)
}
