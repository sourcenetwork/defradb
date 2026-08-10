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

// A descriptor persisted before the Kind interface existed has only a top-level Unique. It must
// still load, reading that Unique into both the ordered kind and the compat Unique field.
func TestIndexDescription_LegacyUnique_LoadsAsOrdered(t *testing.T) {
	json1 := `{"Name":"x","ID":1,"Fields":[{"Name":"age"}],"Unique":true}`

	var actual IndexDescription
	err := json.Unmarshal([]byte(json1), &actual)
	require.NoError(t, err)

	assert.False(t, actual.IsVector())
	assert.True(t, actual.GetUnique())
	assert.True(t, actual.Unique) // compat field synced from the resolved kind
	assert.Equal(t, &OrderedIndexDescription{Unique: true}, actual.Kind)
}

// An embedded caller that predates Kind builds the struct with only the top-level Unique. It must
// behave correctly (GetUnique) and marshal a top-level Unique so an old reader still sees it.
func TestIndexDescription_CompatUniqueOnly_Works(t *testing.T) {
	desc := IndexDescription{
		Name:   "x",
		ID:     1,
		Fields: []IndexedFieldDescription{{Name: "age"}},
		Unique: true, // no Kind set, as old callers do
	}
	assert.True(t, desc.GetUnique())
	assert.False(t, desc.IsVector())

	bytes, err := json.Marshal(desc)
	require.NoError(t, err)

	// An old reader (no Kind field) must still see the uniqueness at the top level.
	var oldReader struct{ Unique bool }
	require.NoError(t, json.Unmarshal(bytes, &oldReader))
	assert.True(t, oldReader.Unique)
}

func TestIndexDescription_OrderedDescriptor_RoundTrips(t *testing.T) {
	original := IndexDescription{
		Name:   "some_unique_index",
		ID:     1,
		Fields: []IndexedFieldDescription{{Name: "age"}},
		// A fully-formed ordered descriptor sets both the compat Unique and the Kind, in sync.
		Unique: true,
		Kind:   &OrderedIndexDescription{Unique: true},
	}

	bytes, err := json.Marshal(original)
	require.NoError(t, err)

	var actual IndexDescription
	err = json.Unmarshal(bytes, &actual)
	require.NoError(t, err)

	assert.False(t, actual.IsVector())
	assert.True(t, actual.GetUnique())
	assert.Equal(t, original, actual)
}

func TestIndexDescription_VectorDescriptor_RoundTrips(t *testing.T) {
	original := IndexDescription{
		Name:   "some_vector_index",
		ID:     2,
		Fields: []IndexedFieldDescription{{Name: "embedding"}},
		Kind: &VectorIndexDescription{
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

	require.True(t, actual.IsVector())
	assert.Equal(t, original, actual)
}
