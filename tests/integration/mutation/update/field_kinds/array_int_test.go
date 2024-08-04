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

func TestMutationUpdate_WithArrayOfIntsToNil(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with integer array, replace with nil",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteIntegers": null
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteIntegers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfIntsToEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with integer array, replace with empty",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteIntegers": []
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteIntegers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []int64{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfIntsToSameSizePositiveValues(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with integer array, replace with same size, positive values",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [8, 5, 3, 2, 1]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteIntegers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []int64{8, 5, 3, 2, 1},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfIntsToSameSizeMixedValues(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with integer array, replace with same size, positive to mixed values",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [-1, 2, -3, 5, -8]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteIntegers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []int64{-1, 2, -3, 5, -8},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfIntsToSmallerSizePositiveValues(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with integer array, replace with smaller size, positive values",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteIntegers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []int64{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfIntsToLargerSizePositiveValues(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with integer array, replace with larger size, positive values",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [1, 2, 3, 5, 8, 13, 21]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteIntegers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []int64{1, 2, 3, 5, 8, 13, 21},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
