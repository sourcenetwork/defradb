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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// structRequestClientTypes restricts a test to the clients that send the index request as a struct.
// The CLI and C clients translate it into flags, collapsing Ordered and the deprecated Unique into a
// single --unique, so they cannot tell the two spellings apart.
var structRequestClientTypes = immutable.Some([]state.ClientType{
	state.GoClientType,
	state.HTTPClientType,
})

func TestAddUniqueIndex_IfFieldValuesAreNotUnique_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	21
					}`,
			},
			// The unique backfill fails on the duplicate values. The failure is not returned from
			// NewIndex; the action waits for the failed state, asserted via ListIndexes below.
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "age",
				Unique:       true,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_age_ASC",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "age"},
						},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"User_age_ASC": {
						Status: client.ErroredActionStatus,
						Action: client.BackfillIndexAction,
						Reason: "can not index",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_UponAddingDocWithExistingFieldValue_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @index(unique: true, name: "age_unique_index")
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.Request{
				Request: `query {
					User(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_IfFieldValuesAreUnique_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int 
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	22
					}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "age_unique_index",
				FieldName:    "age",
				Unique:       true,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_WithMultipleNilFields_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Keenan"
					}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "age_unique_index",
				FieldName:    "age",
				Unique:       true,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_AddingDocWithNilValue_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John"
					}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_UponAddingDocWithExistingNilValue_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Keenan"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy"
					}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueQueryWithIndex_UponAddingDocWithSameDateTime_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						birthday: DateTime @index(unique: true)
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_SetThroughOrderedConfig_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: structRequestClientTypes,
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Ordered:      &client.OrderedIndexDescription{Unique: true},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:            "User_name_ASC",
						ID:              1,
						Fields:          []client.IndexedFieldDescription{{Name: "name"}},
						Unique:          true,
						Kind:            client.IndexKindOrdered,
						KindDescription: &client.OrderedIndexDescription{Unique: true},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_BothSpellingsAgree_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: structRequestClientTypes,
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Unique:       true,
				Ordered:      &client.OrderedIndexDescription{Unique: true},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_SpellingsDisagree_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: structRequestClientTypes,
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
				}`,
			},
			&action.NewIndex{
				CollectionID:  0,
				FieldName:     "name",
				Unique:        true,
				Ordered:       &client.OrderedIndexDescription{Unique: false},
				ExpectedError: "index request sets both the deprecated unique field",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_BothOrderedAndVectorConfig_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: structRequestClientTypes,
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					vector: [Float32!]
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "vector",
				Ordered:      &client.OrderedIndexDescription{},
				Vector: &client.VectorIndexDescription{
					Metric:     client.DistanceMetricCosine,
					Dimensions: 3,
					HNSW:       &client.HNSWParams{},
				},
				ExpectedError: "index request has more than one kind config",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
