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
			&action.AddCollection{
				SDL: `
					type User {
						name:   String
						age:    Int 
						email:  String
					}`,
			},
			&action.AddCollection{
				SDL: `
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
				BaseCase: &action.Request{Request: `
					query {
						User(filter: {age: {_eq: 33}}) {
							name
							age
							email
						}
					}`,
				},
				OptimizedCase: &action.Request{Request: `
					query {
						IndexedUser(filter: {age: {_eq: 33}}) {
							name
							age
							email
						}
					}`,
				},
				FocusClients: []state.ClientType{state.GoClientType},
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
			&action.AddCollection{
				SDL: `
					type User {
						name:   String
						points:    Float32 
						email:  String
					}`,
			},
			&action.AddCollection{
				SDL: `
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
				BaseCase: &action.Request{Request: `
					query {
						User(filter: {points: {_eq: 33}}) {
							name
							points
							email
						}
					}`,
				},
				OptimizedCase: &action.Request{Request: `
					query {
						IndexedUser(filter: {points: {_eq: 33}}) {
							name
							points
							email
						}
					}`,
				},
				FocusClients: []state.ClientType{state.GoClientType},
				Factor:       2,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test1)
}
