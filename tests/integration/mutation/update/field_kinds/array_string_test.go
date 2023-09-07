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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithArrayOfStringsToNil(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with string array, replace with nil",
		// This restriction should be removed when we can, it is here because of
		// https://github.com/sourcenetwork/defradb/issues/1842
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"preferredStrings": null
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: []map[string]any{
					{
						"preferredStrings": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToNil_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with string array, replace with nil",
		// This is a bug, this test should be removed in
		// https://github.com/sourcenetwork/defradb/issues/1842
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"preferredStrings": null
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				ExpectedError: "EOF",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with string array, replace with empty",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"preferredStrings": []
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: []map[string]any{
					{
						"preferredStrings": []string{},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToSameSize(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with string array, replace with same size",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"preferredStrings": ["zeroth", "the previous", "the first", "null string"]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: []map[string]any{
					{
						"preferredStrings": []string{"zeroth", "the previous", "the first", "null string"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToSmallerSize(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with string array, replace with smaller size",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"preferredStrings": ["", "the first"]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: []map[string]any{
					{
						"preferredStrings": []string{"", "the first"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithArrayOfStringsToLargerSize(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with string array, replace with larger size",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						preferredStrings: [String!]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"preferredStrings": ["", "the previous", "the first", "empty string", "blank string", "hitchi"]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}
