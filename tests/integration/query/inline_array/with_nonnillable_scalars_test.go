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

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineArrayWithNonNillableBooleans_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { flags: [Boolean!] }`,
			},
			&action.AddDoc{
				Doc: `{"flags": null}`,
			},
			&action.Request{
				Request: `query { Users { flags } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"flags": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableBooleans_NullElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { flags: [Boolean!] }`,
			},
			&action.AddDoc{
				Doc:           `{"flags": [true, null, false]}`,
				ExpectedError: `null`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableBooleans_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { flags: [Boolean!] }`,
			},
			&action.AddDoc{
				Doc: `{"flags": [true, false, true]}`,
			},
			&action.Request{
				Request: `query { Users { flags } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"flags": []bool{true, false, true}},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableInts_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { scores: [Int!] }`,
			},
			&action.AddDoc{
				Doc: `{"scores": null}`,
			},
			&action.Request{
				Request: `query { Users { scores } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"scores": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableInts_NullElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { scores: [Int!] }`,
			},
			&action.AddDoc{
				Doc:           `{"scores": [1, null, 3]}`,
				ExpectedError: `null value provided for non-nillable field`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableInts_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { scores: [Int!] }`,
			},
			&action.AddDoc{
				Doc: `{"scores": [1, 2, 3]}`,
			},
			&action.Request{
				Request: `query { Users { scores } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"scores": []int64{1, 2, 3}},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat64s_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { ratings: [Float64!] }`,
			},
			&action.AddDoc{
				Doc: `{"ratings": null}`,
			},
			&action.Request{
				Request: `query { Users { ratings } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"ratings": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat64s_NullElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { ratings: [Float64!] }`,
			},
			&action.AddDoc{
				Doc:           `{"ratings": [1.1, null, 3.3]}`,
				ExpectedError: `null value provided for non-nillable field`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat64s_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { ratings: [Float64!] }`,
			},
			&action.AddDoc{
				Doc: `{"ratings": [1.1, 2.2, 3.3]}`,
			},
			&action.Request{
				Request: `query { Users { ratings } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"ratings": []float64{1.1, 2.2, 3.3}},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat32s_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weights: [Float32!] }`,
			},
			&action.AddDoc{
				Doc: `{"weights": null}`,
			},
			&action.Request{
				Request: `query { Users { weights } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"weights": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat32s_NullElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weights: [Float32!] }`,
			},
			&action.AddDoc{
				Doc:           `{"weights": [1.1, null, 3.3]}`,
				ExpectedError: `null value provided for non-nillable field`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat32s_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weights: [Float32!] }`,
			},
			&action.AddDoc{
				Doc: `{"weights": [1.1, 2.2, 3.3]}`,
			},
			&action.Request{
				Request: `query { Users { weights } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"weights": []float32{1.1, 2.2, 3.3}},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableStrings_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { tags: [String!] }`,
			},
			&action.AddDoc{
				Doc: `{"tags": null}`,
			},
			&action.Request{
				Request: `query { Users { tags } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"tags": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableStrings_NullElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { tags: [String!] }`,
			},
			&action.AddDoc{
				Doc:           `{"tags": ["one", null, "three"]}`,
				ExpectedError: `null value provided for non-nillable field`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableStrings_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { tags: [String!] }`,
			},
			&action.AddDoc{
				Doc: `{"tags": ["one", "two", "three"]}`,
			},
			&action.Request{
				Request: `query { Users { tags } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"tags": []string{"one", "two", "three"}},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableDateTime_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, createdAt: DateTime! }`,
			},
			&action.AddDoc{
				Doc:           `{"name": "John", "createdAt": null}`,
				ExpectedError: `null value provided for non-nillable field. Name: createdAt`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableDateTime_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, createdAt: DateTime! }`,
			},
			&action.AddDoc{
				Doc: `{"name": "John", "createdAt": "2024-01-01T00:00:00Z"}`,
			},
			&action.Request{
				Request: `query { Users { name } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableBlob_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, avatar: Blob! }`,
			},
			&action.AddDoc{
				Doc:           `{"name": "John", "avatar": null}`,
				ExpectedError: `null value provided for non-nillable field. Name: avatar`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableBlob_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, avatar: Blob! }`,
			},
			&action.AddDoc{
				Doc: `{"name": "John", "avatar": "ff0011"}`,
			},
			&action.Request{
				Request: `query { Users { name, avatar } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "avatar": "ff0011"},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableJSON_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, metadata: JSON! }`,
			},
			&action.AddDoc{
				Doc:           `{"name": "John", "metadata": null}`,
				ExpectedError: `null value provided for non-nillable field. Name: metadata`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableJSON_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, metadata: JSON! }`,
			},
			&action.AddDoc{
				Doc: `{"name": "John", "metadata": {"role": "admin"}}`,
			},
			&action.Request{
				Request: `query { Users { name, metadata } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "metadata": map[string]any{"role": "admin"}},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableBooleans_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { flags: [Boolean!] }`,
			},
			&action.AddDoc{
				Doc: `{}`,
			},
			&action.Request{
				Request: `query { Users { flags } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"flags": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableInts_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { scores: [Int!] }`,
			},
			&action.AddDoc{
				Doc: `{}`,
			},
			&action.Request{
				Request: `query { Users { scores } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"scores": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat64s_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { ratings: [Float64!] }`,
			},
			&action.AddDoc{
				Doc: `{}`,
			},
			&action.Request{
				Request: `query { Users { ratings } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"ratings": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat32s_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { weights: [Float32!] }`,
			},
			&action.AddDoc{
				Doc: `{}`,
			},
			&action.Request{
				Request: `query { Users { weights } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"weights": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableStrings_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { tags: [String!] }`,
			},
			&action.AddDoc{
				Doc: `{}`,
			},
			&action.Request{
				Request: `query { Users { tags } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"tags": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableDateTime_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, createdAt: DateTime! }`,
			},
			&action.AddDoc{
				Doc: `{"name": "John"}`,
			},
			&action.Request{
				Request: `query { Users { name, createdAt } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "createdAt": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableBlob_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, avatar: Blob! }`,
			},
			&action.AddDoc{
				Doc: `{"name": "John"}`,
			},
			&action.Request{
				Request: `query { Users { name, avatar } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "avatar": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryNonNillableJSON_Omitted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, metadata: JSON! }`,
			},
			&action.AddDoc{
				Doc: `{"name": "John"}`,
			},
			&action.Request{
				Request: `query { Users { name, metadata } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "metadata": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
