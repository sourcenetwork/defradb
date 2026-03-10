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

func TestMutationUpdate_WithArrayOfBooleansToNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						likedIndexes: [Boolean!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": [true, true, false, true]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"likedIndexes": null
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							likedIndexes
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"likedIndexes": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfBooleansToEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						likedIndexes: [Boolean!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": [true, true, false, true]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"likedIndexes": []
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							likedIndexes
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"likedIndexes": []bool{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfBooleansToSameSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						likedIndexes: [Boolean!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": [true, true, false, true]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"likedIndexes": [true, false, true, false]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							likedIndexes
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"likedIndexes": []bool{true, false, true, false},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfBooleansToSmallerSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						likedIndexes: [Boolean!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": [true, true, false, true]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"likedIndexes": [false, true]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							likedIndexes
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"likedIndexes": []bool{false, true},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfBooleansToLargerSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						likedIndexes: [Boolean!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": [true, true, false, true]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"likedIndexes": [true, false, true, false, true, true]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							likedIndexes
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"likedIndexes": []bool{true, false, true, false, true, true},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
