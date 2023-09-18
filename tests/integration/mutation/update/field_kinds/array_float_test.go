// Copyright 2023 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithArrayOfFloatsToNil(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with float array, replace with nil",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteFloats": null
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: []map[string]any{
					{
						"favouriteFloats": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with float array, replace with empty",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteFloats": []
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: []map[string]any{
					{
						"favouriteFloats": []float64{},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToSameSize(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with float array, replace with same size",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteFloats": [3.1425, -0.00000000001, 1000000]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: []map[string]any{
					{
						"favouriteFloats": []float64{3.1425, -0.00000000001, 1000000},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToSmallerSize(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with float array, replace with smaller size",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteFloats": [3.14]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: []map[string]any{
					{
						"favouriteFloats": []float64{3.14},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToLargerSize(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with float array, replace with larger size",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteFloats": [3.1425, 0.00000000001, -10, 6.626070]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: []map[string]any{
					{
						"favouriteFloats": []float64{3.1425, 0.00000000001, -10, 6.626070},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
