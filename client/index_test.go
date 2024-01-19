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
		desc     CollectionDescription
		field    string
		expected []IndexDescription
	}{
		{
			name: "no indexes",
			desc: CollectionDescription{
				Indexes: []IndexDescription{},
			},
			field:    "test",
			expected: []IndexDescription{},
		},
		{
			name: "single index on field",
			desc: CollectionDescription{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "test", Direction: Ascending},
						},
					},
				},
			},
			field: "test",
			expected: []IndexDescription{
				{
					Name: "index1",
					Fields: []IndexedFieldDescription{
						{Name: "test", Direction: Ascending},
					},
				},
			},
		},
		{
			name: "multiple indexes on field",
			desc: CollectionDescription{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "test", Direction: Ascending},
						},
					},
					{
						Name: "index2",
						Fields: []IndexedFieldDescription{
							{Name: "test", Direction: Descending},
						},
					},
				},
			},
			field: "test",
			expected: []IndexDescription{
				{
					Name: "index1",
					Fields: []IndexedFieldDescription{
						{Name: "test", Direction: Ascending},
					},
				},
				{
					Name: "index2",
					Fields: []IndexedFieldDescription{
						{Name: "test", Direction: Descending},
					},
				},
			},
		},
		{
			name: "no indexes on field",
			desc: CollectionDescription{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "other", Direction: Ascending},
						},
					},
				},
			},
			field:    "test",
			expected: []IndexDescription{},
		},
		{
			name: "second field in composite index",
			desc: CollectionDescription{
				Indexes: []IndexDescription{
					{
						Name: "index1",
						Fields: []IndexedFieldDescription{
							{Name: "other", Direction: Ascending},
							{Name: "test", Direction: Ascending},
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
			actual := tt.desc.GetIndexesOnField(tt.field)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
