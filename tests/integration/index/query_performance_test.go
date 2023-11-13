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
	"testing"

	"github.com/sourcenetwork/defradb/tests/gen"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryPerformance_Simple(t *testing.T) {
	const benchReps = 10
	const numDocs = 500

	getOptions := func(col string) []gen.Option {
		return []gen.Option{
			gen.WithTypeDemand(col, numDocs),
			gen.WithFieldRange(col, "age", 0, 99),
		}
	}

	test1 := testUtils.TestCase{
		Actions: []any{
			testUtils.GenerateDocsForSchema{
				Schema: `
					type User {
						name:   String
						age:    Int 
						email:  String
					}`,
				AutoGenOptions: getOptions("User"),
			},
			testUtils.GenerateDocsForSchema{
				Schema: `
					type IndexedUser {
						name:   String
						age:    Int @index
						email:  String
					}`,
				AutoGenOptions: getOptions("IndexedUser"),
			},
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
				Factor:       5,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test1)
}
