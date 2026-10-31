// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func vectorMaintenanceIndex(maintenance client.VectorIndexMaintenance) []client.IndexDescription {
	return []client.IndexDescription{
		{
			Name:   "User_vector_ASC",
			ID:     1,
			Fields: []client.IndexedFieldDescription{{Name: "vector"}},
			Kind:   client.IndexKindVector,
			KindDescription: &client.VectorIndexDescription{
				Fields:      []string{"vector"},
				Algorithm:   client.VectorAlgorithmHNSW,
				Metric:      client.DistanceMetricCosine,
				Dimensions:  3,
				Maintenance: maintenance,
				HNSW: &client.HNSWParams{
					M:              client.DefaultHNSWM,
					EfConstruction: client.DefaultHNSWEfConstruction,
					EfSearch:       client.DefaultHNSWEfSearch,
				},
			},
		},
	}
}

func TestVectorIndex_CreateWithDeferredMaintenanceInSDL_IsStored(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, maintenance: DEFERRED, hnsw: {metric: COSINE}})
				}`,
			},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: vectorMaintenanceIndex(client.VectorIndexMaintenanceDeferred),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Maintenance is left empty when not set, which means SYNC.
func TestVectorIndex_CreateWithoutMaintenance_DefaultsToSync(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
				}`,
			},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: vectorMaintenanceIndex(""),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestVectorIndex_NewIndexWithDeferredMaintenance_IsStored(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!]
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "vector",
				Vector: &client.VectorIndexDescription{
					Metric:      client.DistanceMetricCosine,
					Dimensions:  3,
					Maintenance: client.VectorIndexMaintenanceDeferred,
				},
			},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: vectorMaintenanceIndex(client.VectorIndexMaintenanceDeferred),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestVectorIndex_NewIndexWithUnknownMaintenance_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!]
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "vector",
				Vector: &client.VectorIndexDescription{
					Metric:      client.DistanceMetricCosine,
					Dimensions:  3,
					Maintenance: "LAZY",
				},
				ExpectedError: "vector index with unknown maintenance",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// An explicit SYNC is stored the same way as leaving it out, so there is one stored form.
func TestVectorIndex_CreateWithExplicitSyncMaintenance_IsStoredAsDefault(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, maintenance: SYNC, hnsw: {metric: COSINE}})
				}`,
			},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: vectorMaintenanceIndex(""),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestVectorIndex_CreateWithUnknownMaintenanceInSDL_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, maintenance: LAZY, hnsw: {metric: COSINE}})
				}`,
				ExpectedError: `Expected type "VectorIndexMaintenance", found LAZY`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
