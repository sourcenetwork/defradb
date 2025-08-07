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

	"github.com/sourcenetwork/defradb/tests/action"
	"github.com/sourcenetwork/defradb/tests/gen"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestQueryPerformance_Simple(t *testing.T) {
	const benchReps = 10

	getOptions := func(col string) []gen.Option {
		return []gen.Option{
			gen.WithTypeDemand(col, 500),
			gen.WithFieldRange(col, "age", 0, 99),
		}
	}

	test1 := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name:   String
						age:    Int 
						email:  String
					}`,
			},
			&action.AddSchema{
				Schema: `
					type IndexedUser {
						name:   String
						age:    Int @index
						email:  String
					}`,
			},
			testUtils.GenerateDocs{
				Options: append(getOptions("User"), getOptions("IndexedUser")...),
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
				FocusClients: []state.ClientType{testUtils.GoClientType},
				Factor:       2,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test1)
}

func TestQueryPerformance_WithFloat32(t *testing.T) {
	const benchReps = 10

	getOptions := func(col string) []gen.Option {
		return []gen.Option{
			gen.WithTypeDemand(col, 500),
			gen.WithFieldRange(col, "points", float32(0), float32(99)),
		}
	}

	test1 := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name:   String
						points:    Float32 
						email:  String
					}`,
			},
			&action.AddSchema{
				Schema: `
					type IndexedUser {
						name:   String
						points:    Float32 @index
						email:  String
					}`,
			},
			testUtils.GenerateDocs{
				Options: append(getOptions("User"), getOptions("IndexedUser")...),
			},
			testUtils.Benchmark{
				Reps: benchReps,
				BaseCase: testUtils.Request{Request: `
					query {
						User(filter: {points: {_eq: 33}}) {
							name
							points
							email
						}
					}`,
				},
				OptimizedCase: testUtils.Request{Request: `
					query {
						IndexedUser(filter: {points: {_eq: 33}}) {
							name
							points
							email
						}
					}`,
				},
				FocusClients: []state.ClientType{testUtils.GoClientType},
				Factor:       2,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test1)
}
