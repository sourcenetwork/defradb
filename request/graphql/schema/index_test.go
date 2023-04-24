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
			sdl: `
			type user @index(fields: ["name"]) {
				name: String
			}
			`,
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
			sdl: `
			type user @index(name: "userIndex", fields: ["name"]) {
				name: String
			}
			`,
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
			sdl: `
			type user @index(fields: ["name"], unique: true) {
				name: String
			}
			`,
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
			sdl: `
			type user @index(fields: ["name"], unique: false) {
				name: String
			}
			`,
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
			sdl: `
			type user @index(fields: ["name"], directions: [ASC]) {
				name: String
			}
			`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Ascending}},
				},
			},
		},
		{
			description: "Index with descending field",
			sdl: `
			type user @index(fields: ["name"], directions: [DESC]) {
				name: String
			}
			`,
			targetDescriptions: []client.IndexDescription{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Direction: client.Descending}},
				},
			},
		},
	}

	for _, test := range cases {
		parseIndexAndTest(t, test)
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

type indexTestCase struct {
	description        string
	sdl                string
	targetDescriptions []client.IndexDescription
}
