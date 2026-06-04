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

func TestMutationUpdate_WithArrayOfStringsToNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"preferredStrings": null
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"preferredStrings": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"preferredStrings": []
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"preferredStrings": []string{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToSameSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"preferredStrings": ["zeroth", "the previous", "the first", "null string"]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"preferredStrings": []string{"zeroth", "the previous", "the first", "null string"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToSmallerSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"preferredStrings": ["", "the first"]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"preferredStrings": []string{"", "the first"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToLargerSize(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"preferredStrings": ["", "the previous", "the first", "empty string", "blank string", "hitchi"]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"preferredStrings": []string{
								"",
								"the previous",
								"the first",
								"empty string",
								"blank string",
								"hitchi",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
