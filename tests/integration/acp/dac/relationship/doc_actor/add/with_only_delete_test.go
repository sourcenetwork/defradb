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

package test_acp_dac_relationship_doc_actor_add

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerGivesDeleteAccessToAnotherActorWithoutExplicitReadPerm_OtherActorCanDelete(t *testing.T) {
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
  - expr: reader
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

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can not read yet.

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
					"Users": []map[string]any{}, // Can't see the documents yet
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // This identity can not delete yet.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "deleter",

				ExpectedExistence: false,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can now read.

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

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // This identity can now delete.

				DocID: 0,
			},

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // Check if actually deleted.

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
