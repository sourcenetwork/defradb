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

package test_acp_dac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddWithoutIdentityAndReadWithoutIdentity_CanRead(t *testing.T) {
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

func TestACP_AddWithoutIdentityAndReadWithIdentity_CanRead(t *testing.T) {
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

func TestACP_AddWithIdentityAndReadWithIdentity_CanRead(t *testing.T) {
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

func TestACP_AddWithIdentityAndReadWithoutIdentity_CanNotRead(t *testing.T) {
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
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddWithIdentityAndReadWithWrongIdentity_CanNotRead(t *testing.T) {
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

			&action.Request{
				Identity: testUtils.ClientIdentity(2),

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
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
