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
)

func generateDocsForCollection(colIndex, count int) []any {
	result := make([]any, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, testUtils.CreateDoc{
			CollectionID: colIndex,
			Doc: fmt.Sprintf(`{
				"name": "name-%d",
				"age":  %d,
				"email":  "email%d@gmail.com"
			}`, i, i%100, i),
		})
	}
	return result
}

func TestQueryPerformance_Simple(t *testing.T) {
	const benchReps = 10
	const numDocs = 500

	test1 := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{Schema: `
				type User {
					name:   String
					age:    Int
					email:  String
				}
			`},
			testUtils.SchemaUpdate{
				Schema: `
				    type IndexedUsers {
					    name:   String
					    age:    Int @index
					    email:  String
				    }
			    `,
			},
			generateDocsForCollection(0, numDocs),
			generateDocsForCollection(1, numDocs),
			testUtils.Benchmark{
				Reps: benchReps,
				BaseCase: testUtils.Request{Request: `
					query {
						User(filter: {age: {_eq: 33}}) {
							name
							age
							email
						}
					}`,
				},
				OptimizedCase: testUtils.Request{Request: `
					query {
						IndexedUser(filter: {age: {_eq: 33}}) {
							name
							age
							email
						}
					}`,
				},
				FocusClients: []testUtils.ClientType{testUtils.GoClientType},
				Factor:       10,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test1)
}
