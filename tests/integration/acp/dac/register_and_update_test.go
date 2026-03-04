// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestACP_AddWithoutIdentityAndUpdateWithoutIdentity_CanUpdate(t *testing.T) {
	// The same identity that is used to do the registering/creation should be used in the
	// final read check to see the state of that registered document.
	// Note: In this test that identity is empty (no identity).

	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
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
							"_docID": "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f",
							"name":   "Shahzad Lone",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddWithoutIdentityAndUpdateWithIdentity_CanUpdate(t *testing.T) {
	// The same identity that is used to do the registering/creation should be used in the
	// final read check to see the state of that registered document.
	// Note: In this test that identity is empty (no identity).

	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(1),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
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
							"_docID": "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f",
							"name":   "Shahzad Lone",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddWithIdentityAndUpdateWithIdentity_CanUpdate(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(1),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),

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
							"_docID": "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f",
							"name":   "Shahzad Lone",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddWithIdentityAndUpdateWithoutIdentity_CanNotUpdate(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),

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
							"_docID": "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddWithIdentityAndUpdateWithWrongIdentity_CanNotUpdate(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),

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
							"_docID": "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This separate GQL test should be merged with the ones above when all the clients are fixed
// to behave the same in: https://github.com/sourcenetwork/defradb/issues/2410
func TestACP_AddWithIdentityAndUpdateWithoutIdentityGQL_CanNotUpdate(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			state.GQLRequestMutationType,
		}),

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				SkipLocalUpdateEvent: true,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),

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
							"_docID": "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This separate GQL test should be merged with the ones above when all the clients are fixed
// to behave the same in: https://github.com/sourcenetwork/defradb/issues/2410
func TestACP_AddWithIdentityAndUpdateWithWrongIdentityGQL_CanNotUpdate(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			state.GQLRequestMutationType,
		}),

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				SkipLocalUpdateEvent: true,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(1),

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
							"_docID": "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
