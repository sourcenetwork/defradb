// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestParseIndexOnStruct(t *testing.T) {
	cases := []indexTestCase{
		{
			description: "Index with a single field",
			sdl:         `type user @index(includes: [{field: "name"}]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "",
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: false,
				},
			},
		},
		{
			description: "Index with a name",
			sdl:         `type user @index(name: "userIndex", includes: [{field: "name"}]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "userIndex",
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
				},
			},
		},
		{
			description: "Unique index",
			sdl:         `type user @index(includes: [{field: "name"}], unique: true) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: true,
				},
			},
		},
		{
			description: "Index explicitly not unique",
			sdl:         `type user @index(includes: [{field: "name"}], unique: false) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: false,
				},
			},
		},
		{
			description: "Index with explicit ascending field",
			sdl:         `type user @index(includes: [{field: "name", direction: ASC}]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"}},
				},
			},
		},
		{
			description: "Index with descending field",
			sdl:         `type user @index(includes: [{field: "name", direction: DESC}]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: true}},
				},
			},
		},
		{
			description: "Index with 2 fields",
			sdl:         `type user @index(includes: [{field: "name"}, {field: "age"}]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
						{Name: "age"},
					},
				},
			},
		},
		{
			description: "Index with 2 fields and 2 directions",
			sdl:         `type user @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
						{Name: "age", Descending: true},
					},
				},
			},
		},
	}

	for _, test := range cases {
		parseIndexAndTest(t, test)
	}
}

func TestParseInvalidIndexOnStruct(t *testing.T) {
	cases := []invalidIndexTestCase{
		{
			description: "missing 'includes' argument",
			sdl:         `type user @index(name: "userIndex", unique: true) {}`,
			expectedErr: errIndexMissingFields,
		},
		{
			description: "unknown argument",
			sdl:         `type user @index(unknown: "something", includes: [{field: "name"}]) {}`,
			expectedErr: `Unknown argument "unknown" on directive "@index".`,
		},
		{
			description: "invalid index name type",
			sdl:         `type user @index(name: 1, includes: [{field: "name"}]) {}`,
			expectedErr: `Argument "name" has invalid value 1`,
		},
		{
			description: "index name starts with a number",
			sdl:         `type user @index(name: "1_user_name", includes: [{field: "name"}]) {}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "index with empty name",
			sdl:         `type user @index(name: "", includes: [{field: "name"}]) {}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "index name with spaces",
			sdl:         `type user @index(name: "user name", includes: [{field: "name"}]) {}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "index name with special symbols",
			sdl:         `type user @index(name: "user!name", includes: [{field: "name"}]) {}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "invalid 'unique' value type",
			sdl:         `type user @index(includes: [{field: "name"}], unique: "true") {}`,
			expectedErr: `Argument "unique" has invalid value "true"`,
		},
		{
			description: "invalid 'includes' value type (not a list)",
			sdl:         `type user @index(includes: "name") {}`,
			expectedErr: `Argument "includes" has invalid value "name"`,
		},
		{
			description: "invalid 'includes' value type (not an object list)",
			sdl:         `type user @index(includes: [1]) {}`,
			expectedErr: `Argument "includes" has invalid value [1]`,
		},
	}

	for _, test := range cases {
		parseInvalidIndexAndTest(t, test)
	}
}

func TestParseIndexOnField(t *testing.T) {
	cases := []indexTestCase{
		{
			description: "field index",
			sdl: `type user {
				name: String @index
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "",
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: false,
				},
			},
		},
		{
			description: "field index with name",
			sdl: `type user {
				name: String @index(name: "nameIndex")
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "nameIndex",
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: false,
				},
			},
		},
		{
			description: "unique field index",
			sdl: `type user {
				name: String @index(unique: true)
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: true,
				},
			},
		},
		{
			description: "field index explicitly not unique",
			sdl: `type user {
				name: String @index(unique: false)
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: false,
				},
			},
		},
		{
			description: "field index in ASC order",
			sdl: `type user {
				name: String @index(direction: ASC)
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name"},
					},
					Unique: false,
				},
			},
		},
		{
			description: "field index in DESC order",
			sdl: `type user {
				name: String @index(direction: DESC)
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: true},
					},
					Unique: false,
				},
			},
		},
		{
			description: "composite field index with implicit include and implicit ordering",
			sdl: `type user {
				name: String @index(direction: DESC, includes: [{field: "age"}])
				age: Int
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: true},
						{Name: "age", Descending: true},
					},
					Unique: false,
				},
			},
		},
		{
			description: "composite field index with implicit include and explicit ordering",
			sdl: `type user {
				name: String @index(direction: DESC, includes: [{field: "age", direction: ASC}])
				age: Int
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: true},
						{Name: "age", Descending: false},
					},
					Unique: false,
				},
			},
		},
		{
			description: "composite field index with explicit includes",
			sdl: `type user {
				name: String @index(includes: [{field: "age"}, {field: "name"}])
				age: Int
			}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "",
					Fields: []client.IndexedFieldDescription{
						{Name: "age", Descending: false},
						{Name: "name", Descending: false},
					},
					Unique: false,
				},
			},
		},
	}

	for _, test := range cases {
		parseIndexAndTest(t, test)
	}
}

func TestParseInvalidIndexOnField(t *testing.T) {
	cases := []invalidIndexTestCase{
		{
			description: "forbidden 'field' argument",
			sdl: `type user {
				name: String @index(field: "name") 
			}`,
			expectedErr: `Unknown argument "field" on directive "@index`,
		},
		{
			description: "invalid field index name type",
			sdl: `type user {
				name: String @index(name: 1) 
			}`,
			expectedErr: `Argument "name" has invalid value 1`,
		},
		{
			description: "field index name starts with a number",
			sdl: `type user {
				name: String @index(name: "1_user_name") 
			}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "field index with empty name",
			sdl: `type user {
				name: String @index(name: "") 
			}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "field index name with spaces",
			sdl: `type user {
				name: String @index(name: "user name") 
			}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "field index name with special symbols",
			sdl: `type user {
				name: String @index(name: "user!name") 
			}`,
			expectedErr: errIndexInvalidName,
		},
		{
			description: "invalid 'unique' value type",
			sdl: `type user {
				name: String @index(unique: "true") 
			}`,
			expectedErr: `Argument "unique" has invalid value "true"`,
		},
	}

	for _, test := range cases {
		parseInvalidIndexAndTest(t, test)
	}
}

func parseIndexAndTest(t *testing.T, testCase indexTestCase) {
	schemaManager, err := NewSchemaManager()
	require.NoError(t, err)

	cols, err := schemaManager.ParseSDL(testCase.sdl)
	require.NoError(t, err, testCase.description)

	require.Equal(t, len(cols), 1, testCase.description)
	require.Equal(t, len(cols[0].Description.Indexes), len(testCase.targetDescriptions), testCase.description)

	for i, d := range cols[0].Description.Indexes {
		assert.Equal(t, testCase.targetDescriptions[i], d, testCase.description)
	}
}

func parseInvalidIndexAndTest(t *testing.T, testCase invalidIndexTestCase) {
	schemaManager, err := NewSchemaManager()
	require.NoError(t, err)

	_, err = schemaManager.ParseSDL(testCase.sdl)
	assert.ErrorContains(t, err, testCase.expectedErr, testCase.description)
}

type indexTestCase struct {
	description        string
	sdl                string
	targetDescriptions []client.IndexDescription
}

type invalidIndexTestCase struct {
	description string
	sdl         string
	expectedErr string
}
