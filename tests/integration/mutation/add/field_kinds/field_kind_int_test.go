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

package field_kinds

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAddFieldKinds_WithInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						age: Int
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"age": int64(42),
				},
			},
			&action.Request{
				Request: `query {
					User {
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"age": int64(42),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableInt_Nil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						age: Int
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"age": nil,
				},
			},
			&action.Request{
				Request: `query {
					User {
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"age": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						age: Int!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"age": int64(42),
				},
			},
			&action.Request{
				Request: `query {
					User {
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"age": int64(42),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableIntArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						numbers: [Int]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"numbers": []int64{1, 2, 3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						numbers
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"numbers": []immutable.Option[int64]{
								immutable.Some(int64(1)),
								immutable.Some(int64(2)),
								immutable.Some(int64(3)),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableIntArray_WithNilElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						numbers: [Int]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"numbers": []any{int64(1), nil, int64(3)},
				},
			},
			&action.Request{
				Request: `query {
					User {
						numbers
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"numbers": []immutable.Option[int64]{
								immutable.Some(int64(1)),
								immutable.None[int64](),
								immutable.Some(int64(3)),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableIntArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						numbers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"numbers": []int64{1, 2, 3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						numbers
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"numbers": []int64{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableIntArray_FromAnySlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						numbers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"numbers": []any{int64(1), int64(2), int64(3)},
				},
			},
			&action.Request{
				Request: `query {
					User {
						numbers
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"numbers": []int64{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
