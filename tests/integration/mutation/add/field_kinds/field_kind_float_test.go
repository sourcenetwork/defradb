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

func TestMutationAddFieldKinds_WithFloat(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float64(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float32
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float32(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float64
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float64(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableFloat64_Nil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float64
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": nil,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableFloat32_Nil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float32
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": nil,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableFloat64Array_WithNilElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float64]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []any{1.1, nil, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []immutable.Option[float64]{
								immutable.Some(1.1),
								immutable.None[float64](),
								immutable.Some(3.3),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableFloat32Array_WithNilElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float32]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []any{float32(1.1), nil, float32(3.3)},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []immutable.Option[float32]{
								immutable.Some(float32(1.1)),
								immutable.None[float32](),
								immutable.Some(float32(3.3)),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloat(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": float64(1.2),
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float64(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float32!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": float32(1.2),
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float32(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float64!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": float64(1.2),
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float64(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableFloatArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []float64{1.1, 2.2, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []immutable.Option[float64]{
								immutable.Some(1.1),
								immutable.Some(2.2),
								immutable.Some(3.3),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloatArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []float64{1.1, 2.2, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []float64{1.1, 2.2, 3.3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloatArray_FromAnySlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []any{1.1, 2.2, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []float64{1.1, 2.2, 3.3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableFloat64Array_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float64]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []float64{1.1, 2.2, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []immutable.Option[float64]{
								immutable.Some(1.1),
								immutable.Some(2.2),
								immutable.Some(3.3),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloat64Array_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float64!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []float64{1.1, 2.2, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []float64{1.1, 2.2, 3.3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableFloat32Array_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float32]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []float32{1.1, 2.2, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []immutable.Option[float32]{
								immutable.Some(float32(1.1)),
								immutable.Some(float32(2.2)),
								immutable.Some(float32(3.3)),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloat32Array_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float32!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []float32{1.1, 2.2, 3.3},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []float32{1.1, 2.2, 3.3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableFloat32Array_FromAnySlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: [Float32!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": []any{float32(1.1), float32(2.2), float32(3.3)},
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": []float32{1.1, 2.2, 3.3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
