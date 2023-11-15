// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaParser_ParseGenConfig(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		want   configsMap
	}{
		{
			name: "string values",
			schema: `
				type User {
					name: String # pattern: "some pattern"
				}`,
			want: configsMap{
				"User": {
					"name": {
						props: map[string]any{
							"pattern": "some pattern",
						},
					},
				},
			},
		},
		{
			name: "bool values",
			schema: `
				type User {
					verified: Boolean # default: true
				}`,
			want: configsMap{
				"User": {
					"verified": {
						props: map[string]any{
							"default": true,
						},
					},
				},
			},
		},
		{
			name: "int values",
			schema: `
				type User {
					age: Int # min: 4, max: 10
				}`,
			want: configsMap{
				"User": {
					"age": {
						props: map[string]any{
							"min": 4,
							"max": 10,
						},
					},
				},
			},
		},
		{
			name: "float values",
			schema: `
				type User {
					rating: Float # min: 1.1, max: 5.5
				}`,
			want: configsMap{
				"User": {
					"rating": {
						props: map[string]any{
							"min": 1.1,
							"max": 5.5,
						},
					},
				},
			},
		},
		{
			name: "labels",
			schema: `
				type User {
					name: String # unique, indexed
				}`,
			want: configsMap{
				"User": {
					"name": {
						labels: []string{"unique", "indexed"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConfig(tt.schema)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSchemaParser_IfCanNotParse_ReturnError(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name: "missing value",
			schema: `
				type User {
					name: String # pattern:
				}`,
		},
		{
			name: "missing prop name",
			schema: `
				type User {
					name: String # : 3
				}`,
		},
		{
			name: "no coma between props",
			schema: `
				type User {
					verified: Boolean # label1 label2
				}`,
		},
		{
			name: "invalid value",
			schema: `
				type User {
					age: Int # min: 4 5
				}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseConfig(tt.schema)
			assert.ErrorIs(t, err, NewErrFailedToParse(""))
		})
	}
}
