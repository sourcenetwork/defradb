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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerGivesUpdateWriteAccessToAnotherActorTwice_GQL_ShowThatTheRelationshipAlreadyExists(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives write(update) access to another actor twice, no-op",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			testUtils.GQLRequestMutationType,
		}),

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

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.UserIdentity(2), // This identity can not update yet.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				SkipLocalUpdateEvent: true,
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

func TestACP_OwnerGivesUpdateWriteAccessToAnotherActor_GQL_OtherActorCanUpdate(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives write(update) access to another actor",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			testUtils.GQLRequestMutationType,
		}),

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

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.UserIdentity(2), // This identity can not update yet.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				SkipLocalUpdateEvent: true,
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: false,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.UserIdentity(2), // This identity can now update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(2), // This identity can now also read.

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
							"name":   "Shahzad Lone", // Note: updated name
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
