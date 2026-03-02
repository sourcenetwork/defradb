// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field_kinds

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"

	"github.com/sourcenetwork/immutable"
)

func TestMutationAdd_WithJSONFieldGivenObjectValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": {"tree": "maple", "age": 250}
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": map[string]any{
								"tree": "maple",
								"age":  float64(250),
							},
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithJSONFieldGivenListOfScalarsValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": ["maple", 250]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": []any{"maple", float64(250)},
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithJSONFieldGivenListOfObjectsValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": [
						{"tree": "maple"}, 
						{"tree": "oak"}
					]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": []any{
								map[string]any{"tree": "maple"},
								map[string]any{"tree": "oak"},
							},
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithJSONFieldGivenIntValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": 250
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": float64(250),
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithJSONFieldGivenStringValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": "hello"
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": "hello",
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithJSONFieldGivenBooleanValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": true
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": true,
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithJSONFieldGivenNullValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": nil,
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test confirms that our JSON value encoding is determinstic.
func TestMutationAdd_WithDuplicateJSONField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// Save will not produce an error on duplicate
			// because it will just update the previous doc
			state.GQLRequestMutationType,
			state.CollectionNamedMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": {"one": 1, "two": 2, "three": [0, 1, 2]}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John", 
					"custom": {"three": [0, 1, 2], "two": 2, "one": 1}
				}`,
				ExpectedError: `a document with the given ID already exists`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
