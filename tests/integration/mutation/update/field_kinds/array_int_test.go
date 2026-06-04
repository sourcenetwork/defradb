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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithArrayOfIntsToNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteIntegers": null
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteIntegers": []
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [8, 5, 3, 2, 1]
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [-1, 2, -3, 5, -8]
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteIntegers: [Int!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3, 5, 8]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [1, 2, 3, 5, 8, 13, 21]
				}`,
			},
			&action.Request{
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
