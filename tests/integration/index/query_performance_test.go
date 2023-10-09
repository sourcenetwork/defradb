// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/stretchr/testify/assert"
)

func generateDocs(count int) []any {
	result := make([]any, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, testUtils.CreateDoc{
			CollectionID: 0,
			Doc: fmt.Sprintf(`{
				"name": "name-%d",
				"age":  %d,
				"email":  "email%d@gmail.com"
			}`, i, i%100, i),
		})
	}
	return result
}

func TestQueryPerformance_WithNonIndexedFields_ShouldFetchAllOfThem(t *testing.T) {
	const benchReps = 1
	const numDocs = 1000

	docs := generateDocs(numDocs)

	const req = `query {
		User(filter: {age: {_eq: 33}}) {
			name
			age
			email
			verify
		}
	}`

	var benchResRegular testUtils.BenchmarkResult
	test1 := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{Schema: `
				type User {
					name: String
					age: Int
					email: String
					verify: Boolean
				}
			`},
			docs,
			testUtils.Benchmark{
				Reps:   benchReps,
				Action: testUtils.Request{Request: req},
				Result: &benchResRegular,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test1)

	var benchResIndexed testUtils.BenchmarkResult
	test2 := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{Schema: `
				type User {
					name: String 
					age: Int @index
					email: String
					verify: Boolean
				} 
			`},
			docs,
			testUtils.Benchmark{
				Reps:   benchReps,
				Action: testUtils.Request{Request: req},
				Result: &benchResIndexed,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test2)

	for dbt, regularVal := range benchResRegular.ElapsedTime {
		indexedVal := benchResIndexed.ElapsedTime[dbt]
		regularMs := regularVal.Microseconds()
		indexedMs := indexedVal.Microseconds()
		const factor = 10
		assert.Greater(t, regularMs/factor, indexedMs,
			"Indexed query should be at least %d time as fast as regular (db: %s). Indexed: %d, regular: %d (μs)",
			factor, dbt, indexedMs, regularMs)
	}
}
