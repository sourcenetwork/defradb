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
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func vectorModeSDL(maintenance string) *action.AddCollection {
	return &action.AddCollection{
		SDL: fmt.Sprintf(`type User {
			name: String
			vector: [Float32!] @index(vector: {dimensions: 3, %shnsw: {metric: COSINE}})
		}`, maintenance),
	}
}

const vectorTop2Query = `query {
	User(order: {_alias: {sim: DESC}}, limit: 2){
		name
		sim: SIMILARITY(vector: {vector: [1, 0, 0]})
	}
}`

func cosX(v ...float64) any {
	return testUtils.CosineSimilarity(v, []float64{1, 0, 0})
}

// Eight documents, a mix of writes, updates and a delete, give the expected top three.
func TestVectorIndexDeferred_AddUpdateDelete_SearchIsCorrect(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			vectorModeSDL("maintenance: DEFERRED, "),
			&action.AddDoc{DocMap: map[string]any{"name": "d0", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "d1", "vector": []float32{0.9, 0.1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "d2", "vector": []float32{0.8, 0.3, 0.1}}},
			&action.AddDoc{DocMap: map[string]any{"name": "d3", "vector": []float32{0.5, 0.5, 0.5}}},
			&action.AddDoc{DocMap: map[string]any{"name": "d4", "vector": []float32{0, 1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "d5", "vector": []float32{0, 0, 1}}},
			&action.AddDoc{DocMap: map[string]any{"name": "d6", "vector": []float32{0.2, 0.2, 0.9}}},
			&action.AddDoc{DocMap: map[string]any{"name": "d7", "vector": []float32{-1, 0, 0}}},
			// d4 moves next to the query; d0, the best match, is deleted.
			&action.UpdateDoc{DocID: 4, Doc: `{"vector": [0.95, 0.05, 0]}`},
			&action.Request{
				Request: `mutation { delete_User(docID: ["{{.DocID0_0}}"]) { _docID } }`,
				Results: map[string]any{
					"delete_User": []map[string]any{{"_docID": "{{.DocID0_0}}"}},
				},
			},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 3){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "d4", "sim": cosX(0.95, 0.05, 0)},
						{"name": "d1", "sim": cosX(0.9, 0.1, 0)},
						{"name": "d2", "sim": cosX(0.8, 0.3, 0.1)},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// Several writers committing at once, then a search. Concurrent writers to a SYNC graph index conflict
// on shared graph nodes (the reason for deferred maintenance), so only the deferred index is used.
// Results do not depend on whether the background writer has drained, so no wait is needed.
func TestVectorIndexDeferred_ConcurrentWriters_AllSearchable(t *testing.T) {
	children := make([]action.Action, 0, 8)
	for i := range 8 {
		children = append(children, &action.AddDoc{
			DoNotWaitForEvent: true,
			DocMap: map[string]any{
				"name":   fmt.Sprintf("d%d", i),
				"vector": []float32{1, float32(i) / 10, 0},
			},
		})
	}
	test := testUtils.TestCase{
		Actions: []any{
			vectorModeSDL("maintenance: DEFERRED, "),
			&action.Parallel{Children: children},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 8){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "d0", "sim": cosX(1, 0, 0)},
						{"name": "d1", "sim": cosX(1, 0.1, 0)},
						{"name": "d2", "sim": cosX(1, 0.2, 0)},
						{"name": "d3", "sim": cosX(1, 0.3, 0)},
						{"name": "d4", "sim": cosX(1, 0.4, 0)},
						{"name": "d5", "sim": cosX(1, 0.5, 0)},
						{"name": "d6", "sim": cosX(1, 0.6, 0)},
						{"name": "d7", "sim": cosX(1, 0.7, 0)},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
