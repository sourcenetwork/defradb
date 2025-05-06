// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectIndexesOnField(t *testing.T) {
	tests := []struct {
		name     string
		version  CollectionVersion
		field    string
		expected []IndexDescription
	}{
		{
			name: "no indexes",
			version: CollectionVersion{
				Indexes: []IndexDescription{},
			},
			field:    "test",
			expected: []IndexDescription{},
		},
		{
			name: "single index on field",
			version: CollectionVersion{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "test"},
						},
					},
				},
			},
			field: "test",
			expected: []IndexDescription{
				{
					Name: "index1",
					Fields: []IndexedFieldDescription{
						{Name: "test"},
					},
				},
			},
		},
		{
			name: "multiple indexes on field",
			version: CollectionVersion{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "test"},
						},
					},
					{
						Name: "index2",
						Fields: []IndexedFieldDescription{
							{Name: "test", Descending: true},
						},
					},
				},
			},
			field: "test",
			expected: []IndexDescription{
				{
					Name: "index1",
					Fields: []IndexedFieldDescription{
						{Name: "test"},
					},
				},
				{
					Name: "index2",
					Fields: []IndexedFieldDescription{
						{Name: "test", Descending: true},
					},
				},
			},
		},
		{
			name: "no indexes on field",
			version: CollectionVersion{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "other"},
						},
					},
				},
			},
			field:    "test",
			expected: []IndexDescription{},
		},
		{
			name: "second field in composite index",
			version: CollectionVersion{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "other"},
							{Name: "test"},
						},
					},
				},
			},
			field:    "test",
			expected: []IndexDescription{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.version.GetIndexesOnField(tt.field)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
