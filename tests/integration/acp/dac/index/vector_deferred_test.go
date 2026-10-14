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

package test_acp_dac_index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const vectorACPQuery = `query {
	Users(order: {_alias: {sim: DESC}}, limit: 3){
		name
		sim: SIMILARITY(vector: {vector: [1, 0, 0]})
	}
}`

func vectorACPSim(v ...float64) any {
	return testUtils.CosineSimilarity(v, []float64{1, 0, 0})
}

// A document owned by one identity is not returned to another by a similarity query on a deferred
// vector index, whether its entry is still pending or already applied by the background writer.
func TestACPWithVectorIndex_Deferred_PrivateDocsAreNotLeaked(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						vector: [Float32!] @index(vector: {dimensions: 3, maintenance: DEFERRED, hnsw: {metric: COSINE}})
					}
				`,
			},
			// Public document, far from the query.
			&action.AddDoc{DocMap: map[string]any{"name": "public", "vector": []float32{0, 1, 0}}},
			// Owned by identity 1, the nearest to the query.
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap:   map[string]any{"name": "private", "vector": []float32{1, 0, 0}},
			},
			// Owned by identity 2, second nearest.
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(2),
				DocMap:   map[string]any{"name": "other", "vector": []float32{0.9, 0.4, 0}},
			},
			// The owner moves its document; the stale pending vector must not leak it either.
			&action.UpdateDoc{
				Identity: testUtils.ClientIdentity(1),
				DocID:    1,
				Doc:      `{"vector": [0.99, 0.01, 0]}`,
			},
			&action.Request{
				Request: vectorACPQuery,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "public", "sim": vectorACPSim(0, 1, 0)},
					},
				},
			},
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request:  vectorACPQuery,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "other", "sim": vectorACPSim(0.9, 0.4, 0)},
						{"name": "public", "sim": vectorACPSim(0, 1, 0)},
					},
				},
			},
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request:  vectorACPQuery,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "private", "sim": vectorACPSim(0.99, 0.01, 0)},
						{"name": "public", "sim": vectorACPSim(0, 1, 0)},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
