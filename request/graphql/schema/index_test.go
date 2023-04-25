// Copyright 2022 Democratized Data Foundation
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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
)

func TestSingleIndex(t *testing.T) {
	cases := []indexTestCase{
		{
			description: "Index with a single field",
			sdl:         `type user @index(fields: ["name"]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending},
					},
					IsUnique: false,
				},
			},
		},
		{
			description: "Index with a name",
			sdl:         `type user @index(name: "userIndex", fields: ["name"]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Name: "userIndex",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending},
					},
				},
			},
		},
		{
			description: "Unique index",
			sdl:         `type user @index(fields: ["name"], unique: true) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending},
					},
					IsUnique: true,
				},
			},
		},
		{
			description: "Index explicitly not unique",
			sdl:         `type user @index(fields: ["name"], unique: false) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending},
					},
					IsUnique: false,
				},
			},
		},
		{
			description: "Index with explicit ascending field",
			sdl:         `type user @index(fields: ["name"], directions: [ASC]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending}},
				},
			},
		},
		{
			description: "Index with descending field",
			sdl:         `type user @index(fields: ["name"], directions: [DESC]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Descending}},
				},
			},
		},
		{
			description: "Index with 2 fields",
			sdl:         `type user @index(fields: ["name", "age"]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending},
						{Name: "age", Direction: client.Ascending},
					},
				},
			},
		},
		{
			description: "Index with 2 fields and 2 directions",
			sdl:         `type user @index(fields: ["name", "age"], directions: [ASC, DESC]) {}`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending},
						{Name: "age", Direction: client.Descending},
					},
				},
			},
		},
	}

	for _, test := range cases {
		parseIndexAndTest(t, test)
	}
}

func TestInvalidIndexSyntax(t *testing.T) {
	cases := []invalidIndexTestCase{
		{
			description: "missing 'fields' argument",
			sdl:         `type user @index(name: "userIndex", unique: true) {}`,
			expectedErr: errIndexMissingFields,
		},
		{
			description: "unknown argument",
			sdl:         `type user @index(unknown: "something", fields: ["name"]) {}`,
			expectedErr: errIndexUnknownArgument,
		},
		{
			description: "invalid index name type",
			sdl:         `type user @index(name: 1, fields: ["name"]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "index name starts with a number",
			sdl:         `type user @index(name: "1_user_name", fields: ["name"]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "index with empty name",
			sdl:         `type user @index(name: "", fields: ["name"]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "index name with spaces",
			sdl:         `type user @index(name: "user name", fields: ["name"]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "index name with special symbols",
			sdl:         `type user @index(name: "user!name", fields: ["name"]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "invalid 'unique' value type",
			sdl:         `type user @index(fields: ["name"], unique: "true") {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "invalid 'fields' value type (not a list)",
			sdl:         `type user @index(fields: "name") {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "invalid 'fields' value type (not a string list)",
			sdl:         `type user @index(fields: [1]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "invalid 'directions' value type (not a list)",
			sdl:         `type user @index(fields: ["name"], directions: "ASC") {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "invalid 'directions' value type (not a string list)",
			sdl:         `type user @index(fields: ["name"], directions: [1]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "invalid 'directions' value type (invalid element value)",
			sdl:         `type user @index(fields: ["name"], directions: ["direction"]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "fewer directions than fields",
			sdl:         `type user @index(fields: ["name", "age"], directions: [ASC]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "more directions than fields",
			sdl:         `type user @index(fields: ["name"], directions: [ASC, DESC]) {}`,
			expectedErr: errIndexInvalidArgument,
		},
	}

	for _, test := range cases {
		parseInvalidIndexAndTest(t, test)
	}
}

func parseIndexAndTest(t *testing.T, testCase indexTestCase) {
	ctx := context.Background()

	_, indexes, err := FromString(ctx, testCase.sdl)
	assert.NoError(t, err, testCase.description)
	assert.Equal(t, len(indexes), len(testCase.targetDescriptions), testCase.description)

	for i, d := range indexes {
		assert.Equal(t, testCase.targetDescriptions[i], d, testCase.description)
	}
}

func parseInvalidIndexAndTest(t *testing.T, testCase invalidIndexTestCase) {
	ctx := context.Background()

	_, _, err := FromString(ctx, testCase.sdl)
	assert.EqualError(t, err, testCase.expectedErr, testCase.description)
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
