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
	"fmt"
	"testing"

	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
)

func TestMutationParse_Create_Simple(t *testing.T) {
	var query = (`
	mutation {
		create_Book(data: "{\"a\": 1}") {
			_key
		}
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	fmt.Println(doc)
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	createMutation, ok := q.Mutations[0].Selections[0].(*Mutation)
	assert.True(t, ok)
	assert.NotNil(t, createMutation)
	assert.Equal(t, "create_Book", createMutation.Name)
	assert.Equal(t, CreateObjects, createMutation.Type)
	assert.Equal(t, "Book", createMutation.Schema)
	// assert.Equal(t, map[string]interface{}{
	// 	"a": float64(1), // json numbers are always floats
	// }, createMutation.Data.Object)
	assert.Len(t, createMutation.Fields, 1)
}

func TestMutationParse_Create_Error_Missing_Data(t *testing.T) {
	var query = (`
	mutation {
		create_Book(data: "") {
			_key
		}
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	t.Log(doc)
	assert.NoError(t, err)

	_, err = ParseQuery(doc)
	assert.Error(t, err)

	// createMutation, ok := q.Mutations[0].Selections[0].(*Mutation)
	// assert.True(t, ok)
	// assert.NotNil(t, createMutation)
	// assert.Equal(t, "create_Book", createMutation.Name)
	// assert.Equal(t, CreateObjects, createMutation.Type)
	// assert.Equal(t, "Book", createMutation.Schema)
	// assert.Equal(t, map[string]interface{}{
	// 	"a": float64(1), // json numbers are always floats
	// }, createMutation.Data.Object)
	// assert.Len(t, createMutation.Fields, 1)
}

func TestMutationParse_Error_Invalid_Mutation(t *testing.T) {
	var query = (`
	mutation {
		dostuff_Book(data: "") {
			_key
		}
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	fmt.Println(doc)
	assert.NoError(t, err)

	_, err = ParseQuery(doc)
	assert.Error(t, err)

	// createMutation, ok := q.Mutations[0].Selections[0].(*Mutation)
	// assert.True(t, ok)
	// assert.NotNil(t, createMutation)
	// assert.Equal(t, "create_Book", createMutation.Name)
	// assert.Equal(t, CreateObjects, createMutation.Type)
	// assert.Equal(t, "Book", createMutation.Schema)
	// assert.Equal(t, map[string]interface{}{
	// 	"a": float64(1), // json numbers are always floats
	// }, createMutation.Data.Object)
	// assert.Len(t, createMutation.Fields, 1)
}

func TestMutationParse_Update_Simple_Object(t *testing.T) {
	var query = (`
	mutation {
		update_Book(data: "{\"a\": 1}") {
			_key
		}
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	fmt.Println(doc)
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	mut, ok := q.Mutations[0].Selections[0].(*Mutation)
	assert.True(t, ok)
	assert.NotNil(t, mut)
	assert.Equal(t, "update_Book", mut.Name)
	assert.Equal(t, UpdateObjects, mut.Type)
	assert.Equal(t, "Book", mut.Schema)
	// assert.Equal(t, map[string]interface{}{
	// 	"a": float64(1), // json numbers are always floats
	// }, mut.Data.Object)
	assert.Len(t, mut.Fields, 1)
}

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
	fmt.Println(doc)
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
	// assert.Equal(t, []interface{}{
	// 	map[string]interface{}{
	// 		"a": float64(1), // json numbers are always floats
	// 	},
	// }, mut.Data.Array)
	// assert.Len(t, mut.Fields, 1)
}

func TestMutationParse_Update_Filter(t *testing.T) {
	var query = (`
	mutation {
		update_Book(filter: {rating: {_gt: 4.5}}, data: "{\"a\": 1}") {
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
	assert.Equal(t, "update_Book", mut.Name)
	assert.Equal(t, UpdateObjects, mut.Type)
	assert.Equal(t, "Book", mut.Schema)
	// assert.Equal(t, map[string]interface{}{
	// 	"a": float64(1), // json numbers are always floats
	// }, mut.Data.Object)
	assert.Len(t, mut.Fields, 1)
	assert.NotNil(t, mut.Filter)
	assert.Equal(t, map[string]interface{}{
		"rating": map[string]interface{}{
			"$gt": 4.5,
		},
	}, mut.Filter.Conditions)
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
	fmt.Println(doc)
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	mut, ok := q.Mutations[0].Selections[0].(*Mutation)
	assert.True(t, ok)
	assert.NotNil(t, mut)
	assert.Equal(t, "update_my_book", mut.Name)
	assert.Equal(t, UpdateObjects, mut.Type)
	assert.Equal(t, "my_book", mut.Schema)
	// assert.Equal(t, map[string]interface{}{
	// 	"a": float64(1), // json numbers are always floats
	// }, mut.Data.Object)
	assert.Len(t, mut.Fields, 1)
}

func TestMutationParse_Delete_Simple(t *testing.T) {
	var query = (`
	mutation {
		delete_Book(data: "{\"a\": 1}") {
			_key
		}
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	fmt.Println(doc)
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	mut, ok := q.Mutations[0].Selections[0].(*Mutation)
	assert.True(t, ok)
	assert.NotNil(t, mut)
	assert.Equal(t, "delete_Book", mut.Name)
	assert.Equal(t, DeleteObjects, mut.Type)
	assert.Equal(t, "Book", mut.Schema)
	// assert.Equal(t, map[string]interface{}{
	// 	"a": float64(1), // json numbers are always floats
	// }, mut.Data.Object)
	assert.Len(t, mut.Fields, 1)
}
