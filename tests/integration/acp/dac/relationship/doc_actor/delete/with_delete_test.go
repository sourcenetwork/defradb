// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_relationship_doc_actor_delete

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerRevokesDeleteAccess_OtherActorCanNoLongerDelete(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner revokes delete access from another actor, they can not delete anymore",

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + updater + deleter

                          update:
                            expr: owner + updater

                          delete:
                            expr: owner + deleter

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          updater:
                            types:
                              - actor

                          deleter:
                            types:
                              - actor

                          admin:
                            manages:
                              - reader
                            types:
                              - actor

                          dummy:
                            types:
                              - actor
                `,
			},

			testUtils.SchemaUpdate{
				Schema: `
						type Users @policy(
							id: "{{.Policy0}}",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},

			// Creating two documents because need one to do the test on after one is deleted.
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad Lone",
						"age": 28
					}
				`,
			},

			// Give access to the other actor to delete and read both documents.
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "deleter",

				ExpectedExistence: false,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 1,

				Relation: "deleter",

				ExpectedExistence: false,
			},

			// Now the other identity can read both and delete both of those documents
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can read.

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(28),
						},
						{
							"name": "Shahzad Lone",
							"age":  int64(28),
						},
					},
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // This identity can also delete.

				DocID: 1,
			},

			testUtils.DeleteDACActorRelationship{ // Revoke access from being able to delete (and read) the document.
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "deleter",

				ExpectedRecordFound: true,
			},

			// The other identity can neither delete nor read the other document anymore.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{}, // Can't read the document anymore
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				ExpectedError: "document not found or not authorized to access", // Can't delete the document anymore.
			},

			// Ensure document was not accidentally deleted using owner identity.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(1),

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
