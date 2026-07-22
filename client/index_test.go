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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIndexDescription_NewUniqueSecondary_RoundTrips(t *testing.T) {
	original := IndexDescription{
		Name:      "some_index",
		ID:        1,
		Fields:    []IndexedFieldDescription{{Name: "name"}},
		Secondary: &SecondaryIndexDescription{Unique: true},
	}

	bytes, err := json.Marshal(original)
	require.NoError(t, err)

	var actual IndexDescription
	err = json.Unmarshal(bytes, &actual)
	require.NoError(t, err)

	require.NotNil(t, actual.Secondary)
	assert.True(t, actual.Secondary.Unique)
	assert.Nil(t, actual.Vector)
	assert.Equal(t, IndexKindSecondary, actual.Kind())
	assert.Equal(t, original, actual)
}

func TestIndexDescription_LegacyTopLevelUniqueTrue_MigratesToSecondary(t *testing.T) {
	legacyJSON := `{"Name":"x","ID":1,"Fields":[{"Name":"age"}],"Unique":true}`

	var actual IndexDescription
	err := json.Unmarshal([]byte(legacyJSON), &actual)
	require.NoError(t, err)

	require.NotNil(t, actual.Secondary)
	assert.True(t, actual.Secondary.Unique)
	assert.Nil(t, actual.Vector)
	assert.Equal(t, IndexKindSecondary, actual.Kind())
}

func TestIndexDescription_LegacyTopLevelUniqueFalse_MigratesToSecondary(t *testing.T) {
	legacyJSON := `{"Unique":false}`

	var actual IndexDescription
	err := json.Unmarshal([]byte(legacyJSON), &actual)
	require.NoError(t, err)

	require.NotNil(t, actual.Secondary)
	assert.False(t, actual.Secondary.Unique)
	assert.Nil(t, actual.Vector)
}

func TestIndexDescription_NoUniqueOrKindFields_DefaultsToSecondary(t *testing.T) {
	legacyJSON := `{"Name":"x","ID":1,"Fields":[{"Name":"age"}]}`

	var actual IndexDescription
	err := json.Unmarshal([]byte(legacyJSON), &actual)
	require.NoError(t, err)

	require.NotNil(t, actual.Secondary)
	assert.False(t, actual.Secondary.Unique)
	assert.Nil(t, actual.Vector)
	assert.Equal(t, IndexKindSecondary, actual.Kind())
}

func TestIndexDescription_VectorDescriptor_RoundTrips(t *testing.T) {
	original := IndexDescription{
		Name:   "some_vector_index",
		ID:     2,
		Fields: []IndexedFieldDescription{{Name: "embedding"}},
		Vector: &VectorIndexDescription{
			Algorithm:  VectorAlgorithmHNSW,
			Metric:     DistanceMetricCosine,
			Dimensions: 128,
			HNSW: &HNSWParams{
				M:              16,
				EfConstruction: 200,
				EfSearch:       50,
			},
		},
	}

	bytes, err := json.Marshal(original)
	require.NoError(t, err)

	var actual IndexDescription
	err = json.Unmarshal(bytes, &actual)
	require.NoError(t, err)

	require.NotNil(t, actual.Vector)
	assert.Nil(t, actual.Secondary)
	assert.Equal(t, IndexKindVector, actual.Kind())
	assert.Equal(t, original, actual)
}
