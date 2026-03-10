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

package test_acp_dac_relationship_doc_actor_delete

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerRevokesAccessFromAllNonExplicitActors_ActorsCanNotReadAnymore(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
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
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(), // Give implicit access to all identities.

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // Any identity can read

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

			&action.Request{
				Identity: testUtils.ClientIdentity(3), // Any identity can read

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

			testUtils.DeleteDACActorRelationship{ // Revoke access from all actors, (ones given access through * implicitly).
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: true,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // Can not read anymore

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
					"Users": []map[string]any{}, // Can't see the documents now
				},
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(3), // Can not read anymore

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
					"Users": []map[string]any{}, // Can't see the documents now
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_OwnerRevokesAccessFromAllNonExplicitActors_ExplicitActorsCanStillRead(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
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
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2), // Give access to this identity explictly before.

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(), // Give implicit access to all identities.

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(4), // Give access to this identity explictly after.

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // Any identity can read

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

			&action.Request{
				Identity: testUtils.ClientIdentity(3), // Any identity can read

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

			&action.Request{
				Identity: testUtils.ClientIdentity(4), // Any identity can read

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

			&action.Request{
				Identity: testUtils.ClientIdentity(5), // Any identity can read

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

			testUtils.DeleteDACActorRelationship{ // Revoke access from all actors, (ones given access through * implicitly).
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: true,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(3), // Can not read anymore, because it gained access implicitly.

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
					"Users": []map[string]any{}, // Can't see the documents now
				},
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(5), // Can not read anymore, because it gained access implicitly.

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
					"Users": []map[string]any{}, // Can't see the documents now
				},
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // Can still read because it was given access explictly.

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

			&action.Request{
				Identity: testUtils.ClientIdentity(4), // Can still read because it was given access explictly.

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

func TestACP_OwnerRevokesAccessFromAllNonExplicitActors_NonIdentityRequestsCanNotReadAnymore(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
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
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(), // Give implicit access to all identities.

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			&action.Request{
				Identity: testUtils.NoIdentity(), // Can read even without identity

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

			testUtils.DeleteDACActorRelationship{ // Revoke access from all actors, (ones given access through * implicitly).
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: true,
			},

			&action.Request{
				Identity: testUtils.NoIdentity(), // Can not read anymore

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
					"Users": []map[string]any{}, // Can't see the documents now
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
