// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexAdd_WithPNCounterField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			&action.AddIndex{
				CollectionID:  0,
				IndexName:     "points_index",
				FieldName:     "points",
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexAdd_WithPCounterField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			&action.AddIndex{
				CollectionID:  0,
				IndexName:     "points_index",
				FieldName:     "points",
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pcounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexAdd_WithPNCounterFieldViaDirective_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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

func TestIndexAdd_WithPCounterFieldViaDirective_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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

func TestIndexAdd_WithPNCounterFloatField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						score: Float @crdt(type: pncounter)
					}
				`,
			},
			&action.AddIndex{
				CollectionID:  0,
				IndexName:     "score_index",
				FieldName:     "score",
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("score", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexAdd_WithLWWField_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddIndex{
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

func TestIndexAdd_WithCompositeIndexIncludingPNCounter_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			&action.AddIndex{
				CollectionID:  0,
				IndexName:     "composite_index",
				Fields:        []client.IndexedFieldDescription{{Name: "name"}, {Name: "points"}},
				ExpectedError: db.NewErrCannotIndexAccumulatedCRDTField("points", "pncounter").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexAdd_WithUniqueIndexOnPNCounter_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			&action.AddIndex{
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

func TestIndexAdd_WithCollectionLevelIndexOnPNCounter_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
