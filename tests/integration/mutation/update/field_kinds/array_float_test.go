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

func TestMutationUpdate_WithArrayOfFloatsToNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			&action.AddDoc{
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
			&action.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteFloats": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			&action.AddDoc{
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
			&action.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteFloats": []float64{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToSameSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			&action.AddDoc{
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
			&action.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteFloats": []float64{3.1425, -0.00000000001, 1000000},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToSmallerSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			&action.AddDoc{
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
			&action.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteFloats": []float64{3.14},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfFloatsToLargerSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteFloats: [Float!]
					}
				`,
			},
			&action.AddDoc{
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
			&action.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteFloats": []float64{3.1425, 0.00000000001, -10, 6.626070},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
