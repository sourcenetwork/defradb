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
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexNew_WithPNCounterField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			&action.NewIndex{
				CollectionID:  0,
				IndexName:     "points_index",
				FieldName:     "points",
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithPCounterField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			&action.NewIndex{
				CollectionID:  0,
				IndexName:     "points_index",
				FieldName:     "points",
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pcounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithPNCounterFieldViaDirective_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						points: Int @crdt(type: pncounter) @index
					}
				`,
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithPCounterFieldViaDirective_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						points: Int @crdt(type: pcounter) @index
					}
				`,
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pcounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithPNCounterFloatField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						score: Float @crdt(type: pncounter)
					}
				`,
			},
			&action.NewIndex{
				CollectionID:  0,
				IndexName:     "score_index",
				FieldName:     "score",
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("score", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithLWWField_ShouldSucceed(t *testing.T) {
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
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "age_index",
				FieldName:    "age",
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 30
				}`,
			},
			&action.Request{
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(30),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithCompositeIndexIncludingPNCounter_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			&action.NewIndex{
				CollectionID:  0,
				IndexName:     "composite_index",
				Fields:        []client.IndexedFieldDescription{{Name: "name"}, {Name: "points"}},
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithUniqueIndexOnPNCounter_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			&action.NewIndex{
				CollectionID:  0,
				IndexName:     "unique_points_index",
				FieldName:     "points",
				Unique:        true,
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_WithCollectionLevelIndexOnPNCounter_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(includes: [{field: "points"}]) {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
