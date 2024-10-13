// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_relationship_doc_actor_add

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerGivesDeleteWriteAccessToAnotherActorTwice_ShowThatTheRelationshipAlreadyExists(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives write(delete) access to another actor twice, no-op",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + writer

                          write:
                            expr: owner + writer

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          writer:
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

				ExpectedPolicyID: expectedPolicyID,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
						type Users @policy(
							id: "%s",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
					expectedPolicyID,
				),
			},

			testUtils.CreateDoc{
				Identity: testUtils.UserIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(2), // This identity can not read yet.

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

				Identity: testUtils.UserIdentity(2), // This identity can not delete yet.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: false,
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: true, // is a no-op
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_OwnerGivesDeleteWriteAccessToAnotherActor_OtherActorCanDelete(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives write(delete) access to another actor",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + writer

                          write:
                            expr: owner + writer

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          writer:
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

				ExpectedPolicyID: expectedPolicyID,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
						type Users @policy(
							id: "%s",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
					expectedPolicyID,
				),
			},

			testUtils.CreateDoc{
				Identity: testUtils.UserIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(2), // This identity can not read yet.

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

				Identity: testUtils.UserIdentity(2), // This identity can not delete yet.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(2), // This identity can now read.

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
							"_docID": "bae-9d443d0c-52f6-568b-8f74-e8ff0825697b",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.UserIdentity(2), // This identity can now delete.

				DocID: 0,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(2), // Check if actually deleted.

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

func TestACP_OwnerGivesDeleteWriteAccessToAnotherActor_OtherActorCanDeleteSoCanTheOwner(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives write(delete) access to another actor, both can read",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + writer

                          write:
                            expr: owner + writer

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          writer:
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

				ExpectedPolicyID: expectedPolicyID,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
						type Users @policy(
							id: "%s",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
					expectedPolicyID,
				),
			},

			testUtils.CreateDoc{
				Identity: testUtils.UserIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(1), // Owner can still also delete (ownership not transferred)

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
							"_docID": "bae-9d443d0c-52f6-568b-8f74-e8ff0825697b",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.UserIdentity(1), // Owner can still also delete.

				DocID: 0,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(1), // Check if actually deleted.

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
