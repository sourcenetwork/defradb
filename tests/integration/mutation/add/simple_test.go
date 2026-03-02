// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package add

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationAdd_GivenNonExistantField_Errors(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"fieldDoesNotExist": 27
				}`,
				ExpectedError: "the given field does not exist. Name: fieldDoesNotExist",
			},
			&action.Request{
				// Ensure that no documents have been written.
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-32e84498-d467-5f01-b93e-fc2dca59be76",
							"name":   "John",
							"age":    int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_GivenDuplicate_Errors(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// Collection.Save would treat the second create as an update, and so
			// is excluded from this test.
			state.CollectionNamedMutationType,
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
				ExpectedError: "a document with the given ID already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_GivenEmptyInput(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					add_Users(input: {}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
						{
							"_docID": "bae-d97a4927-9fad-53a0-bda2-8e9d8dd33551",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_With10Collections(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Foo1 {
						# The name used for the fields is important as the field shortID
						# is serially assigned based on the alphabetical order of field names.
						about: String
						name: String
					}
					type Foo2 {
						name: String
					}
					type Foo3 {
						name: String
					}
					type Foo4 {
						name: String
					}
					type Foo5 {
						name: String
					}
					type Foo6 {
						name: String
					}
					type Foo7 {
						name: String
					}
					type Foo8 {
						name: String
					}
					type Foo9 {
						name: String
					}
					type Foo10 {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					add_Foo1(input: {about: "something", name: "John"}) {
						about
						name
					}
				}`,
				Results: map[string]any{
					"add_Foo1": []map[string]any{
						{
							"about": "something",
							"name":  "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
